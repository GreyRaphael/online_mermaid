//go:build windows

package updater

import (
	"log/slog"
	"os"
	"os/exec"
	"time"
)

// RestartServer restarts the process on Windows by spawning a detached process and exiting.
func RestartServer(execPath string) {
	go func() {
		// Wait 500ms to allow client response to finish sending over HTTP
		time.Sleep(500 * time.Millisecond)
		slog.Info("restarting server on Windows", "executable", execPath)

		cmd := exec.Command(execPath, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()

		if err := cmd.Start(); err != nil {
			slog.Error("failed to start new process on Windows", "error", err)
		}
		os.Exit(0)
	}()
}
