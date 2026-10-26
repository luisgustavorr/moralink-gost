//go:build windows

package updater

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/kardianos/service"
)

func launchApply(tmp, target string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	src, err := os.Open(self)
	if err != nil {
		return err
	}
	defer src.Close()

	// Create a separate executable image for the updater helper.
	helper, err := os.CreateTemp(os.TempDir(), "moralink-updater-*.exe")
	if err != nil {
		return err
	}
	helperPath := helper.Name()

	_, copyErr := io.Copy(helper, src)
	closeErr := helper.Close()

	if copyErr != nil {
		_ = os.Remove(helperPath)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(helperPath)
		return closeErr
	}

	cmd := exec.Command(helperPath, applyArg, tmp, target)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x00000008, // DETACHED_PROCESS
	}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		_ = os.Remove(helperPath)
		return fmt.Errorf("start updater helper: %w", err)
	}

	// Do not wait for the helper; it must stop this service.
	return nil
}

func applyUpdate(tmp, target string) error {
	if err := validatePaths(tmp, target); err != nil {
		return err
	}

	svc, err := service.New(nil, &service.Config{
		Name: serviceName,
	})
	if err != nil {
		return fmt.Errorf("open Windows service: %w", err)
	}

	// Request a graceful service stop through the Service Control Manager.
	if err := svc.Stop(); err != nil {
		if !waitStoppedWindows(svc, 20*time.Second) {
			return fmt.Errorf("stop Windows service: %w", err)
		}
	}

	if !waitStoppedWindows(svc, 20*time.Second) {
		return fmt.Errorf("Windows service did not stop")
	}

	backup := target + ".previous"
	_ = os.Remove(backup)

	if err := os.Rename(target, backup); err != nil {
		_ = svc.Start()
		return fmt.Errorf("move old executable: %w", err)
	}

	if err := os.Rename(tmp, target); err != nil {
		rollbackErr := os.Rename(backup, target)
		startErr := svc.Start()
		return combineErrors(
			fmt.Errorf("install new executable: %w", err),
			rollbackErr, startErr,
		)
	}

	if err := svc.Start(); err != nil {
		return rollbackWindows(svc, target, backup, err)
	}

	if !waitRunningWindows(svc, 30*time.Second) {
		return rollbackWindows(
			svc, target, backup,
			fmt.Errorf("new service did not become running"),
		)
	}

	// Keep the previous executable for recovery.
	return nil
}

func rollbackWindows(
	svc service.Service,
	target, backup string,
	cause error,
) error {
	_ = svc.Stop()
	_ = waitStoppedWindows(svc, 20*time.Second)

	// Never overwrite the backup with the failed version.
	failed := target + ".failed"
	_ = os.Remove(failed)

	if err := os.Rename(target, failed); err != nil {
		return combineErrors(
			cause,
			fmt.Errorf("preserve failed executable: %w", err),
		)
	}

	if err := os.Rename(backup, target); err != nil {
		// The old binary remains at backup if restoration fails.
		return combineErrors(
			cause,
			fmt.Errorf("restore backup: %w", err),
		)
	}

	if err := svc.Start(); err != nil {
		return combineErrors(
			cause,
			fmt.Errorf("restart previous version: %w", err),
		)
	}

	return cause
}

func waitStoppedWindows(svc service.Service, timeout time.Duration) bool {
	return waitWindowsStatus(svc, service.StatusStopped, timeout)
}

func waitRunningWindows(svc service.Service, timeout time.Duration) bool {
	return waitWindowsStatus(svc, service.StatusRunning, timeout)
}

func waitWindowsStatus(
	svc service.Service,
	want service.Status,
	timeout time.Duration,
) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		status, err := svc.Status()
		if err == nil && status == want {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}

	return false
}
