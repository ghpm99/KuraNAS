package engine

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
)

func insertHomeFile(t *testing.T, dbCtx *database.DbContext, name, path, parent string) {
	t.Helper()
	repo := files.NewRepository(dbCtx)
	err := dbCtx.ExecTx(func(tx *sql.Tx) error {
		_, e := repo.CreateFile(tx, files.FileModel{
			Name:       name,
			Path:       path,
			ParentPath: parent,
			Format:     filepath.Ext(name),
			Size:       1,
			UpdatedAt:  time.Now(),
			CreatedAt:  time.Now(),
			Type:       files.File,
		})
		return e
	})
	if err != nil {
		t.Fatalf("insert home_file %q: %v", path, err)
	}
}

func isMarkedDeleted(t *testing.T, dbCtx *database.DbContext, path string) bool {
	t.Helper()
	var deletedAt sql.NullTime
	err := dbCtx.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow("SELECT deleted_at FROM home_file WHERE path = $1", path).Scan(&deletedAt)
	})
	if err != nil {
		t.Fatalf("read deleted_at for %q: %v", path, err)
	}
	return deletedAt.Valid
}

// setPhysicalPath turns an indexed row into a tiered one: the bytes now live at
// physicalPath while the logical path stays the row identity.
func setPhysicalPath(t *testing.T, dbCtx *database.DbContext, logicalPath, physicalPath string) {
	t.Helper()
	err := dbCtx.ExecTx(func(tx *sql.Tx) error {
		_, e := tx.Exec("UPDATE home_file SET physical_path = $1 WHERE path = $2", physicalPath, logicalPath)
		return e
	})
	if err != nil {
		t.Fatalf("set physical_path for %q: %v", logicalPath, err)
	}
}

// TestMarkDeletedStep_DetectsMissingFiles_Postgres is the positive control: when
// the stored paths are POSIX-style (no backslashes), the PathPrefix query finds
// the rows and mark_deleted correctly flags the file that no longer exists on
// disk while leaving the present one active.
func TestMarkDeletedStep_DetectsMissingFiles_Postgres(t *testing.T) {
	dbCtx := testutil.NewPostgresDB(t, "kuranas_worker_it")
	truncateWorkerAndFiles(t, dbCtx)

	root := t.TempDir()
	filesSvc := files.NewService(files.NewRepository(dbCtx), jobs.NewRepository(dbCtx), nil)
	workerCtx := &WorkerContext{FilesService: filesSvc}

	presentPath := filepath.Join(root, "present.txt")
	if err := os.WriteFile(presentPath, []byte("present"), 0o644); err != nil {
		t.Fatalf("write present file: %v", err)
	}
	ghostPath := filepath.Join(root, "ghost.txt") // never created on disk

	insertHomeFile(t, dbCtx, "present.txt", presentPath, root)
	insertHomeFile(t, dbCtx, "ghost.txt", ghostPath, root)

	payload, _ := marshalPayload(StepFilePayload{Path: root})
	if err := executeMarkDeletedStep(workerCtx, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("executeMarkDeletedStep returned error: %v", err)
	}

	if !isMarkedDeleted(t, dbCtx, ghostPath) {
		t.Fatalf("ghost file should have been marked deleted, but deleted_at is still NULL")
	}
	if isMarkedDeleted(t, dbCtx, presentPath) {
		t.Fatalf("present file must NOT be marked deleted")
	}
}

// TestMarkDeletedStep_SoftDeletesMissingFileWithWindowsPath_Postgres asserts the
// EXPECTED behaviour for a Windows-style stored path: a fully-populated, active
// row (deleted_at NULL) whose file is no longer present in the folder must be
// soft-deleted by mark_deleted — exactly like the POSIX case above.
//
// While the PathPrefix backslash defect is present, mark_deleted loads its
// candidates with `path LIKE $prefix || '%'`; PostgreSQL treats '\' as the LIKE
// escape character, so the prefix matches ZERO rows, the missing file is never
// identified, and this test FAILS — which is precisely the bug reproduced.
func TestMarkDeletedStep_SoftDeletesMissingFileWithWindowsPath_Postgres(t *testing.T) {
	dbCtx := testutil.NewPostgresDB(t, "kuranas_worker_it")
	truncateWorkerAndFiles(t, dbCtx)

	filesSvc := files.NewService(files.NewRepository(dbCtx), jobs.NewRepository(dbCtx), nil)
	workerCtx := &WorkerContext{FilesService: filesSvc}

	winRoot := `D:\Pasta`
	ghostPath := `D:\Pasta\ghost.pdf` // active row in the DB; file does not exist on disk
	insertHomeFile(t, dbCtx, "ghost.pdf", ghostPath, winRoot)

	payload, _ := marshalPayload(StepFilePayload{Path: winRoot})
	if err := executeMarkDeletedStep(workerCtx, jobs.StepModel{Payload: payload}); err != nil && err != ErrStepSkipped {
		t.Fatalf("executeMarkDeletedStep returned unexpected error: %v", err)
	}

	if !isMarkedDeleted(t, dbCtx, ghostPath) {
		t.Fatalf("expected the missing Windows-path file to be soft-deleted (deleted_at set), " +
			"but it is still active — mark_deleted failed to identify a DB row absent from the folder")
	}
}

