package autoshutdown

import (
	"fmt"
	"os/exec"
	"runtime"
)

// ExecuteShutdown powers the host machine off immediately, choosing the command
// for the running OS. The warning period is counted by the scheduler, not by
// the OS, so it can be cancelled.
func ExecuteShutdown() error {
	cmd := buildShutdownCommand(runtime.GOOS)
	if cmd == nil {
		return fmt.Errorf("autoshutdown: unsupported platform %q", runtime.GOOS)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("autoshutdown: shutdown command failed: %w", err)
	}
	return nil
}

func buildShutdownCommand(goos string) *exec.Cmd {
	switch goos {
	case "windows":
		return exec.Command("shutdown", "/s", "/t", "0")
	case "linux", "darwin":
		return exec.Command("shutdown", "-h", "now")
	default:
		return nil
	}
}
