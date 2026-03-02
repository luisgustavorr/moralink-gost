package Service

import (
	Grpcclient "MoraLinkGOst/modules/grpc"
	"context"
	"log"

	"github.com/kardianos/service"
)

type Program struct {
	exit   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *Program) run() {
	p.ctx, p.cancel = context.WithCancel(context.Background())
	Grpcclient.GRPCGuardian(p.ctx)
	<-p.exit
	log.Println("agent shutdown")
}
func (p *Program) Stop(s service.Service) error {
	close(p.exit)
	return nil
}
func (p *Program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go p.run()
	return nil
}
