package commandManagers

import (
	dbmanagers "MoraLinkGOst/modules/db_managers"
	pb "MoraLinkGOst/modules/proto/agentpb"
	"MoraLinkGOst/modules/utils"
	"fmt"
)

func ExecCommand(c *pb.Commands) {
	switch c.GetType() {
	case pb.Command_RESTART_APP:
		fmt.Println("RESTART APP COMMAND...")
		RestartSelf()
	case pb.Command_RESTART_DB:
		addInfo := c.GetAckReturn()
		if addInfo != nil {
			connectedUser := addInfo.ConnectedUser
			db_info, err := utils.ParseDBConfig(connectedUser.ConfigJson)
			if err != nil {
				fmt.Println(err)
			}
			db, err := dbmanagers.DecideWhoActs(connectedUser.DbType, db_info)
			utils.Conn.DB = db
			if err != nil {
				fmt.Println("⚠️ 🔗 Tunnel connected - DB not working")
			} else {
				fmt.Println("✅ 🔗 Tunnel connected - ALL WORKING")
			}
		}

		// RestartDB()
	}

}
