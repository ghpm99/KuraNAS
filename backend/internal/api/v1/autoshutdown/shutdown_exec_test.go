package autoshutdown

import (
	"strings"
	"testing"
)

func TestBuildShutdownCommandWindows(t *testing.T) {
	cmd := buildShutdownCommand("windows", 60)
	if cmd == nil {
		t.Fatal("expected a command for windows")
	}
	if got := strings.Join(cmd.Args, " "); !strings.Contains(got, "shutdown /s /t 60") {
		t.Fatalf("unexpected windows args: %q", got)
	}
}

func TestBuildShutdownCommandLinuxRoundsToMinutes(t *testing.T) {
	cmd := buildShutdownCommand("linux", 61) // 61s -> 2 minutes
	if got := strings.Join(cmd.Args, " "); !strings.Contains(got, "shutdown -h +2") {
		t.Fatalf("unexpected linux args: %q", got)
	}
}

func TestBuildShutdownCommandLinuxImmediate(t *testing.T) {
	cmd := buildShutdownCommand("linux", 0)
	if got := strings.Join(cmd.Args, " "); !strings.Contains(got, "shutdown -h now") {
		t.Fatalf("unexpected immediate args: %q", got)
	}
}

func TestBuildShutdownCommandUnsupported(t *testing.T) {
	if cmd := buildShutdownCommand("plan9", 0); cmd != nil {
		t.Fatalf("expected nil command for unsupported OS, got %v", cmd.Args)
	}
}
