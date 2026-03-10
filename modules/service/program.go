package Service

import (
	Grpcclient "MoraLinkGOst/modules/grpc"
	"MoraLinkGOst/modules/logger"
	"MoraLinkGOst/modules/utils"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/service"
)

type Program struct {
	exit   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func emergencyLog(msg string) {
	// Writes to C:\moralink-boot.log — no dependencies, no config needed.
	// Delete this after confirming logs work.
	f, err := os.OpenFile(`C:\Users\Public\Documents\moralink-boot.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("ERR", err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s\n", time.Now().Format(time.RFC3339), msg)
}
func (p *Program) run() {
	emergencyLog("run() called")

	exe, err := os.Executable()
	emergencyLog(fmt.Sprintf("exe=%s err=%v", exe, err))

	if err == nil {
		logger.Init(filepath.Dir(exe))
	}
	emergencyLog("logger.Init(exe dir) done")

	utils.LoadConfig()
	emergencyLog(fmt.Sprintf("LoadConfig done, ConfigPath=%s", utils.ConfigPath()))

	logger.Init(utils.ConfigPath())
	emergencyLog("logger.Init(ConfigPath) done")

	p.ctx, p.cancel = context.WithCancel(context.Background())
	emergencyLog("starting GRPCGuardian")

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
