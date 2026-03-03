package main

import (
	"MoraLinkGOst/modules/logger"
	Service "MoraLinkGOst/modules/service"
	"MoraLinkGOst/modules/updater"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/joho/godotenv"
	"github.com/kardianos/service"
)

var (
	Version   = "dev"
	ReleaseGH = "" // injected at build time from .env via ldflags
)

func main() {
	updater.Configure(ReleaseGH)
	exePath, err := os.Executable()
	if err == nil {
		_ = godotenv.Load(exePath[:len(exePath)-len("moralink-gost.exe")] + ".env")
	}
	_ = godotenv.Load()

	if os.Getenv("dev") == "1" {
		go func() {
			log.Println("pprof listening on :6060")
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	svcConfig := &service.Config{
		Name:        "moralink-gost",
		DisplayName: "MoraLink GOst",
		Description: "Gerencia a integração com o Shark Business",
		Option: service.KeyValue{
			"DelayedAutoStart":       true,
			"OnFailure":              "restart",
			"OnFailureDelayDuration": "5s",
			"OnFailureResetPeriod":   10,
		},
	}

	prg := &Service.Program{}
	svc, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger.InitDefault()

	if len(os.Args) > 1 {
		action := os.Args[1]

		if action == "help" {
			fmt.Println("Usage: moralink-gost [command]")
			fmt.Println("Commands:")
			fmt.Println("  install    - Register as a Windows/Linux service")
			fmt.Println("  uninstall  - Remove the service")
			fmt.Println("  start      - Start the service")
			fmt.Println("  stop       - Stop the service")
			fmt.Println("  restart    - Restart the service")
			fmt.Println("  status     - Print service status")
			fmt.Println("  (no arg)   - Run interactively")
			return
		}

		err = service.Control(svc, action)
		if err != nil {
			log.Fatalf("Failed to %s service: %v\n(try running as Administrator on Windows)", action, err)
		}
		fmt.Printf("Service '%s' action '%s' completed successfully.\n", svcConfig.Name, action)
		return
	}

	if err = svc.Run(); err != nil {
		logger.Error(err)
	}
}
