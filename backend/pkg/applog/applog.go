// Package applog configures the process-wide structured logger (slog) that
// produces the forensic file log: every level, with source location and
// correlation fields (job_id, step_id, path). It also bridges the standard
// library log package so the existing log.Printf calls keep landing in the
// same file, serialized so their lines never interleave with slog records.
package applog

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// syncWriter serializes writes from the slog handler and the bridged std log
// package so their lines never interleave in the shared destination.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

// Options configures Setup.
type Options struct {
	// Writer is the log destination (the rotating file in production,
	// os.Stdout in dev). Defaults to os.Stdout when nil.
	Writer io.Writer
	// Level is the minimum level emitted. Adjustable later via SetLevel.
	Level slog.Level
	// AddSource records the caller file:line on every record.
	AddSource bool
}

var (
	current  = &syncWriter{w: os.Stdout}
	levelVar = new(slog.LevelVar) // INFO by default
)

// Setup installs a slog text handler as the default logger and routes the
// standard log package through the same synchronized writer, so both
// structured (slog) and legacy (log.Printf) lines share one ordered stream.
// It is safe to call more than once (the destination/level are swapped in).
func Setup(opts Options) {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	current.mu.Lock()
	current.w = opts.Writer
	current.mu.Unlock()

	levelVar.Set(opts.Level)

	handler := slog.NewTextHandler(current, &slog.HandlerOptions{
		Level:     levelVar,
		AddSource: opts.AddSource,
	})
	slog.SetDefault(slog.New(handler))

	// Legacy code still calls the std log package. Point it at the same writer
	// and keep its UTC timestamp so those lines remain readable next to the
	// structured ones during the gradual migration to slog.
	log.SetOutput(current)
	log.SetFlags(log.LstdFlags | log.LUTC)
}

// SetLevel adjusts the minimum emitted level at runtime (e.g. after the .env
// LOG_LEVEL is loaded, which happens after the early file logger is installed).
func SetLevel(level slog.Level) {
	levelVar.Set(level)
}

// ParseLevel maps a human level name (DEBUG/INFO/WARN/ERROR) to slog.Level,
// defaulting to INFO for empty or unknown values.
func ParseLevel(name string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Debug/Info/Warn/Error log a structured record at the given level. They record
// the immediate caller as the source (not this wrapper) so file:line stays
// accurate. Extra fields are passed as alternating key/value args:
//
//	applog.Info("job finished", "job_id", id, "status", status)
func Debug(msg string, args ...any) { logAt(slog.LevelDebug, msg, args...) }
func Info(msg string, args ...any)  { logAt(slog.LevelInfo, msg, args...) }
func Warn(msg string, args ...any)  { logAt(slog.LevelWarn, msg, args...) }
func Error(msg string, args ...any) { logAt(slog.LevelError, msg, args...) }

func logAt(level slog.Level, msg string, args ...any) {
	logger := slog.Default()
	ctx := context.Background()
	if !logger.Enabled(ctx, level) {
		return
	}
	// Skip [Callers, logAt, exported wrapper] so source points at the caller.
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	record := slog.NewRecord(time.Now(), level, msg, pcs[0])
	record.Add(args...)
	_ = logger.Handler().Handle(ctx, record)
}

// Recover runs fn and turns any panic into an ERROR log line carrying the
// stack trace, so a panic in a goroutine is captured in the forensic file
// instead of crashing the whole process. name identifies the goroutine.
func Recover(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			Error("panic recovered",
				"goroutine", name,
				"panic", fmt.Sprint(r),
				"stack", string(debug.Stack()),
			)
		}
	}()
	fn()
}

// Go starts fn in a new goroutine guarded by Recover. Use it for fire-and-forget
// work that must not take the process down if it panics.
func Go(name string, fn func()) {
	go Recover(name, fn)
}

// GoRestart starts a long-lived loop body in its own goroutine; if it panics it
// is logged and restarted after a short delay, so a perpetual loop (worker
// pool, scheduler) survives a single bad iteration. It stops when done is
// closed or returns a non-nil signal via the returned stop function.
func GoRestart(name string, fn func()) {
	go func() {
		for {
			panicked := runGuarded(name, fn)
			if !panicked {
				return
			}
			time.Sleep(time.Second)
			Warn("goroutine restarting after panic", "goroutine", name)
		}
	}()
}

// runGuarded runs fn and reports whether it panicked (true) or returned
// normally (false).
func runGuarded(name string, fn func()) (panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			Error("panic recovered",
				"goroutine", name,
				"panic", fmt.Sprint(r),
				"stack", string(debug.Stack()),
			)
		}
	}()
	fn()
	return false
}