// TestMarkDeletedStep_KeepsTieredFileActive_Postgres is the tiering blindage
// (task 13): a file migrated to cold storage has no bytes at its logical path,
// so a naive existence check would soft-delete every cold file on the next
// scan. The step must stat the resolved content path instead — keeping the
// tiered row active while still flagging a tiered file whose cold copy is
// genuinely gone.
func TestMarkDeletedStep_KeepsTieredFileActive_Postgres(t *testing.T) {
	dbCtx := testutil.NewPostgresDB(t, "kuranas_worker_it")
	truncateWorkerAndFiles(t, dbCtx)

	root := t.TempDir()
	coldDir := t.TempDir()
	filesSvc := files.NewService(files.NewRepository(dbCtx), jobs.NewRepository(dbCtx), nil)
	workerCtx := &WorkerContext{FilesService: filesSvc}

	// Tiered file: absent from the hot tree, bytes present in the cold area.
	tieredLogical := filepath.Join(root, "cold.txt")
	tieredPhysical := filepath.Join(coldDir, "cold.txt")
	if err := os.WriteFile(tieredPhysical, []byte("cold bytes"), 0o644); err != nil {
		t.Fatalf("write cold copy: %v", err)
	}
	insertHomeFile(t, dbCtx, "cold.txt", tieredLogical, root)
	setPhysicalPath(t, dbCtx, tieredLogical, tieredPhysical)

	// Tiered file whose cold copy was destroyed: really missing everywhere.
	goneLogical := filepath.Join(root, "gone.txt")
	gonePhysical := filepath.Join(coldDir, "gone.txt") // never created
	insertHomeFile(t, dbCtx, "gone.txt", goneLogical, root)
	setPhysicalPath(t, dbCtx, goneLogical, gonePhysical)

	payload, _ := marshalPayload(StepFilePayload{Path: root})
	if err := executeMarkDeletedStep(workerCtx, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("executeMarkDeletedStep returned error: %v", err)
	}

	if isMarkedDeleted(t, dbCtx, tieredLogical) {
		t.Fatalf("tiered file with live cold copy must NOT be marked deleted")
	}
	if !isMarkedDeleted(t, dbCtx, goneLogical) {
		t.Fatalf("tiered file missing from BOTH tiers should be marked deleted")
	}
}

// TestMarkDeletedStep_TieredFileSurvivesWatcherRemoveEvent_Postgres mirrors what
// the fsnotify watcher does when the migration job removes the hot copy: it
// enqueues a targeted mark_deleted for that exact logical path. The blindage
// must make that targeted check a no-op for a healthy tiered file.
func TestMarkDeletedStep_TieredFileSurvivesWatcherRemoveEvent_Postgres(t *testing.T) {
	dbCtx := testutil.NewPostgresDB(t, "kuranas_worker_it")
	truncateWorkerAndFiles(t, dbCtx)

	root := t.TempDir()
	coldDir := t.TempDir()
	filesSvc := files.NewService(files.NewRepository(dbCtx), jobs.NewRepository(dbCtx), nil)
	workerCtx := &WorkerContext{FilesService: filesSvc}

	tieredLogical := filepath.Join(root, "video.mp4")
	tieredPhysical := filepath.Join(coldDir, "video.mp4")
	if err := os.WriteFile(tieredPhysical, []byte("frames"), 0o644); err != nil {
		t.Fatalf("write cold copy: %v", err)
	}
	insertHomeFile(t, dbCtx, "video.mp4", tieredLogical, root)
	setPhysicalPath(t, dbCtx, tieredLogical, tieredPhysical)

	// The watcher payload targets the removed file's own path, not the root.
	payload, _ := marshalPayload(StepFilePayload{Path: tieredLogical})
	err := executeMarkDeletedStep(workerCtx, jobs.StepModel{Payload: payload})
	if err != nil && err != ErrStepSkipped {
		t.Fatalf("executeMarkDeletedStep returned unexpected error: %v", err)
	}

	if isMarkedDeleted(t, dbCtx, tieredLogical) {
		t.Fatalf("watcher remove event for a migrated file must not soft-delete the tiered row")
	}
}

// TestMarkDeletedStep_RestoresReappearedFile_Postgres covers the restore half
// of mark_deleted: a soft-deleted row whose file is back on disk must have its
// deleted_at cleared. This depends on the step querying with DeletedFilterAny —
// an only-active filter would hide the row and silently kill the restore flow.
func TestMarkDeletedStep_RestoresReappearedFile_Postgres(t *testing.T) {
	dbCtx := testutil.NewPostgresDB(t, "kuranas_worker_it")
	truncateWorkerAndFiles(t, dbCtx)

	root := t.TempDir()
	filesSvc := files.NewService(files.NewRepository(dbCtx), jobs.NewRepository(dbCtx), nil)
	workerCtx := &WorkerContext{FilesService: filesSvc}

	backPath := filepath.Join(root, "back.txt")
	if err := os.WriteFile(backPath, []byte("back"), 0o644); err != nil {
		t.Fatalf("write reappeared file: %v", err)
	}

	insertHomeFile(t, dbCtx, "back.txt", backPath, root)
	markErr := dbCtx.ExecTx(func(tx *sql.Tx) error {
		_, e := tx.Exec("UPDATE home_file SET deleted_at = now() WHERE path = $1", backPath)
		return e
	})
	if markErr != nil {
		t.Fatalf("soft-delete row: %v", markErr)
	}

	payload, _ := marshalPayload(StepFilePayload{Path: root})
	if err := executeMarkDeletedStep(workerCtx, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("executeMarkDeletedStep returned error: %v", err)
	}

	if isMarkedDeleted(t, dbCtx, backPath) {
		t.Fatalf("expected reappeared file to be restored (deleted_at cleared), but it is still soft-deleted")
	}
}
