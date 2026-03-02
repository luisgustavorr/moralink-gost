package main

import (
	Service "MoraLinkGOst/modules/service"
	"MoraLinkGOst/modules/utils"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/joho/godotenv"
	"github.com/kardianos/service"
)

var _ = godotenv.Load()

func main() {
	utils.LoadConfig()
	if os.Getenv("dev") == "1" {
		go func() {
			log.Println("pprof listening on :6060")
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	svcConfig := &service.Config{
		Name:        "moralink-gost",
		DisplayName: "MoraLink",
		Description: "Gerencia a integração com o SharkBusiness",
	}
	prg := &Service.Program{}
	svc, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		err = service.Control(svc, os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	err = svc.Run()
	if err != nil {
		log.Fatal(err)
	}
}
