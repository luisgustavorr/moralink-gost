//go:build linux

package updater

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func launchApply(tmp, target string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	unit := fmt.Sprintf("moralink-gost-updater-%d", os.Getpid())

	cmd := exec.Command(
		"/usr/bin/systemd-run",
		"--unit="+unit,
		"--collect",
		"--property=Type=oneshot",
		"--no-block",
		self,
		applyArg,
		tmp,
		target,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("schedule systemd updater: %w: %s", err, out)
	}

	return nil
}

func applyUpdate(tmp, target string) error {
	if err := validatePaths(tmp, target); err != nil {
		return err
	}

	// The service must run with sufficient privileges to manage systemd
	// and replace the installed binary.
	if err := runSystemctl("stop", serviceName); err != nil {
		// A stop command can report an error even when the service stopped.
		if !waitStopped(15 * time.Second) {
			return fmt.Errorf("stop service: %w", err)
		}
	}

	if !waitStopped(20 * time.Second) {
		return fmt.Errorf("service did not stop")
	}

	backup := target + ".previous"
	_ = os.Remove(backup)

	if err := os.Rename(target, backup); err != nil {
		_ = runSystemctl("start", serviceName)
		return fmt.Errorf("move old executable: %w", err)
	}

	if err := os.Rename(tmp, target); err != nil {
		rollbackErr := os.Rename(backup, target)
		startErr := runSystemctl("start", serviceName)
		return combineErrors(
			fmt.Errorf("install new executable: %w", err),
			rollbackErr, startErr,
		)
	}

	if err := os.Chmod(target, 0755); err != nil {
		return rollbackLinux(target, backup, err)
	}

	if err := runSystemctl("start", serviceName); err != nil {
		return rollbackLinux(target, backup, err)
	}

	if !waitActive(30 * time.Second) {
		return rollbackLinux(
			target, backup,
			fmt.Errorf("new service did not become active"),
		)
	}

	// Keep the backup for a later cleanup/rollback policy.
	return nil
}

func rollbackLinux(target, backup string, cause error) error {
	_ = runSystemctl("stop", serviceName)
	_ = os.Remove(target)

	if err := os.Rename(backup, target); err != nil {
		return combineErrors(cause, fmt.Errorf("restore backup: %w", err))
	}

	if err := runSystemctl("start", serviceName); err != nil {
		return combineErrors(cause, fmt.Errorf("restart old version: %w", err))
	}

	return cause
}

func runSystemctl(args ...string) error {
	cmd := exec.Command("/usr/bin/systemctl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"systemctl %v failed: %w: %s",
			args, err, string(out),
		)
	}
	return nil
}

func waitStopped(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		cmd := exec.Command("/usr/bin/systemctl", "is-active", "--quiet", serviceName)
		if err := cmd.Run(); err != nil {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

func waitActive(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		cmd := exec.Command("/usr/bin/systemctl", "is-active", "--quiet", serviceName)
		if err := cmd.Run(); err == nil {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}
