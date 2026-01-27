package main

import (
	Grpcclient "MoraLinkGOst/modules/grpc"
	"MoraLinkGOst/modules/utils"
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var _ = godotenv.Load()

func main() {
	utils.LoadConfig()
	// service logic
	// svcConfig := &service.Config{
	// 	Name:        "moralink-gost",
	// 	DisplayName: "Aplicativo de integração",
	// 	Description: "Gerencia a integração com o SharkBusiness",
	// }
	// prg := &Service.Program{}
	// svc, err := service.New(prg, svcConfig)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // Handle install / uninstall / start / stop
	// if len(os.Args) > 1 {
	// 	err = service.Control(svc, os.Args[1])
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	return
	// }
	// err = svc.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// Moralink logic
	log.Println("MoraLinkGOst started")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := Grpcclient.New(
		viper.GetString("api.user"),
		"0.1.0",
		"localhost:50051",
	)

	if err := client.Run(ctx); err != nil {
		log.Println("grpc error:", err)
	}

}
