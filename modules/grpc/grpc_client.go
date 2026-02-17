package Grpcclient

import (
	"context"
	"fmt"
	"log"

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
	case pb.MessageType_ACK:
		ackReturn := msg.Payload.GetAckReturn()
		if ackReturn.Status == 1 {
			fmt.Println("✅ 🔐 ACK APPROVED")
		} else {
			fmt.Println("❌ 🔐 ACK DISAPPROVED")
		}
		connectedUser := msg.Payload.GetAckReturn().ConnectedUser
		db_info, err := utils.ParseDBConfig(connectedUser.ConfigJson)
		if err != nil {
			fmt.Println(err)
		}
		db, _ := dbmanagers.DecideWhoActs(connectedUser.DbType, db_info)
		utils.Conn.DB = db
		fmt.Println("✅ 🔗 Tunnel connected - ALL WORKING")
	case pb.MessageType_HEARTBEAT:
		log.Println("heartbeat received")
	case pb.MessageType_QUERY:
		dbConn := utils.Conn.DB
		batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
		tableAskedFor := msg.GetTable()
		switch tableAskedFor {
		case 0:
			fmt.Println("Produtos...")
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			err := dbConn.Queries.Products(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.ProdutoRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				utils.LogMemUsage()
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
			fmt.Println("error", err)
			if err != nil {

				c.SendError(err.Error())
			}
		case 1:
			fmt.Println("Clientes...")
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			err := dbConn.Queries.Clientes(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.ClienteRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				fmt.Println("Passando", isLast, len(result))
				utils.LogMemUsage()
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
			fmt.Println("error", err)
			if err != nil {

				c.SendError(err.Error())
			}
		case 2:
			result, err := dbConn.Queries.Categorias(msg.Payload.GetQueryRequest().Query, dbConn.DB)
			if err == nil {
				if len(result) == 0 {
					c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				for i := 0; i < len(result); i += batchSize {
					fmt.Println("devolver resultado", len(result))
					end := i + batchSize
					isLast := end == len(result)
					fmt.Println("Passando")
					if end > len(result) {
						fmt.Println("Último ? ")
						end = len(result)
					}
					resultPb := utils.ToProtoCategorias(result[i:end])
					utils.LogMemUsage()

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
				c.SendError(err.Error())
			}
		case 3:
			fmt.Println("Vendas...")
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
			err := dbConn.Queries.Vendas(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []utils.VendaRow) error {
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				fmt.Println("Passando", isLast, len(result))
				utils.LogMemUsage()
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
			fmt.Println("error", err)
			if err != nil {

				c.SendError(err.Error())
			}
		case 4:
			result, err := dbConn.Queries.Vendedores(msg.Payload.GetQueryRequest().Query, dbConn.DB)
			if err == nil {
				if len(result) == 0 {
					c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				for i := 0; i < len(result); i += batchSize {
					fmt.Println("devolver resultado", len(result))
					end := i + batchSize
					isLast := end == len(result)
					fmt.Println("Passando")
					if end > len(result) {
						fmt.Println("Último ? ")
						end = len(result)
					}
					resultPb := utils.ToProtoVendedores(result[i:end])
					utils.LogMemUsage()

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
				c.SendError(err.Error())
			}

		case 6:
			batchSize := int(msg.Payload.GetQueryRequest().BatchSize)

			err := dbConn.Queries.Generic(msg.Payload.GetQueryRequest().Query, dbConn.DB, batchSize, func(result []map[string]interface{}) error {
				fmt.Println("devolver resultado", len(result))
				if len(result) == 0 {
					return c.SendMessage(buildEmptyMimicReturn(tableAskedFor, msg.GetBatchId()))
				}
				isLast := len(result) < batchSize
				fmt.Println("Passando", isLast, len(result))

				utils.LogMemUsage()

				resultPb, err := utils.ToProtoGenecric(result)
				if err != nil {
					fmt.Println("ERRO NO GENERIC :", err)
				}

				fmt.Println("Bid :", msg.BatchId)
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
				c.SendError(err.Error())
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
	fmt.Println("✅ 🌐 Grpc started")
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
		fmt.Println(err)
		return err
	}
	c.stream = stream
	// Send HELLO
	message := &pb.AgentMessage{
		AgentId: viper.GetString("api.token"),
		Message: "Olá, recebi sua mensagem",
		Type:    pb.MessageType_HELLO,
		Payload: &pb.AgentPayload{Data: &pb.AgentPayload_Produtos{
			Produtos: &pb.Produtos{Items: []*pb.Produto{
				{
					IdExterno: "teste_123",
					Valor:     12.30,
				},
			}},
		}},
	}
	// err = stream.Send(message)
	c.SendMessage(message)

	if err != nil {
		fmt.Println(err)
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
	err := c.stream.Send(msg)
	return err
}
func (c *Client) SendError(message string) error {
	err := c.SendMessage(&pb.AgentMessage{
		AgentId: viper.GetString("api.token"),
		Message: "Ocorreu um erro...",
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
