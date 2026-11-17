package takeout

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

type storedImport struct {
	state            string
	fileName         string
	size             int64
	totalEntries     int
	processedEntries int
	imported         int
	alreadyExisting  int
	ignored          int
	failedEntries    int
	failureMessage   string
	endedAt          sql.NullTime
}

func createJobRow(t *testing.T, repository *ImportRepository) int {
	t.Helper()
	var jobID int
	err := repository.DbContext.GetDatabase().QueryRow(
		`INSERT INTO worker_job (type, priority, status) VALUES ('takeout_import', 'high', 'queued') RETURNING id`,
	).Scan(&jobID)
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}
	return jobID
}

func createQueuedImport(t *testing.T, repository *ImportRepository, jobID int) int {
	t.Helper()
	var importID int
	err := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		createdID, createErr := repository.CreateQueuedImport(tx, jobID, "takeout.zip", 2048)
		importID = createdID
		return createErr
	})
	if err != nil {
		t.Fatalf("CreateQueuedImport: %v", err)
	}
	return importID
}

func loadStoredImport(t *testing.T, repository *ImportRepository, importID int) storedImport {
	t.Helper()
	var stored storedImport
	err := repository.DbContext.GetDatabase().QueryRow(
		`SELECT state, file_name, size, total_entries, processed_entries, imported, already_existing, ignored, failed_entries, error, ended_at
		 FROM takeout_import WHERE id = $1`, importID,
	).Scan(&stored.state, &stored.fileName, &stored.size, &stored.totalEntries, &stored.processedEntries,
		&stored.imported, &stored.alreadyExisting, &stored.ignored, &stored.failedEntries, &stored.failureMessage,
		&stored.endedAt)
	if err != nil {
		t.Fatalf("load import: %v", err)
	}
	return stored
}

func TestImportRepositoryLifecycle_Postgres(t *testing.T) {
	repository := NewImportRepository(testutil.NewPostgresDB(t, "kuranas_takeout_import_it"))
	jobID := createJobRow(t, repository)
	importID := createQueuedImport(t, repository, jobID)

	queued := loadStoredImport(t, repository, importID)
	if queued.state != "queued" || queued.fileName != "takeout.zip" || queued.size != 2048 || queued.endedAt.Valid {
		t.Fatalf("unexpected queued import: %+v", queued)
	}

	if err := repository.MarkImportRunning(jobID); err != nil {
		t.Fatalf("MarkImportRunning: %v", err)
	}
	if running := loadStoredImport(t, repository, importID); running.state != "running" || running.endedAt.Valid {
		t.Fatalf("unexpected running import: %+v", running)
	}

	summary := ImportSummary{TotalEntries: 10, Imported: 6, AlreadyExisting: 2, Ignored: 1, FailedEntries: 1}
	if err := repository.CloseImportCompleted(jobID, summary); err != nil {
		t.Fatalf("CloseImportCompleted: %v", err)
	}
	completed := loadStoredImport(t, repository, importID)
	if completed.state != "completed" || !completed.endedAt.Valid || completed.failureMessage != "" {
		t.Fatalf("unexpected completed import: %+v", completed)
	}
	if completed.totalEntries != 10 || completed.processedEntries != 10 || completed.imported != 6 ||
		completed.alreadyExisting != 2 || completed.ignored != 1 || completed.failedEntries != 1 {
		t.Fatalf("unexpected counts: %+v", completed)
	}
}

func TestImportRepositoryCloseFailed_Postgres(t *testing.T) {
	repository := NewImportRepository(testutil.NewPostgresDB(t, "kuranas_takeout_import_it"))
	jobID := createJobRow(t, repository)
	importID := createQueuedImport(t, repository, jobID)

	if err := repository.CloseImportFailed(jobID, "arquivo invalido"); err != nil {
		t.Fatalf("CloseImportFailed: %v", err)
	}
	failed := loadStoredImport(t, repository, importID)
	if failed.state != "failed" || failed.failureMessage != "arquivo invalido" || !failed.endedAt.Valid {
		t.Fatalf("unexpected failed import: %+v", failed)
	}
}

func TestImportStateConstraintRejectsUnknownState_Postgres(t *testing.T) {
	repository := NewImportRepository(testutil.NewPostgresDB(t, "kuranas_takeout_import_it"))
	_, err := repository.DbContext.GetDatabase().Exec(`INSERT INTO takeout_import (file_name, state) VALUES ('a.zip', 'bogus')`)
	if err == nil {
		t.Fatalf("expected check constraint violation")
	}
}
