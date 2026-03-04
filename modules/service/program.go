package Service

import (
	Grpcclient "MoraLinkGOst/modules/grpc"
	"MoraLinkGOst/modules/logger"
	"MoraLinkGOst/modules/utils"
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
)

type Program struct {
	exit   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *Program) run() {
	exe, err := os.Executable()
	if err == nil {
		logger.Init(filepath.Dir(exe))
	}

	// Now load config (which may log internally)
	utils.LoadConfig()

	logger.Init(utils.ConfigPath())
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
