package updater

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

func resetUpdaterOSFns() {
	osExecutableFunc = os.Executable
	evalSymlinksFunc = filepath.EvalSymlinks
	osRenameFunc = os.Rename
	osOpenFunc = os.Open
	osCreateFunc = os.Create
	osRemoveFunc = os.Remove
	osChmodFunc = os.Chmod
	osStartProcessFunc = os.StartProcess
	osExitFunc = os.Exit
	syscallExecFunc = syscall.Exec
	runtimeGOOS = runtime.GOOS
}

func TestApplyUpdateErrorBranches(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	osExecutableFunc = func() (string, error) {
		return "", errors.New("exec path error")
	}
	if err := applyUpdate("/tmp/new-bin"); err == nil || !strings.Contains(err.Error(), "failed to get current executable path") {
		t.Fatalf("expected executable path error, got %v", err)
	}

	osExecutableFunc = func() (string, error) {
		return "/tmp/current", nil
	}
	evalSymlinksFunc = func(path string) (string, error) {
		return "", errors.New("symlink error")
	}
	if err := applyUpdate("/tmp/new-bin"); err == nil || !strings.Contains(err.Error(), "failed to resolve symlinks") {
		t.Fatalf("expected eval symlink error, got %v", err)
	}
}

func TestApplyUpdateReplacesBinaryAndRollsBackOnOpenFailure(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	tmpDir := t.TempDir()
	currentPath := filepath.Join(tmpDir, "current-bin")
	newPath := filepath.Join(tmpDir, "new-bin")

	if err := os.WriteFile(currentPath, []byte("old-content"), 0755); err != nil {
		t.Fatalf("failed to create current binary: %v", err)
	}
	if err := os.WriteFile(newPath, []byte("new-content"), 0755); err != nil {
		t.Fatalf("failed to create new binary: %v", err)
	}

	osExecutableFunc = func() (string, error) {
		return currentPath, nil
	}
	evalSymlinksFunc = func(path string) (string, error) {
		return path, nil
	}

	if err := applyUpdate(newPath); err != nil {
		t.Fatalf("expected applyUpdate success, got %v", err)
	}
	data, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("failed to read replaced binary: %v", err)
	}
	if string(data) != "new-content" {
		t.Fatalf("unexpected replaced content: %q", string(data))
	}
	if _, err := os.Stat(currentPath + ".old"); err != nil {
		t.Fatalf("expected old backup file to exist: %v", err)
	}

	if err := os.WriteFile(currentPath, []byte("original"), 0755); err != nil {
		t.Fatalf("failed to reset current binary: %v", err)
	}
	if err := applyUpdate(filepath.Join(tmpDir, "missing-bin")); err == nil || !strings.Contains(err.Error(), "failed to open new binary") {
		t.Fatalf("expected open new binary error, got %v", err)
	}
	restored, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("failed to read restored binary: %v", err)
	}
	if string(restored) != "original" {
		t.Fatalf("expected rollback to keep original content, got %q", string(restored))
	}
}

func TestRestartProcessBranches(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	osExecutableFunc = func() (string, error) {
		return "", errors.New("path error")
	}
	restartProcess()

	called := false
	osExecutableFunc = func() (string, error) {
		return "/tmp/kuranas", nil
	}
	runtimeGOOS = "linux"
	syscallExecFunc = func(path string, args []string, env []string) error {
		called = true
		if path != "/tmp/kuranas" {
			t.Fatalf("unexpected exec path: %s", path)
		}
		return nil
	}

	restartProcess()
	if !called {
		t.Fatalf("expected syscall exec to be called on linux branch")
	}

	startCalled := false
	exitCalled := false
	runtimeGOOS = "windows"
	osStartProcessFunc = func(name string, argv []string, attr *os.ProcAttr) (*os.Process, error) {
		startCalled = true
		return &os.Process{}, nil
	}
	osExitFunc = func(code int) {
		exitCalled = true
	}
	restartProcess()
	if !startCalled || !exitCalled {
		t.Fatalf("expected windows branch to start process and exit")
	}
}

func TestApplyUpdateAdditionalErrorBranches(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	tmpDir := t.TempDir()
	currentPath := filepath.Join(tmpDir, "current-bin")
	newPath := filepath.Join(tmpDir, "new-bin")

	if err := os.WriteFile(currentPath, []byte("old"), 0755); err != nil {
		t.Fatalf("failed to create current binary: %v", err)
	}
	if err := os.WriteFile(newPath, []byte("new"), 0755); err != nil {
		t.Fatalf("failed to create new binary: %v", err)
	}

	osExecutableFunc = func() (string, error) { return currentPath, nil }
	evalSymlinksFunc = func(path string) (string, error) { return path, nil }

	osRenameFunc = func(oldpath, newpath string) error { return errors.New("rename failed") }
	if err := applyUpdate(newPath); err == nil || !strings.Contains(err.Error(), "failed to rename current binary") {
		t.Fatalf("expected rename error, got %v", err)
	}
	osRenameFunc = os.Rename

	osCreateFunc = func(name string) (*os.File, error) { return nil, errors.New("create failed") }
	if err := applyUpdate(newPath); err == nil || !strings.Contains(err.Error(), "failed to create new binary") {
		t.Fatalf("expected create error, got %v", err)
	}
	osCreateFunc = os.Create

	osCreateFunc = func(name string) (*os.File, error) {
		f, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		_ = f.Close()
		return f, nil
	}
	if err := applyUpdate(newPath); err == nil || !strings.Contains(err.Error(), "failed to copy new binary") {
		t.Fatalf("expected copy error, got %v", err)
	}
	osCreateFunc = os.Create

	osOpenFunc = func(name string) (*os.File, error) {
		return nil, io.EOF
	}
	if err := applyUpdate(newPath); err == nil || !strings.Contains(err.Error(), "failed to open new binary") {
		t.Fatalf("expected open error, got %v", err)
	}
	osOpenFunc = os.Open

	if runtime.GOOS != "windows" {
		osChmodFunc = func(name string, mode os.FileMode) error { return errors.New("chmod failed") }
		if err := applyUpdate(newPath); err == nil || !strings.Contains(err.Error(), "failed to set executable permissions") {
			t.Fatalf("expected chmod error, got %v", err)
		}
	}
}
