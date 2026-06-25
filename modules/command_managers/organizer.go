package commandManagers

import (
	dbmanagers "MoraLinkGOst/modules/db_managers"
	pb "MoraLinkGOst/modules/proto/agentpb"
	"MoraLinkGOst/modules/updater"
	"MoraLinkGOst/modules/utils"
	"log"
	"path/filepath"

	"github.com/spf13/viper"
)

func ExecCommand(c *pb.Commands) {
	switch c.GetType() {
	case pb.Command_RESTART_APP:
		log.Println("RESTART APP COMMAND...")
		RestartSelf()
	case pb.Command_RESTART_DB:
		addInfo := c.GetAckReturn()
		if addInfo != nil {
			connectedUser := addInfo.ConnectedUser
			db_info, err := utils.ParseDBConfig(connectedUser.GetDbConfigJson())
			if err != nil {
				log.Println(err)
			}
			db, err := dbmanagers.DecideWhoActs(connectedUser.GetDbConn(), db_info)
			utils.Conn.DB = db
			if err != nil {
				log.Println("⚠️ 🔗 Tunnel connected - DB not working")
			} else {
				log.Println("✅ 🔗 Tunnel connected - ALL WORKING")
			}
		}

		// RestartDB()
	case pb.Command_UPDATE_APP:
		version := c.GetVersion()
		if version == "" {
			version = "latest" // you can resolve "latest" in GetRelease
		}
		go func() { // run in goroutine so gRPC stream isn't blocked
			err := updater.DownloadRelease(version)
			if err != nil {
				log.Println("❌ Update failed:", err.Error())
			}
		}()
	case pb.Command_CONFIGURE:
		cfgDir := utils.ConfigPath()
		configSet := c.GetConfigure()
		viper.Set("api.mode", "prod")
		viper.Set("api.token", configSet.Token)
		configFile := filepath.Join(cfgDir, "config.yaml")
		if err := viper.WriteConfigAs(configFile); err != nil {
			log.Fatalf("failed to write default config to %s: %v", configFile, err)
		}
		RestartSelf()
	}

}
