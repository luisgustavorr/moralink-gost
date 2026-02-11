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
			fmt.Println("Aprovado")
		} else {
			fmt.Println("Desaprovado")
		}
		log.Printf("received ACK: %s - %s", msg.Message, ackReturn.Status)
		connectedUser := msg.Payload.GetAckReturn().ConnectedUser
		db_info, err := utils.ParseDBConfig(connectedUser.ConfigJson)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("DB info : ", db_info, connectedUser.DbType.Number())
		db, _ := dbmanagers.DecideWhoActs(connectedUser.DbType, db_info)
		utils.Conn.DB = db
	case pb.MessageType_HEARTBEAT:
		log.Println("heartbeat received")
	case pb.MessageType_QUERY:
		dbConn := utils.Conn.DB
		batchSize := int(msg.Payload.GetQueryRequest().BatchSize)
		tableAskedFor := msg.GetTable()
		switch tableAskedFor {
		case 1:
			fmt.Println("CLientes...")
			result, err := dbConn.Queries.Clientes(msg.Payload.GetQueryRequest().Query, dbConn.DB)
			fmt.Println("error", err)
			if err == nil {
				resultPb := utils.ToProtoClientes(result)
				isLast := false
				for i := 0; i < len(resultPb); i += batchSize {
					end := i + batchSize

					fmt.Println("Passando")
					if end > len(resultPb) {
						fmt.Println("Último ? ")
						isLast = true
						end = len(resultPb)
					}

					c.SendMessage(&agentpb.AgentMessage{
						AgentId: viper.GetString("api.token"),
						Type:    agentpb.MessageType_RESULT,
						Table:   tableAskedFor,
						IsLast:  isLast,
						Payload: &pb.AgentPayload{
							Data: &pb.AgentPayload_Clientes{
								Clientes: &pb.Clientes{
									Items: resultPb[i:end],
								},
							},
						},
					})
				}

				log.Println("Query para ", agentpb.Table_name[int32(msg.GetTable())], " retornando ", len(result))
			} else {
				c.SendError(err.Error())
			}
		case 2:
			result, err := dbConn.Queries.Categorias(msg.Payload.GetQueryRequest().Query, dbConn.DB)
			if err == nil {
				fmt.Println("devolver resultado", len(result))
				resultPb := utils.ToProtoCategorias(result)

				isLast := false
				for i := 0; i < len(resultPb); i += batchSize {
					end := i + batchSize

					fmt.Println("Passando")
					if end > len(resultPb) {
						fmt.Println("Último ? ")
						isLast = true
						end = len(resultPb)
					}
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
				}

				log.Println("Query para ", agentpb.Table_name[int32(msg.GetTable())], " retornando ", len(result))
			} else {
				c.SendError(err.Error())
			}
		case 6:
			result, err := dbConn.Queries.Generic(msg.Payload.GetQueryRequest().Query, dbConn.DB)
			if err == nil {
				fmt.Println("devolver resultado", len(result))
				resultPb, err := utils.ToProtoGenecric(result)
				if err != nil {
					fmt.Println("ERRO NO GENERIC :", err)
				}
				isLast := false
				i := 0
				if len(result) == 0 {
					c.SendMessage(&agentpb.AgentMessage{
						AgentId: viper.GetString("api.token"),
						Type:    agentpb.MessageType_RESULT,
						Table:   tableAskedFor,
						BatchId: msg.GetBatchId(),
						IsLast:  isLast,
						IsEmpty: true,
						Payload: &pb.AgentPayload{},
					})
				}
				for i = 0; i < len(result); i += batchSize {
					end := i + batchSize
					fmt.Println("Passando")
					if end > len(result) {
						fmt.Println("Último ? ")
						isLast = true
						end = len(result)
					}
					fmt.Println("Bid :", msg.BatchId)
					c.SendMessage(&agentpb.AgentMessage{
						AgentId: viper.GetString("api.token"),
						Type:    agentpb.MessageType_RESULT,
						Table:   tableAskedFor,
						BatchId: msg.GetBatchId(),
						IsLast:  isLast,
						Payload: &pb.AgentPayload{
							Data: &pb.AgentPayload_GenericReturn{
								GenericReturn: resultPb,
							},
						},
					})
				}

				log.Println("Query para ", agentpb.Table_name[int32(msg.GetTable())], " retornando ", len(result))
			} else {
				c.SendError(err.Error())
			}
		}

	default:
		log.Println("unknown message", msg.Type)
	}
}
func (c *Client) Run(ctx context.Context) error {
	fmt.Println("Iniciando gRPC")
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

	log.Println("connected to server")

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
