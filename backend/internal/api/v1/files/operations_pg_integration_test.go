package files

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/utils"
)

func insertRowForOps(t *testing.T, repo *Repository, name, path, parent string, fileType FileType) int {
	t.Helper()
	var id int
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		created, e := repo.CreateFile(tx, FileModel{
			Name:       name,
			Path:       path,
			ParentPath: parent,
			Format:     filepath.Ext(name),
			Size:       1,
			UpdatedAt:  time.Now(),
			CreatedAt:  time.Now(),
			Type:       fileType,
		})
		id = created.ID
		return e
	})
	if err != nil {
		t.Fatalf("insert row %q: %v", path, err)
	}
	return id
}

func activeRowByPath(t *testing.T, repo *Repository, path string) (FileModel, bool) {
	t.Helper()
	res, err := repo.GetActiveFilesByPath(path, 1, 10)
	if err != nil {
		t.Fatalf("GetFiles(Path=%q): %v", path, err)
	}
	if len(res.Items) == 0 {
		return FileModel{}, false
	}
	return res.Items[0], true
}

// TestPostgres_OperationsReflectImmediatelyWithoutWorkers is the acceptance
// proof for task 05: right after each file operation returns, a plain read
// against the database already shows the new state — no worker, watcher or
// rescan ever runs in this test (the service's task channel is drained by
// nobody; it is only buffered).
func TestPostgres_OperationsReflectImmediatelyWithoutWorkers(t *testing.T) {
	ctx := testutil.NewPostgresDB(t, "kuranas_files_it")
	repo := NewRepository(ctx)
	truncateHomeFile(t, repo)

	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	service := &Service{Repository: repo, Tasks: make(chan utils.Task, 32)}

	// Disk layout: library/{child.txt, sub/deep.txt} and an empty target/.
	libraryDir := filepath.Join(entryPoint, "library")
	subDir := filepath.Join(libraryDir, "sub")
	targetDir := filepath.Join(entryPoint, "target")
	for _, dir := range []string{subDir, targetDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("MkdirAll %q: %v", dir, err)
		}
	}
	childFile := filepath.Join(libraryDir, "child.txt")
	deepFile := filepath.Join(subDir, "deep.txt")
	for _, f := range []string{childFile, deepFile} {
		if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
			t.Fatalf("WriteFile %q: %v", f, err)
		}
	}

	libraryID := insertRowForOps(t, repo, "library", libraryDir, entryPoint, Directory)
	insertRowForOps(t, repo, "child.txt", childFile, libraryDir, File)
	insertRowForOps(t, repo, "sub", subDir, libraryDir, Directory)
	insertRowForOps(t, repo, "deep.txt", deepFile, subDir, File)
	targetID := insertRowForOps(t, repo, "target", targetDir, entryPoint, Directory)

	// MoveFile: directory moved under target/, descendants must follow at once.
	movedPath, err := service.MoveFile(libraryID, &targetID, "")
	if err != nil {
		t.Fatalf("MoveFile: %v", err)
	}
	if movedPath != filepath.Join(targetDir, "library") {
		t.Fatalf("unexpected moved path %q", movedPath)
	}
	movedRow, found := activeRowByPath(t, repo, movedPath)
	if !found || movedRow.ParentPath != targetDir {
		t.Fatalf("moved directory row not updated: found=%v row=%+v", found, movedRow)
	}
	movedDeep, found := activeRowByPath(t, repo, filepath.Join(movedPath, "sub", "deep.txt"))
	if !found {
		t.Fatalf("descendant row did not follow the move")
	}
	if movedDeep.ParentPath != filepath.Join(movedPath, "sub") {
		t.Fatalf("descendant parent_path not rewritten: %+v", movedDeep)
	}
	if _, stale := activeRowByPath(t, repo, libraryDir); stale {
		t.Fatalf("old directory path still active after move")
	}

	// RenameFile: directory renamed in place, descendants must follow at once.
	renamedPath, err := service.RenameFile(libraryID, "books")
	if err != nil {
		t.Fatalf("RenameFile: %v", err)
	}
	renamedRow, found := activeRowByPath(t, repo, renamedPath)
	if !found || renamedRow.Name != "books" {
		t.Fatalf("renamed row not updated: found=%v row=%+v", found, renamedRow)
	}
	if _, found = activeRowByPath(t, repo, filepath.Join(renamedPath, "child.txt")); !found {
		t.Fatalf("descendant row did not follow the rename")
	}

	// CreateFolder: row visible immediately.
	createdPath, err := service.CreateFolder(nil, "fresh")
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	createdRow, found := activeRowByPath(t, repo, createdPath)
	if !found || createdRow.Type != Directory {
		t.Fatalf("created folder row not visible: found=%v row=%+v", found, createdRow)
	}

	// DeleteFileFromDisk: whole subtree soft-deleted immediately.
	if err := service.DeleteFileFromDisk(libraryID); err != nil {
		t.Fatalf("DeleteFileFromDisk: %v", err)
	}
	for _, gone := range []string{
		renamedPath,
		filepath.Join(renamedPath, "child.txt"),
		filepath.Join(renamedPath, "sub"),
		filepath.Join(renamedPath, "sub", "deep.txt"),
	} {
		if _, stillActive := activeRowByPath(t, repo, gone); stillActive {
			t.Fatalf("row %q still active after delete", gone)
		}
	}
	deletedRes, err := repo.GetFilesByPathPrefix(renamedPath, 1, 50)
	if err != nil {
		t.Fatalf("GetFiles(deleted subtree): %v", err)
	}
	deletedCount := 0
	for _, row := range deletedRes.Items {
		if row.DeletedAt.Valid {
			deletedCount++
		}
	}
	if deletedCount != 4 {
		t.Fatalf("expected 4 soft-deleted rows in subtree, got %d", deletedCount)
	}
}
