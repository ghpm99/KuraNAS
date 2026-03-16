package updater

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func resetUpdaterOSFns() {
	osExecutableFunc = os.Executable
	evalSymlinksFunc = filepath.EvalSymlinks
	osRenameFunc = os.Rename
	osRemoveAllFunc = os.RemoveAll
	osMkdirAllFunc = os.MkdirAll
	runtimeGOOS = runtime.GOOS
}

func TestGetInstallDirErrorBranches(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	osExecutableFunc = func() (string, error) {
		return "", errors.New("exec path error")
	}
	_, err := getInstallDir()
	if err == nil || err.Error() != "failed to get executable path: exec path error" {
		t.Fatalf("expected executable path error, got %v", err)
	}

	osExecutableFunc = func() (string, error) {
		return "/tmp/current", nil
	}
	evalSymlinksFunc = func(path string) (string, error) {
		return "", errors.New("symlink error")
	}
	_, err = getInstallDir()
	if err == nil || err.Error() != "failed to resolve symlinks: symlink error" {
		t.Fatalf("expected eval symlink error, got %v", err)
	}
}

func TestApplyBinaryUpdateSuccess(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "install")
	extractedDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(installDir, 0755)
	os.MkdirAll(extractedDir, 0755)

	binName := "kuranas"
	if runtime.GOOS == "windows" {
		binName = "kuranas.exe"
	}

	currentBin := filepath.Join(installDir, binName)
	newBin := filepath.Join(extractedDir, binName)

	os.WriteFile(currentBin, []byte("old-binary"), 0755)
	os.WriteFile(newBin, []byte("new-binary"), 0755)

	if err := applyBinaryUpdate(extractedDir, installDir); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	data, err := os.ReadFile(currentBin)
	if err != nil {
		t.Fatalf("failed to read binary: %v", err)
	}
	if string(data) != "new-binary" {
		t.Fatalf("expected new-binary, got %q", string(data))
	}

	if _, err := os.Stat(currentBin + ".old"); err != nil {
		t.Fatalf("expected .old backup to exist: %v", err)
	}
}

func TestApplyBinaryUpdateMissingNewBinary(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	extractedDir := t.TempDir()
	installDir := t.TempDir()

	if err := applyBinaryUpdate(extractedDir, installDir); err == nil {
		t.Fatalf("expected error for missing new binary")
	}
}

func TestApplyBinaryUpdateRenameError(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "install")
	extractedDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(installDir, 0755)
	os.MkdirAll(extractedDir, 0755)

	binName := "kuranas"
	if runtime.GOOS == "windows" {
		binName = "kuranas.exe"
	}

	os.WriteFile(filepath.Join(installDir, binName), []byte("old"), 0755)
	os.WriteFile(filepath.Join(extractedDir, binName), []byte("new"), 0755)

	osRenameFunc = func(oldpath, newpath string) error {
		return errors.New("rename failed")
	}

	if err := applyBinaryUpdate(extractedDir, installDir); err == nil {
		t.Fatalf("expected rename error")
	}
}

func TestApplyBinaryUpdateCopyErrorRollsBack(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "install")
	extractedDir := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(installDir, 0755)
	os.MkdirAll(extractedDir, 0755)

	binName := "kuranas"
	if runtime.GOOS == "windows" {
		binName = "kuranas.exe"
	}

	currentBin := filepath.Join(installDir, binName)
	os.WriteFile(currentBin, []byte("original"), 0755)
	// Create new binary as a directory to force copy failure
	os.MkdirAll(filepath.Join(extractedDir, binName), 0755)

	err := applyBinaryUpdate(extractedDir, installDir)
	if err == nil {
		t.Fatalf("expected copy error")
	}

	// Check that rollback happened
	data, _ := os.ReadFile(currentBin)
	if string(data) != "original" {
		t.Fatalf("expected rollback to restore original binary, got %q", string(data))
	}
}

