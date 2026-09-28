//go:build !windows

package updater

import (
	"log/slog"
	"os"
	"syscall"
	"time"
)

// RestartServer gracefully restarts the process by in-place exec.
func RestartServer(execPath string) {
	go func() {
		// Wait 500ms to allow client response to finish sending over HTTP
		time.Sleep(500 * time.Millisecond)
		slog.Info("restarting server via syscall.Exec", "executable", execPath)

		err := syscall.Exec(execPath, os.Args, os.Environ())
		if err != nil {
			slog.Error("syscall.Exec failed, exiting process for supervisor restart", "error", err)
			os.Exit(0)
		}
	}()
}
