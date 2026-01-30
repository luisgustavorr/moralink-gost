package dbmanagers

import (
	pb "MoraLinkGOst/modules/proto/agentpb"
	"MoraLinkGOst/modules/utils"
	"fmt"
	"log"
)

// DecideWhoActs is the organizer, the one that get the db type and return which connection will be used to handle the connection
func DecideWhoActs(dType pb.DbType, connInfo map[string]interface{}) (*utils.DbInfos, error) {
	var err error

	var db = &utils.DbInfos{
		Type: dType,
	}

	switch dType {
	case 0:
		fmt.Println("Requerindo conexão MySql")
	case 1:
		fmt.Println("Requerindo conexão Postgresql")
		db, err = connectPostgresql(connInfo, db)
		fmt.Println("Resultado :", err)

	case 2:
		fmt.Println("Requerindo conexão Firebird")
	case 3:
		fmt.Println("Requerindo conexão Mssql")
	case 4:
		fmt.Println("Requerindo conexão MySql Antigo")

	}
	if err != nil {
		log.Fatal(err)
	}
	return db, err
}