func TestUpdateScriptsDirPreservesVenv(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src-scripts")
	dstDir := filepath.Join(tmpDir, "dst-scripts")

	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(dstDir, 0755)

	// Create existing .venv
	venvDir := filepath.Join(dstDir, ".venv")
	os.MkdirAll(venvDir, 0755)
	os.WriteFile(filepath.Join(venvDir, "python"), []byte("python-bin"), 0755)

	// Create new script files
	os.WriteFile(filepath.Join(srcDir, "script.py"), []byte("new-script"), 0644)

	if err := updateScriptsDir(srcDir, dstDir); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	// Verify new script exists
	data, err := os.ReadFile(filepath.Join(dstDir, "script.py"))
	if err != nil {
		t.Fatalf("expected new script to exist: %v", err)
	}
	if string(data) != "new-script" {
		t.Fatalf("expected new-script, got %q", string(data))
	}

	// Verify .venv was preserved
	venvData, err := os.ReadFile(filepath.Join(venvDir, "python"))
	if err != nil {
		t.Fatalf("expected .venv to be preserved: %v", err)
	}
	if string(venvData) != "python-bin" {
		t.Fatalf("expected python-bin in .venv, got %q", string(venvData))
	}
}

func TestApplyFullUpdateCopiesAllAssets(t *testing.T) {
	resetUpdaterOSFns()
	t.Cleanup(resetUpdaterOSFns)

	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "install")
	extractedDir := filepath.Join(tmpDir, "extracted")

	binName := "kuranas"
	if runtime.GOOS == "windows" {
		binName = "kuranas.exe"
	}

	// Setup install dir with old files
	os.MkdirAll(filepath.Join(installDir, "dist"), 0755)
	os.WriteFile(filepath.Join(installDir, "dist", "old.js"), []byte("old"), 0644)
	os.WriteFile(filepath.Join(installDir, binName), []byte("old-bin"), 0755)

	// Setup extracted dir with new files
	os.MkdirAll(filepath.Join(extractedDir, "dist"), 0755)
	os.MkdirAll(filepath.Join(extractedDir, "icons"), 0755)
	os.MkdirAll(filepath.Join(extractedDir, "translations"), 0755)
	os.WriteFile(filepath.Join(extractedDir, binName), []byte("new-bin"), 0755)
	os.WriteFile(filepath.Join(extractedDir, "dist", "new.js"), []byte("new"), 0644)
	os.WriteFile(filepath.Join(extractedDir, "icons", "icon.png"), []byte("icon"), 0644)
	os.WriteFile(filepath.Join(extractedDir, "translations", "en.json"), []byte("{}"), 0644)

	osExecutableFunc = func() (string, error) {
		return filepath.Join(installDir, binName), nil
	}
	evalSymlinksFunc = func(path string) (string, error) {
		return path, nil
	}

	if err := applyFullUpdate(extractedDir); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	// Verify binary was updated
	binData, _ := os.ReadFile(filepath.Join(installDir, binName))
	if string(binData) != "new-bin" {
		t.Fatalf("expected new-bin, got %q", string(binData))
	}

	// Verify old dist was replaced
	if _, err := os.Stat(filepath.Join(installDir, "dist", "old.js")); !os.IsNotExist(err) {
		t.Fatalf("expected old.js to be removed")
	}

	// Verify new assets exist
	newJS, _ := os.ReadFile(filepath.Join(installDir, "dist", "new.js"))
	if string(newJS) != "new" {
		t.Fatalf("expected new dist content")
	}

	iconData, _ := os.ReadFile(filepath.Join(installDir, "icons", "icon.png"))
	if string(iconData) != "icon" {
		t.Fatalf("expected icon content")
	}

	transData, _ := os.ReadFile(filepath.Join(installDir, "translations", "en.json"))
	if string(transData) != "{}" {
		t.Fatalf("expected translation content")
	}
}

func TestShutdownFnCalled(t *testing.T) {
	service := NewService()

	called := false
	service.SetShutdownFn(func() {
		called = true
	})

	if service.shutdownFn == nil {
		t.Fatalf("expected shutdownFn to be set")
	}

	service.shutdownFn()
	if !called {
		t.Fatalf("expected shutdownFn to be called")
	}
}
