package Grpcclient

import (
	"context"
	"log"
	"os"
	"time"

	commandManagers "MoraLinkGOst/modules/command_managers"
	dbmanagers "MoraLinkGOst/modules/db_managers"
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
			log.Println("✅ 🔐 ACK APPROVED")
		} else {
			log.Println("❌ 🔐 ACK DISAPPROVED")
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
		dbConn := utils.Conn.DB
		batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
		tableAskedFor := msg.GetTable()
		switch tableAskedFor {
		case 0:
			log.Println("Produtos...")
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			err := dbConn.Queries.Products(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.ProdutoRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				resultPb := utils.ToProtoProdutos(result)
				return c.SendMessage(&agentpb.AgentMessage{
					AgentId: viper.GetString("api.token"),
					Type:    agentpb.MessageType_RESULT,
					Table:   tableAskedFor,
					IsLast:  isLast,
					Payload: &pb.AgentPayload{
						Data: &pb.AgentPayload_Produtos{
							Produtos: &pb.Produtos{
								Items: resultPb,
							},
						},
					},
				})
			})
			log.Println("error", err)
			if err != nil {

				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 1:
			log.Println("Clientes...")
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			err := dbConn.Queries.Clientes(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.ClienteRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				resultPb := utils.ToProtoClientes(result)
				return c.SendMessage(&agentpb.AgentMessage{
					AgentId: viper.GetString("api.token"),
					Type:    agentpb.MessageType_RESULT,
					Table:   tableAskedFor,
					IsLast:  isLast,
					Payload: &pb.AgentPayload{
						Data: &pb.AgentPayload_Clientes{
							Clientes: &pb.Clientes{
								Items: resultPb,
							},
						},
					},
				})
			})
			log.Println("error", err)
			if err != nil {

				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 2:
			result, err := dbConn.Queries.Categorias(msg.Payload.GetQueryRequest().Query, dbConn.DB)
			if err == nil {
				if len(result) == 0 {
					c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				for i := 0; i < len(result); i += batchSize {
					log.Println("devolver resultado", len(result))
					end := i + batchSize
					isLast := len(result) < batchSize
					log.Println("Passando")
					if end > len(result) {
						log.Println("Último ? ")
						end = len(result)
					}
					resultPb := utils.ToProtoCategorias(result[i:end])

					c.SendMessage(&agentpb.AgentMessage{
						AgentId: viper.GetString("api.token"),
						Type:    agentpb.MessageType_RESULT,
						Table:   tableAskedFor,
						IsLast:  isLast,
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

				log.Println("Query para ", agentpb.Table_name[int32(msg.GetTable())], " retornando ", len(result))
			} else {
				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 3:
			log.Println("Vendas...")
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			err := dbConn.Queries.Vendas(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.VendaRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				log.Println("Passando", isLast, len(result))
				resultPb := utils.ToProtoVendas(result)
				return c.SendMessage(&agentpb.AgentMessage{
					AgentId: viper.GetString("api.token"),
					Type:    agentpb.MessageType_RESULT,
					Table:   tableAskedFor,
					IsLast:  isLast,
					Payload: &pb.AgentPayload{
						Data: &pb.AgentPayload_Vendas{
							Vendas: &pb.Vendas{
								Items: resultPb,
							},
						},
					},
				})
			})
			log.Println("error", err)
			if err != nil {

				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 4:
			result, err := dbConn.Queries.Vendedores(msg.Payload.GetQueryRequest().Query, dbConn.DB)
			if err == nil {
				if len(result) == 0 {
					c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				for i := 0; i < len(result); i += batchSize {
					log.Println("devolver resultado", len(result))
					end := i + batchSize
					isLast := len(result) < batchSize
					log.Println("Passando")
					if end > len(result) {
						log.Println("Último ? ")
						end = len(result)
					}
					resultPb := utils.ToProtoVendedores(result[i:end])

					c.SendMessage(&agentpb.AgentMessage{
						AgentId: viper.GetString("api.token"),
						Type:    agentpb.MessageType_RESULT,
						Table:   tableAskedFor,
						IsLast:  isLast,
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

				log.Println("Query para ", agentpb.Table_name[int32(msg.GetTable())], " retornando ", len(result))
			} else {
				c.SendError(err.Error(), msg.GetBatchId())
			}
		case 5:
			log.Println("Financeiro...")
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			err := dbConn.Queries.Financeiros(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.FinanceiroRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				resultPb := utils.ToProtoFinanceiro(result)
				return c.SendMessage(&agentpb.AgentMessage{
					AgentId: viper.GetString("api.token"),
					Type:    agentpb.MessageType_RESULT,
					Table:   tableAskedFor,
					IsLast:  isLast,
					Payload: &pb.AgentPayload{
						Data: &pb.AgentPayload_Financeiros{
							Financeiros: &pb.Financeiros{
								Items: resultPb,
							},
						},
					},
				})
			})
			log.Println("error", err)
			if err != nil {

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
				log.Println("devolver resultado", len(result))
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				log.Println("Passando", isLast, len(result))

				resultPb, err := utils.ToProtoGenecric(result)
				if err != nil {
					log.Println("ERRO NO PROTO GENERIC :", err)
				}

				log.Println("Bid :", msg.BatchId)
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
	log.Println("✅ 🌐 Grpc started")
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

func (c *Client) SendMessage(msg *pb.AgentMessage) error {
	msg.Version = c.version
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
	log.Println("✅ 🛡️  Guardian started, search connection at : ", host, utils.Version)

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
