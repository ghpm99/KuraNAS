//go:build !windows

package applog

import "os"

// RedirectStderr reassigns os.Stderr to file so code that writes to it lands in
// the forensic file. On non-Windows targets (dev runs to stdout, the only
// production target is Windows) we keep this minimal and do not dup the OS
// stderr fd. Best-effort: never fatal.
func RedirectStderr(file *os.File) error {
	if file == nil {
		return nil
	}
	os.Stderr = file
	return nil
}
