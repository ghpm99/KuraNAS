package autoshutdown

import (
	"strings"
	"testing"
)

func TestBuildShutdownCommandWindowsIsImmediate(t *testing.T) {
	cmd := buildShutdownCommand("windows")
	if cmd == nil {
		t.Fatal("expected a command for windows")
	}
	if got := strings.Join(cmd.Args, " "); got != "shutdown /s /t 0" {
		t.Fatalf("unexpected windows args: %q", got)
	}
}

func TestBuildShutdownCommandLinuxIsImmediate(t *testing.T) {
	for _, goos := range []string{"linux", "darwin"} {
		cmd := buildShutdownCommand(goos)
		if got := strings.Join(cmd.Args, " "); got != "shutdown -h now" {
			t.Fatalf("unexpected %s args: %q", goos, got)
		}
	}
}

func TestBuildShutdownCommandUnsupported(t *testing.T) {
	if cmd := buildShutdownCommand("plan9"); cmd != nil {
		t.Fatalf("expected nil command for unsupported OS, got %v", cmd.Args)
	}
}
