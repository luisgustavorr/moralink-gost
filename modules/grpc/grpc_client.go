package Grpcclient

import (
	"context"
	"fmt"
	"log"

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
		log.Printf("received ACK: %s - %s", msg.Message, ackReturn.Status, utils.JsonViewInterface(msg.Payload.GetAckReturn()))
		// c.SendMessage(&pb.AgentMessage{
		// 	AgentId: viper.GetString("api.token"),
		// 	Message: "Resultado QUery : 1q231",
		// 	Type:    pb.MessageType_RESULT,
		// 	Payload: &pb.AgentPayload{Data: &pb.AgentPayload_Produtos{
		// 		Produtos: &pb.Produtos{Items: []*pb.Produto{
		// 			{
		// 				IdExterno: "teste_123",
		// 				Valor:     12.30,
		// 			},
		// 		}},
		// 	}},
		// })
	case pb.MessageType_HEARTBEAT:
		log.Println("heartbeat received")

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
	fmt.Println("send:", utils.JsonViewInterface(msg))
	err := c.stream.Send(msg)
	return err
}
