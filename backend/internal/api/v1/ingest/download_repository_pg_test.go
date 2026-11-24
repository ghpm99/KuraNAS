package ingest

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
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
