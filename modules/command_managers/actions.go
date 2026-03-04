package commandManagers

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/kardianos/service"
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

var ServiceRunning service.Service

// func Uninstall() {

// 	if ServiceRunning != nil {
// 		err := service.Control(ServiceRunning, "uninstall")
// 		if err != nil {
// 			log.Fatalf("Failed to %s service: %v\n(try running as Administrator on Windows)", "uninstall", err)
// 		}
// 	}

// }
