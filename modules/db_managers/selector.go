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
	fmt.Printf("✅ ❔ Using %s connection \n", pb.DbType_name[int32(dType)])

	var db = &utils.DbInfos{
		Type: dType,
	}

	switch dType {
	case 0:
		db, err = connectMysql(connInfo, db)
		if err != nil {
			fmt.Println("Error psql: ", err)
			panic(err)
		}
	case 1:
		db, err = connectPostgresql(connInfo, db)
		if err != nil {
			fmt.Println("Error psql: ", err)
			panic(err)
		}

	case 2:
	case 3:
	case 4:

	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ 💽 Connected to %s Database \n", pb.DbType_name[int32(dType)])
	return db, err
}
