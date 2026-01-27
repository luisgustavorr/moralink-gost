package Service

import (
	Grpcclient "MoraLinkGOst/modules/grpc"
	"context"
	"log"

	"github.com/kardianos/service"
)

type Program struct {
	exit chan struct{}
}

func (p *Program) run() {
	log.Println("MoraLinkGOst started")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := Grpcclient.New(
		"agent-123",
		"0.1.0",
		"localhost:50051",
	)
	go func() {
		if err := client.Run(ctx); err != nil {
			log.Println("grpc error:", err)
		}
	}()

	<-p.exit
	cancel()
	log.Println("agent shutdown")
}
func (p *Program) Stop(s service.Service) error {
	close(p.exit)
	return nil
}
func (p *Program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go p.Start(s)
	return nil
}
