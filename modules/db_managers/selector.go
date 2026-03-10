package dbmanagers

import (
	"MoraLinkGOst/modules/logger"
	pb "MoraLinkGOst/modules/proto/agentpb"
	"MoraLinkGOst/modules/utils"
	"log"
)

var OnStartupError string = ""

// DecideWhoActs is the organizer, the one that get the db type and return which connection will be used to handle the connection
func DecideWhoActs(dType pb.DbType, connInfo map[string]interface{}) (*utils.DbInfos, error) {
	var err error
	logger.Debugf("✅ ❔ Using %s connection \n", pb.DbType_name[int32(dType)])

	var db = &utils.DbInfos{
		Type: dType,
	}

	switch dType {
	case 0:
		db, err = connectMysql(connInfo, db)
		if err != nil {
			log.Println("Error mysql: ", err)

		}
	case 1:
		db, err = connectPostgresql(connInfo, db)
		if err != nil {
			log.Println("Error psql: ", err)

		}

	case 2:
		db, err = connectFirebird(connInfo, db)
		if err != nil {
			log.Println("Error firebird: ", err)

		}
	case 3:
		db, err = connectMssql(connInfo, db)
		if err != nil {
			log.Println("Error mssql: ", err)

		}
	case 4:
		db, err = connectMysql(connInfo, db)
		if err != nil {
			log.Println("Error mysql old : ", err)

		}

	}
	if err == nil {
		logger.Debugf("✅ 💽 Connected to %s Database \n", pb.DbType_name[int32(dType)])

	} else {
		OnStartupError = err.Error()
		log.Printf("❌  💽 Connection to %s Database failed : '%s' \n", pb.DbType_name[int32(dType)], err.Error())
	}
	return db, err
}
