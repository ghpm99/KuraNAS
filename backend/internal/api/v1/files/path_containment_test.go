package files

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

var traversalAttempts = []string{
	"../outside",
	"docs/../../outside",
	`..\outside`,
	`docs\..\..\outside`,
	"/../outside",
	"docs/../..",
	"..",
}

func TestResolveContainedPathRejectsTraversal(t *testing.T) {
	setEntryPointForTest(t, t.TempDir())

	for _, attempt := range traversalAttempts {
		if _, err := resolveContainedPath(attempt); err == nil {
			t.Errorf("expected %q to be rejected", attempt)
		}
	}
}

func TestResolveContainedPathNeverResolvesAbsolutePathOutsideRoots(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)
	outsideDir := t.TempDir()

	resolvedPath, err := resolveContainedPath(outsideDir)
	if err != nil {
		return
	}
	if !isInsideDirectory(entryPoint, resolvedPath) {
		t.Fatalf("absolute path outside roots escaped to %q", resolvedPath)
	}
}

func TestResolveContainedPathAcceptsNormalPaths(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	for _, accepted := range []string{"", "docs/file.txt", "/docs/file.txt", `\docs\file.txt`, "a..b/c", filepath.Join(entryPoint, "docs")} {
		resolvedPath, err := resolveContainedPath(accepted)
		if err != nil {
			t.Errorf("expected %q to be accepted, got %v", accepted, err)
			continue
		}
		if !isInsideDirectory(entryPoint, resolvedPath) {
			t.Errorf("%q resolved outside entry point: %q", accepted, resolvedPath)
		}
	}
}

func TestResolveContainedChildPathRejectsNestedOrTraversingNames(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	for _, name := range []string{"", ".", "..", "../x", "a/b", `a\b`, `..\x`} {
		if _, err := resolveContainedChildPath(entryPoint, name); err == nil {
			t.Errorf("expected child name %q to be rejected", name)
		}
	}
	if _, err := resolveContainedChildPath(entryPoint, "ok.txt"); err != nil {
		t.Fatalf("expected plain name accepted, got %v", err)
	}
}

func TestResolveColdContentPathRejectsUnsafePaths(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)
	coldDir := t.TempDir()

	for _, unsafe := range []string{"", "relative/cold.bin", coldDir + "/../escape", filepath.Join(entryPoint, "inside.bin")} {
		if _, err := resolveColdContentPath(unsafe); err == nil {
			t.Errorf("expected cold path %q to be rejected", unsafe)
		}
	}
	if _, err := resolveColdContentPath(filepath.Join(coldDir, "ok.bin")); err != nil {
		t.Fatalf("expected cold path accepted, got %v", err)
	}
}

func TestCopyColdFileRejectsPathsOutsideAllowedAreas(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)
	coldDir := t.TempDir()
	coldFile := filepath.Join(coldDir, "doc.bin")
	if err := os.WriteFile(coldFile, []byte("cold"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	escapeTarget := filepath.Join(t.TempDir(), "escaped.bin")

	if err := copyColdFile(coldFile, escapeTarget); err == nil {
		if _, statErr := os.Stat(escapeTarget); statErr == nil {
			t.Fatalf("destination outside roots must not be written")
		}
	}
	if _, err := os.Stat(escapeTarget); err == nil {
		t.Fatalf("destination outside roots must not be written")
	}
	if err := copyColdFile(coldDir+"/../doc.bin", filepath.Join(entryPoint, "out.bin")); err == nil {
		t.Fatalf("expected traversing cold path to be rejected")
	}
	if err := copyColdFile(coldFile, filepath.Join(entryPoint, "out.bin")); err != nil {
		t.Fatalf("expected valid cold copy, got %v", err)
	}
}

func newTraversalTestService(t *testing.T) (*Service, string) {
	t.Helper()
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)
	sourcePath := filepath.Join(entryPoint, "source.txt")
	writeExistingFile(t, sourcePath, "data")
	records := []FileModel{
		{ID: 1, Name: "source.txt", Path: sourcePath, Type: File},
		{ID: 2, Name: "evil", Path: entryPoint + "/../evil.txt", Type: File},
	}
	return newTestServiceWithFileRecords(t, entryPoint, records), entryPoint
}

func TestFileOperationsRejectTraversalDestinations(t *testing.T) {
	service, entryPoint := newTraversalTestService(t)

	for _, attempt := range traversalAttempts {
		_, moveErr := service.MoveFile(1, nil, attempt)
		requireOperationError(t, moveErr, http.StatusBadRequest, "ERROR_INVALID_PATH")

		_, copyErr := service.CopyFile(1, nil, attempt, "")
		requireOperationError(t, copyErr, http.StatusBadRequest, "ERROR_INVALID_PATH")
	}

	if _, err := os.Stat(filepath.Join(entryPoint, "source.txt")); err != nil {
		t.Fatalf("source must be untouched: %v", err)
	}
}

func TestCopyFileRejectsTraversalNewName(t *testing.T) {
	service, entryPoint := newTraversalTestService(t)

	for _, newName := range []string{"../escaped.txt", `..\escaped.txt`, "sub/escaped.txt"} {
		_, err := service.CopyFile(1, nil, "", newName)
		requireOperationError(t, err, http.StatusBadRequest, "ERROR_INVALID_PATH")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(entryPoint), "escaped.txt")); err == nil {
		t.Fatalf("copy escaped the entry point")
	}
}

func TestRenameFileRejectsTraversalNewName(t *testing.T) {
	service, _ := newTraversalTestService(t)

	for _, newName := range []string{"../escaped.txt", `..\escaped.txt`, "..", "sub/escaped.txt"} {
		if _, err := service.RenameFile(1, newName); err == nil {
			t.Errorf("expected rename to %q to be rejected", newName)
		}
	}
}

func TestCreateFolderRejectsTraversalName(t *testing.T) {
	service, _ := newTraversalTestService(t)

	for _, name := range []string{"../escaped", `..\escaped`, "a/b", ".."} {
		if _, err := service.CreateFolder(nil, name); err == nil {
			t.Errorf("expected folder name %q to be rejected", name)
		}
	}
}

func TestDeleteFileRejectsRecordOutsideRoots(t *testing.T) {
	service, _ := newTraversalTestService(t)

	err := service.DeleteFileFromDisk(2, true)
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_INVALID_PATH")
}

func TestMoveFileRejectsRecordOutsideRoots(t *testing.T) {
	service, _ := newTraversalTestService(t)

	_, err := service.MoveFile(2, nil, "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_INVALID_PATH")
}

func TestUploadRejectsTraversalRelativePaths(t *testing.T) {
	service, entryPoint := newTraversalTestService(t)

	for _, relativePath := range []string{"../x/a.txt", `..\x\a.txt`, "ok/../../x/a.txt"} {
		result, err := uploadSingle(t, service, "a.txt", "payload", UploadOptions{RelativePaths: []string{relativePath}})
		if err == nil && (len(result.Files) == 0 || result.Files[0].Status != UploadStatusFailed) {
			t.Errorf("expected upload with relative path %q to fail", relativePath)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(entryPoint), "x")); err == nil {
		t.Fatalf("upload escaped the entry point")
	}
}
