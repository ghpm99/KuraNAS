package applog

import (
	"bytes"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestSetupRoutesStructuredAndStdLogToWriter(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Writer: &buf, Level: slog.LevelInfo})

	Info("structured line", "job_id", 42)
	log.Println("legacy line")

	out := buf.String()
	if !strings.Contains(out, "structured line") {
		t.Fatalf("expected slog line in output, got: %q", out)
	}
	if !strings.Contains(out, "job_id=42") {
		t.Fatalf("expected structured field in output, got: %q", out)
	}
	if !strings.Contains(out, "legacy line") {
		t.Fatalf("expected bridged std log line in output, got: %q", out)
	}
}

func TestSetLevelFiltersBelowThreshold(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Writer: &buf, Level: slog.LevelInfo})

	SetLevel(slog.LevelWarn)
	Info("should be filtered")
	Warn("should pass")

	out := buf.String()
	if strings.Contains(out, "should be filtered") {
		t.Fatalf("info line should have been filtered at WARN level, got: %q", out)
	}
	if !strings.Contains(out, "should pass") {
		t.Fatalf("warn line should have passed, got: %q", out)
	}
}

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"DEBUG":   slog.LevelDebug,
		"debug":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"":        slog.LevelInfo,
		"garbage": slog.LevelInfo,
		"WARN":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		"ERROR":   slog.LevelError,
	}
	for in, want := range cases {
		if got := ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestRecoverCapturesPanicWithStack(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Writer: &buf, Level: slog.LevelDebug})

	Recover("unit-test", func() {
		panic("boom")
	})

	out := buf.String()
	if !strings.Contains(out, "panic recovered") {
		t.Fatalf("expected panic to be logged, got: %q", out)
	}
	if !strings.Contains(out, "boom") {
		t.Fatalf("expected panic value in log, got: %q", out)
	}
	if !strings.Contains(out, "goroutine=unit-test") {
		t.Fatalf("expected goroutine name in log, got: %q", out)
	}
	if !strings.Contains(out, "stack=") {
		t.Fatalf("expected stack trace field in log, got: %q", out)
	}
}

func TestGoRunsWithRecover(t *testing.T) {
	var buf bytes.Buffer
	Setup(Options{Writer: &buf, Level: slog.LevelDebug})

	var wg sync.WaitGroup
	wg.Add(1)
	Go("bg", func() {
		defer wg.Done()
		panic("async boom")
	})
	wg.Wait()

	if !strings.Contains(buf.String(), "async boom") {
		t.Fatalf("expected async panic to be recovered and logged, got: %q", buf.String())
	}
}

func TestRotatingFileRollsOverOnSize(t *testing.T) {
	dir := t.TempDir()
	rf, err := NewRotatingFile(RotateConfig{
		Dir:        dir,
		Prefix:     "test-",
		MaxSizeMB:  1, // 1 MiB threshold; we cross it with raw writes below
		MaxBackups: 5,
	})
	if err != nil {
		t.Fatalf("NewRotatingFile: %v", err)
	}

	chunk := bytes.Repeat([]byte("x"), 256*1024) // 256 KiB
	for i := 0; i < 8; i++ {                     // ~2 MiB total -> at least one rotation
		if _, err := rf.Write(chunk); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "test-") && strings.HasSuffix(e.Name(), ".log") {
			count++
		}
	}
	if count < 2 {
		t.Fatalf("expected at least 2 log files after rotation, got %d", count)
	}
}

func TestRotatingFilePrunesBeyondMaxBackups(t *testing.T) {
	dir := t.TempDir()

	// Seed three stale rotated files; the active file plus these would exceed
	// MaxBackups=1, so pruning must remove the oldest stale ones.
	for _, name := range []string{"test-2020-01-01_00-00-00.log", "test-2020-01-02_00-00-00.log", "test-2020-01-03_00-00-00.log"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("old"), 0o644); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	rf, err := NewRotatingFile(RotateConfig{Dir: dir, Prefix: "test-", MaxSizeMB: 1, MaxBackups: 1})
	if err != nil {
		t.Fatalf("NewRotatingFile: %v", err)
	}
	// Force a rotation so prune runs against the seeded files.
	if err := rf.rotate(); err != nil {
		t.Fatalf("rotate: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "test-") {
			count++
		}
	}
	// active file + at most MaxBackups kept.
	if count > 2 {
		t.Fatalf("expected pruning to cap files at 2 (active + 1 backup), got %d", count)
	}
}
