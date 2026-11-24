package ingest

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
)

func TestDownloadRepositoryCreateQueuedDownload_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_ingest_download_it")
	repository := NewDownloadRepository(dbContext)

	var jobID int
	err := dbContext.GetDatabase().QueryRow(
		`INSERT INTO worker_job (type, priority, status) VALUES ('remote_fetch', 'normal', 'queued') RETURNING id`,
	).Scan(&jobID)
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}

	var downloadID int
	err = dbContext.ExecTx(func(tx *sql.Tx) error {
		createdID, createErr := repository.CreateQueuedDownload(tx, QueuedDownload{
			JobID:          jobID,
			URL:            "https://x.test/v",
			Preset:         "video_1080",
			TargetRootID:   2,
			TargetRootName: "Videos",
			Subfolder:      "Shows",
		})
		downloadID = createdID
		return createErr
	})
	if err != nil {
		t.Fatalf("CreateQueuedDownload: %v", err)
	}

	var (
		storedJobID  sql.NullInt64
		url, preset  string
		rootID       int
		rootName     string
		subfolder    string
		state        string
		progress     int
		finalPath    string
		failureText  string
		title        sql.NullString
		endedAt      sql.NullTime
		createdAtSet bool
	)
	err = dbContext.GetDatabase().QueryRow(
		`SELECT job_id, url, preset, target_root_id, target_root_name, subfolder, state, progress, final_path, error, title, ended_at, created_at IS NOT NULL
		 FROM ingest_download WHERE id = $1`, downloadID,
	).Scan(&storedJobID, &url, &preset, &rootID, &rootName, &subfolder, &state, &progress, &finalPath, &failureText, &title, &endedAt, &createdAtSet)
	if err != nil {
		t.Fatalf("load download: %v", err)
	}
	if !storedJobID.Valid || int(storedJobID.Int64) != jobID {
		t.Fatalf("job_id = %+v, want %d", storedJobID, jobID)
	}
	if url != "https://x.test/v" || preset != "video_1080" || rootID != 2 || rootName != "Videos" || subfolder != "Shows" {
		t.Fatalf("unexpected stored request: %s %s %d %s %s", url, preset, rootID, rootName, subfolder)
	}
	if state != "queued" || progress != 0 || finalPath != "" || failureText != "" || title.Valid || endedAt.Valid || !createdAtSet {
		t.Fatalf("unexpected defaults: state=%s progress=%d finalPath=%q error=%q title=%v endedAt=%v", state, progress, finalPath, failureText, title, endedAt)
	}

	if _, err := dbContext.GetDatabase().Exec(`DELETE FROM worker_job WHERE id = $1`, jobID); err != nil {
		t.Fatalf("delete job: %v", err)
	}
	err = dbContext.GetDatabase().QueryRow(`SELECT job_id FROM ingest_download WHERE id = $1`, downloadID).Scan(&storedJobID)
	if err != nil || storedJobID.Valid {
		t.Fatalf("job_id after job delete = %+v err=%v, want NULL", storedJobID, err)
	}
}

type storedDownloadState struct {
	state     string
	progress  int
	speed     sql.NullInt64
	eta       sql.NullInt64
	title     sql.NullString
	finalPath string
	failure   string
	endedAt   sql.NullTime
}

func seedQueuedDownload(t *testing.T, dbContext *database.DbContext, repository *DownloadRepository) int {
	t.Helper()
	var jobID int
	err := dbContext.GetDatabase().QueryRow(
		`INSERT INTO worker_job (type, priority, status) VALUES ('remote_fetch', 'normal', 'running') RETURNING id`,
	).Scan(&jobID)
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}
	var downloadID int
	err = dbContext.ExecTx(func(tx *sql.Tx) error {
		createdID, createErr := repository.CreateQueuedDownload(tx, QueuedDownload{JobID: jobID, URL: "https://x.test/v", Preset: "video_1080", TargetRootID: 1, TargetRootName: "Videos"})
		downloadID = createdID
		return createErr
	})
	if err != nil {
		t.Fatalf("seed download: %v", err)
	}
	return downloadID
}

