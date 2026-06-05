//go:build windows

package applog

import (
	"os"

	"golang.org/x/sys/windows"
)

// RedirectStderr points the process standard error at file so the Go runtime's
// uncaught-panic and fatal-error stack traces (written directly to the OS
// stderr handle, bypassing the slog/log writers) are captured in the forensic
// file instead of being lost. Best-effort: a failure is non-fatal.
func RedirectStderr(file *os.File) error {
	if file == nil {
		return nil
	}
	if err := windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(file.Fd())); err != nil {
		return err
	}
	os.Stderr = file
	return nil
}
