package commandManagers

import (
	"MoraLinkGOst/modules/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/kardianos/service"
	"github.com/spf13/viper"
)

func RestartSelf() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	args := os.Args
	env := os.Environ()
	if runtime.GOOS == "windows" {
		cmd := exec.Command(self, args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env
		return cmd.Run()
	}
	return syscall.Exec(self, args, env)
}
func DeactivateSelf() {
	cfgDir := utils.ConfigPath()
	viper.Set("api.active", false)
	configFile := filepath.Join(cfgDir, "config.yaml")
	if err := viper.WriteConfigAs(configFile); err != nil {
		log.Fatalf("failed to write default config to %s: %v", configFile, err)
	}
	RestartSelf()
}
func ReactivateSelf() {
	cfgDir := utils.ConfigPath()
	viper.Set("api.active", true)
	configFile := filepath.Join(cfgDir, "config.yaml")
	if err := viper.WriteConfigAs(configFile); err != nil {
		log.Fatalf("failed to write default config to %s: %v", configFile, err)
	}
	RestartSelf()
}

var ServiceRunning service.Service

// func Uninstall() {

// 	if ServiceRunning != nil {
// 		err := service.Control(ServiceRunning, "uninstall")
// 		if err != nil {
// 			log.Fatalf("Failed to %s service: %v\n(try running as Administrator on Windows)", "uninstall", err)
// 		}
// 	}

// }
