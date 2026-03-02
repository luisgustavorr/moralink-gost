//go:build !windows

package updater

import (
	"os/exec"
	"syscall"
)

func setSysProcAttrDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // detach from parent session
	}
}
