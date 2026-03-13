package Grpcclient

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	commandManagers "MoraLinkGOst/modules/command_managers"
	dbmanagers "MoraLinkGOst/modules/db_managers"
	"MoraLinkGOst/modules/logger"
	"MoraLinkGOst/modules/proto/agentpb"
	pb "MoraLinkGOst/modules/proto/agentpb"
	"MoraLinkGOst/modules/utils"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	agentID string
	version string
	addr    string
	stream  grpc.BidiStreamingClient[pb.AgentMessage, pb.AgentMessage]
}

func New(agentID, version, addr string) *Client {
	return &Client{
		agentID: agentID,
		version: version,
		addr:    addr,
	}
}
func (c *Client) handleMessage(msg *pb.AgentMessage, s grpc.BidiStreamingClient[pb.AgentMessage, pb.AgentMessage]) {
	switch msg.Type {
	case pb.MessageType_ERROR:
		log.Printf("received command: %s", msg.Message)
	case pb.MessageType_COMMAND:
		log.Println("Comando recebido ... ", pb.Command_name[int32(msg.GetPayload().GetCommand().GetType())])
		commandManagers.ExecCommand(msg.GetPayload().GetCommand())

	case pb.MessageType_ACK:
		ackReturn := msg.Payload.GetAckReturn()
		if ackReturn.Status == 1 {
			logger.Debug("✅ 🔐 ACK APPROVED")
		} else {
			logger.Debug("❌ 🔐 ACK DISAPPROVED")
		}
		connectedUser := msg.Payload.GetAckReturn().ConnectedUser
		db_info, err := utils.ParseDBConfig(connectedUser.ConfigJson)
		if err != nil {
			log.Println(err)
		}
		db, err := dbmanagers.DecideWhoActs(connectedUser.DbType, db_info)
		utils.Conn.DB = db
		if err != nil {
			log.Println("⚠️ 🔗 Tunnel connected - DB not working")

		} else {
			log.Println("✅ 🔗 Tunnel connected - ALL WORKING")

		}
	case pb.MessageType_HEARTBEAT:
		log.Println("heartbeat received")
	case pb.MessageType_QUERY:
		runned := 0
		c.SendTrace(msg.GetBatchId(), pb.TraceStep_GOST_QUERY_RECEIVED)
		dbConn := utils.Conn.DB
		batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
		tableAskedFor := msg.GetTable()
		switch tableAskedFor {
		case 0:
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			if strings.TrimSpace(msg.Payload.GetQueryRequest().GetQuery()) == "" {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				return
			}
			err := dbConn.Queries.Products(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.ProdutoRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				resultPb := utils.ToProtoProdutos(result)
				if runned == 0 {
					c.SendTrace(msg.GetBatchId(), pb.TraceStep_GOST_DB_EXECUTED)
				}
				runned++
				return c.SendMessage(&agentpb.AgentMessage{
					AgentId: viper.GetString("api.token"),
					Type:    agentpb.MessageType_RESULT,
					BatchId: msg.GetBatchId(),

					Table:  tableAskedFor,
					IsLast: isLast,

					Payload: &pb.AgentPayload{
						Data: &pb.AgentPayload_Produtos{
							Produtos: &pb.Produtos{
								Items: resultPb,
							},
						},
					},
				})
			})
			if err != nil {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 1:
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			if strings.TrimSpace(msg.Payload.GetQueryRequest().GetQuery()) == "" {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				return
			}
			err := dbConn.Queries.Clientes(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.ClienteRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				resultPb := utils.ToProtoClientes(result)
				if runned == 0 {
					c.SendTrace(msg.GetBatchId(), pb.TraceStep_GOST_DB_EXECUTED)
				}
				runned++
				return c.SendMessage(&agentpb.AgentMessage{
					AgentId: viper.GetString("api.token"),
					Type:    agentpb.MessageType_RESULT,
					Table:   tableAskedFor,
					IsLast:  isLast,
					BatchId: msg.GetBatchId(),

					Payload: &pb.AgentPayload{
						Data: &pb.AgentPayload_Clientes{
							Clientes: &pb.Clientes{
								Items: resultPb,
							},
						},
					},
				})
			})
			if err != nil {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 2:
			result, err := dbConn.Queries.Categorias(msg.Payload.GetQueryRequest().Query, dbConn.DB)
			if strings.TrimSpace(msg.Payload.GetQueryRequest().GetQuery()) == "" {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				return
			}
			if err == nil {
				if len(result) == 0 {
					c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				for i := 0; i < len(result); i += batchSize {
					logger.Debug("devolver resultado", len(result))
					end := i + batchSize
					isLast := len(result) < batchSize
					if end > len(result) {
						end = len(result)
					}
					resultPb := utils.ToProtoCategorias(result[i:end])
					if runned == 0 {
						c.SendTrace(msg.GetBatchId(), pb.TraceStep_GOST_DB_EXECUTED)
					}
					runned++
					c.SendMessage(&agentpb.AgentMessage{
						AgentId: viper.GetString("api.token"),
						Type:    agentpb.MessageType_RESULT,
						Table:   tableAskedFor,
						BatchId: msg.GetBatchId(),

						IsLast: isLast,
						Payload: &pb.AgentPayload{
							Data: &pb.AgentPayload_Categorias{
								Categorias: &pb.Categorias{
									Items: resultPb,
								},
							},
						},
					})
					resultPb = nil

				}

			} else {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 3:
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			if strings.TrimSpace(msg.Payload.GetQueryRequest().GetQuery()) == "" {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				return
			}
			err := dbConn.Queries.Vendas(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.VendaRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				resultPb := utils.ToProtoVendas(result)
				if runned == 0 {
					c.SendTrace(msg.GetBatchId(), pb.TraceStep_GOST_DB_EXECUTED)
				}
				runned++
				return c.SendMessage(&agentpb.AgentMessage{
					AgentId: viper.GetString("api.token"),
					Type:    agentpb.MessageType_RESULT,
					Table:   tableAskedFor,
					IsLast:  isLast,
					BatchId: msg.GetBatchId(),

					Payload: &pb.AgentPayload{
						Data: &pb.AgentPayload_Vendas{
							Vendas: &pb.Vendas{
								Items: resultPb,
							},
						},
					},
				})
			})
			if err != nil {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 4:
			result, err := dbConn.Queries.Vendedores(msg.Payload.GetQueryRequest().Query, dbConn.DB)
			if strings.TrimSpace(msg.Payload.GetQueryRequest().GetQuery()) == "" {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				return
			}
			if err == nil {
				if len(result) == 0 {
					c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				for i := 0; i < len(result); i += batchSize {
					logger.Debug("devolver resultado", len(result))
					end := i + batchSize
					isLast := len(result) < batchSize
					if end > len(result) {
						log.Println("Último ? ")
						end = len(result)
					}
					resultPb := utils.ToProtoVendedores(result[i:end])
					if runned == 0 {
						c.SendTrace(msg.GetBatchId(), pb.TraceStep_GOST_DB_EXECUTED)
					}
					runned++
					c.SendMessage(&agentpb.AgentMessage{
						AgentId: viper.GetString("api.token"),
						Type:    agentpb.MessageType_RESULT,
						Table:   tableAskedFor,
						BatchId: msg.GetBatchId(),

						IsLast: isLast,
						Payload: &pb.AgentPayload{
							Data: &pb.AgentPayload_Vendedores{
								Vendedores: &pb.Vendedores{
									Items: resultPb,
								},
							},
						},
					})
					resultPb = nil

				}

			} else {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 5:
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			if strings.TrimSpace(msg.Payload.GetQueryRequest().GetQuery()) == "" {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				return
			}
			err := dbConn.Queries.Financeiros(msg.Payload.GetQueryRequest().GetQuery(), dbConn.DB, batchSize, func(result []utils.FinanceiroRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				resultPb := utils.ToProtoFinanceiro(result)
				if runned == 0 {
					c.SendTrace(msg.GetBatchId(), pb.TraceStep_GOST_DB_EXECUTED)
				}
				runned++
				return c.SendMessage(&agentpb.AgentMessage{
					AgentId: viper.GetString("api.token"),
					Type:    agentpb.MessageType_RESULT,
					Table:   tableAskedFor,
					BatchId: msg.GetBatchId(),

					IsLast: isLast,
					Payload: &pb.AgentPayload{
						Data: &pb.AgentPayload_Financeiros{
							Financeiros: &pb.Financeiros{
								Items: resultPb,
							},
						},
					},
				})
			})
			if err != nil {
				c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 6:
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			if dbConn == nil {
				dbConn = &utils.DbInfos{
					DB: nil,
				}
			}
			err := dbConn.Queries.Generic(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []map[string]interface{}) error {
				logger.Debug("devolver resultado", len(result))
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize

				resultPb, err := utils.ToProtoGenecric(result)
				if err != nil {
					log.Println("ERRO NO PROTO GENERIC :", err)
				}

				logger.Debug("Bid :", msg.BatchId)
				if runned == 0 {
					c.SendTrace(msg.GetBatchId(), pb.TraceStep_GOST_DB_EXECUTED)
				}
				runned++
				return c.SendMessage(&agentpb.AgentMessage{
					AgentId: viper.GetString("api.token"),
					Type:    agentpb.MessageType_RESULT,
					Table:   tableAskedFor,
					BatchId: msg.GetBatchId(),

					IsLast: isLast,
					Payload: &pb.AgentPayload{

						Data: &pb.AgentPayload_GenericReturn{
							GenericReturn: resultPb,
						},
					},
				})

			})
			if err != nil {
				c.SendError(err.Error(), msg.GetBatchId())
			}
		}

	default:
		log.Println("unknown message", msg.Type)
	}
}
func buildEmptyMimicReturn(table pb.Table, batchId string) *agentpb.AgentMessage {
	return &agentpb.AgentMessage{
		AgentId: viper.GetString("api.token"),
		Type:    agentpb.MessageType_RESULT,
		Table:   table,
		BatchId: batchId,
		IsLast:  true,
		IsEmpty: true,
		Payload: &pb.AgentPayload{},
	}
}
func (c *Client) Run(ctx context.Context) error {
	logger.Debug("✅ 🌐 Grpc started")
	conn, err := grpc.DialContext(
		ctx,
		c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewAgentServiceClient(conn)

	stream, err := client.Connect(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	c.stream = stream
	// Send HELLO
	message := &pb.AgentMessage{
		AgentId: viper.GetString("api.token"),
		Message: "Requesting connection",
		Type:    pb.MessageType_HELLO,
		Version: c.version,
	}

	c.SendMessage(message)

	if err != nil {
		log.Println(err)
		return err
	}

	// Receive loop
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}

		c.handleMessage(in, stream)
	}
}

func (c *Client) SendTrace(batch_id string, step pb.TraceStep) error {
	err := c.stream.Send(&pb.AgentMessage{
		AgentId: viper.GetString("api.token"),
		Type:    pb.MessageType_TRACE,
		Payload: &pb.AgentPayload{Data: &pb.AgentPayload_QueryTrace{
			QueryTrace: &pb.TraceEvent{
				BatchId:    batch_id,
				Step:       step,
				Timestamp:  time.Now().UnixMilli(),
				DurationMs: 0,
				Success:    true,
			},
		},
		},
	})
	return err
}
func (c *Client) SendMessage(msg *pb.AgentMessage) error {
	msg.Version = c.version
	fmt.Println("sHOULD ? ", msg.GetBatchId() != "" && msg.IsLast)
	if msg.GetBatchId() != "" && msg.IsLast {
		c.SendTrace(msg.GetBatchId(), pb.TraceStep_GOST_RESULT_SENT)
	}
	err := c.stream.Send(msg)
	return err
}
func (c *Client) SendError(message string, batchid string) error {
	err := c.SendMessage(&pb.AgentMessage{
		AgentId: viper.GetString("api.token"),
		Message: "Ocorreu um erro...",
		BatchId: batchid,
		Version: c.version,
		Type:    pb.MessageType_ERROR,
		Payload: &pb.AgentPayload{Data: &pb.AgentPayload_Erros{
			Erros: &pb.Erros{
				Error: []*pb.Error{
					{
						Message: message,
					},
				},
			},
		}},
	})
	return err
}

func GRPCGuardian(ctx context.Context) {
	host := "134.209.215.199:50051"
	if os.Getenv("dev") == "1" {
		host = "localhost:50051"
	}
	logger.Debug("✅ 🛡️  Guardian started, search connection at : ", host, utils.Version)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Guardian stopping")
			return
		default:
		}

		innerCtx, cancel := context.WithCancel(ctx)

		client := New(
			viper.GetString("api.token"),
			utils.Version,
			host,
		)

		err := client.Run(innerCtx)
		cancel()
		if err != nil {
			log.Println("⛔ -> grpc disconnected error:", err, viper.GetString("api.token"), viper.GetString("api.user"))
		}
		time.Sleep(1 * time.Minute)
	}
}