func loadStoredDownloadState(t *testing.T, dbContext *database.DbContext, downloadID int) storedDownloadState {
	t.Helper()
	var stored storedDownloadState
	err := dbContext.GetDatabase().QueryRow(
		`SELECT state, progress, speed_bytes_per_second, eta_seconds, title, final_path, error, ended_at FROM ingest_download WHERE id = $1`, downloadID,
	).Scan(&stored.state, &stored.progress, &stored.speed, &stored.eta, &stored.title, &stored.finalPath, &stored.failure, &stored.endedAt)
	if err != nil {
		t.Fatalf("load download: %v", err)
	}
	return stored
}

func TestDownloadRepositoryLifecycleUpdates_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_ingest_download_lifecycle_it")
	repository := NewDownloadRepository(dbContext)

	completedID := seedQueuedDownload(t, dbContext, repository)
	if err := repository.UpdateProgress(completedID, DownloadProgress{Percent: 10}); err != nil {
		t.Fatalf("UpdateProgress while queued: %v", err)
	}
	if stored := loadStoredDownloadState(t, dbContext, completedID); stored.state != "queued" || stored.progress != 0 {
		t.Fatalf("progress must not touch a queued download: %+v", stored)
	}

	if err := repository.MarkRunning(completedID); err != nil {
		t.Fatalf("MarkRunning: %v", err)
	}
	speed := int64(1289748)
	eta := 12
	if err := repository.UpdateProgress(completedID, DownloadProgress{Percent: 42, SpeedBytesPerSecond: &speed, ETASeconds: &eta}); err != nil {
		t.Fatalf("UpdateProgress: %v", err)
	}
	if err := repository.UpdateTitle(completedID, "clip"); err != nil {
		t.Fatalf("UpdateTitle: %v", err)
	}
	running := loadStoredDownloadState(t, dbContext, completedID)
	if running.state != "running" || running.progress != 42 || running.speed.Int64 != speed || running.eta.Int64 != int64(eta) || running.title.String != "clip" || running.endedAt.Valid {
		t.Fatalf("unexpected running state: %+v", running)
	}

	if err := repository.MarkCompleted(completedID, "clip", "/library/clip.mp4"); err != nil {
		t.Fatalf("MarkCompleted: %v", err)
	}
	completed := loadStoredDownloadState(t, dbContext, completedID)
	if completed.state != "completed" || completed.progress != 100 || completed.speed.Valid || completed.eta.Valid || completed.finalPath != "/library/clip.mp4" || !completed.endedAt.Valid {
		t.Fatalf("unexpected completed state: %+v", completed)
	}
	if err := repository.MarkFailed(completedID, "late"); err != nil {
		t.Fatalf("MarkFailed after completion: %v", err)
	}
	if stored := loadStoredDownloadState(t, dbContext, completedID); stored.state != "completed" || stored.failure != "" {
		t.Fatalf("a finished download must stay as it ended: %+v", stored)
	}

	failedID := seedQueuedDownload(t, dbContext, repository)
	if err := repository.MarkRunning(failedID); err != nil {
		t.Fatalf("MarkRunning: %v", err)
	}
	if err := repository.MarkFailed(failedID, "translated reason"); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}
	failed := loadStoredDownloadState(t, dbContext, failedID)
	if failed.state != "failed" || failed.failure != "translated reason" || !failed.endedAt.Valid {
		t.Fatalf("unexpected failed state: %+v", failed)
	}

	canceledID := seedQueuedDownload(t, dbContext, repository)
	if err := repository.MarkRunning(canceledID); err != nil {
		t.Fatalf("MarkRunning: %v", err)
	}
	if err := repository.MarkCanceled(canceledID); err != nil {
		t.Fatalf("MarkCanceled: %v", err)
	}
	if canceled := loadStoredDownloadState(t, dbContext, canceledID); canceled.state != "canceled" || !canceled.endedAt.Valid {
		t.Fatalf("unexpected canceled state: %+v", canceled)
	}
}
