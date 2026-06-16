package autoshutdown

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

// ExecuteShutdown powers the host machine off after graceSeconds, choosing the
// command for the running OS. In production this runs on Windows; dev/test runs
// on Linux, so both are handled. graceSeconds is clamped to a sane range.
func ExecuteShutdown(graceSeconds int) error {
	if graceSeconds < 0 {
		graceSeconds = 0
	}
	if graceSeconds > maxGracePeriodSeconds {
		graceSeconds = maxGracePeriodSeconds
	}

	cmd := buildShutdownCommand(runtime.GOOS, graceSeconds)
	if cmd == nil {
		return fmt.Errorf("autoshutdown: unsupported platform %q", runtime.GOOS)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("autoshutdown: shutdown command failed: %w", err)
	}
	return nil
}

// buildShutdownCommand returns the OS-specific shutdown command. Windows takes
// the delay in seconds (`/t`); Linux/macOS shutdown takes minutes, so the grace
// is rounded up to the next whole minute (or "now" when zero). goos is a
// parameter (not read directly) so every branch is unit-testable.
func buildShutdownCommand(goos string, graceSeconds int) *exec.Cmd {
	switch goos {
	case "windows":
		return exec.Command("shutdown", "/s", "/t", strconv.Itoa(graceSeconds))
	case "linux", "darwin":
		when := "now"
		if graceSeconds > 0 {
			when = "+" + strconv.Itoa((graceSeconds+59)/60)
		}
		return exec.Command("shutdown", "-h", when)
	default:
		return nil
	}
}
