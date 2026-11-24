package ingest

import (
	"database/sql"
	"testing"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
)

type seededInterruptedDownload struct {
	downloadID int
	jobID      int
	stepID     int
}

func seedDownloadWithJob(t *testing.T, dbContext *database.DbContext, repository *DownloadRepository, jobStatus string, downloadState string) seededInterruptedDownload {
	t.Helper()
	var seeded seededInterruptedDownload
	err := dbContext.GetDatabase().QueryRow(
		`INSERT INTO worker_job (type, priority, status) VALUES ('remote_fetch', 'normal', $1) RETURNING id`, jobStatus,
	).Scan(&seeded.jobID)
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}
	err = dbContext.GetDatabase().QueryRow(
		`INSERT INTO worker_step (job_id, type, status) VALUES ($1, 'remote_fetch', $2) RETURNING id`, seeded.jobID, jobStatus,
	).Scan(&seeded.stepID)
	if err != nil {
		t.Fatalf("seed step: %v", err)
	}
	err = dbContext.ExecTx(func(tx *sql.Tx) error {
		createdID, createErr := repository.CreateQueuedDownload(tx, QueuedDownload{JobID: seeded.jobID, URL: "https://x.test/v", Preset: "video_1080", TargetRootID: 1, TargetRootName: "Videos"})
		seeded.downloadID = createdID
		return createErr
	})
	if err != nil {
		t.Fatalf("seed download: %v", err)
	}
	if _, err := dbContext.GetDatabase().Exec(`UPDATE ingest_download SET state = $2 WHERE id = $1`, seeded.downloadID, downloadState); err != nil {
		t.Fatalf("set download state: %v", err)
	}
	return seeded
}

func loadJobAndStepStatus(t *testing.T, dbContext *database.DbContext, seeded seededInterruptedDownload) (string, string, string) {
	t.Helper()
	var jobStatus, stepStatus, jobError string
	err := dbContext.GetDatabase().QueryRow(
		`SELECT j.status, s.status, COALESCE(j.last_error, '') FROM worker_job j JOIN worker_step s ON s.job_id = j.id WHERE j.id = $1 AND s.id = $2`,
		seeded.jobID, seeded.stepID,
	).Scan(&jobStatus, &stepStatus, &jobError)
	if err != nil {
		t.Fatalf("load job and step: %v", err)
	}
	return jobStatus, stepStatus, jobError
}

func TestDownloadRepositoryFailInterruptedDownloads_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_ingest_download_interrupted_it")
	repository := NewDownloadRepository(dbContext)
	const restartMessage = "interrompido pelo reinício do servidor"

	queued := seedDownloadWithJob(t, dbContext, repository, "queued", "queued")
	running := seedDownloadWithJob(t, dbContext, repository, "running", "running")
	completed := seedDownloadWithJob(t, dbContext, repository, "completed", "completed")
	canceled := seedDownloadWithJob(t, dbContext, repository, "canceled", "canceled")

	interruptedCount, err := repository.FailInterruptedDownloads(restartMessage)
	if err != nil {
		t.Fatalf("FailInterruptedDownloads: %v", err)
	}
	if interruptedCount != 2 {
		t.Fatalf("interruptedCount = %d, want 2", interruptedCount)
	}

	for name, seeded := range map[string]seededInterruptedDownload{"queued": queued, "running": running} {
		stored := loadStoredDownloadState(t, dbContext, seeded.downloadID)
		if stored.state != "failed" || stored.failure != restartMessage || !stored.endedAt.Valid {
			t.Fatalf("%s download = %+v, want failed with message and ended_at", name, stored)
		}
		jobStatus, stepStatus, jobError := loadJobAndStepStatus(t, dbContext, seeded)
		if jobStatus != "failed" || stepStatus != "failed" || jobError != restartMessage {
			t.Fatalf("%s job/step = %s/%s/%q, want failed", name, jobStatus, stepStatus, jobError)
		}
	}

	for name, untouched := range map[string]seededInterruptedDownload{"completed": completed, "canceled": canceled} {
		stored := loadStoredDownloadState(t, dbContext, untouched.downloadID)
		if stored.state != name || stored.failure != "" || stored.endedAt.Valid {
			t.Fatalf("%s download changed: %+v", name, stored)
		}
		jobStatus, stepStatus, _ := loadJobAndStepStatus(t, dbContext, untouched)
		if jobStatus != name || stepStatus != name {
			t.Fatalf("%s job/step changed: %s/%s", name, jobStatus, stepStatus)
		}
	}
}

func TestInterruptedDownloadJobsAreNotRecoveredByEngine_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_ingest_download_not_resumed_it")
	repository := NewDownloadRepository(dbContext)
	running := seedDownloadWithJob(t, dbContext, repository, "running", "running")

	if _, err := repository.FailInterruptedDownloads("reinício"); err != nil {
		t.Fatalf("FailInterruptedDownloads: %v", err)
	}

	jobsRepository := jobs.NewRepository(dbContext)
	var jobsReset, stepsReset int64
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		var recoverErr error
		jobsReset, stepsReset, recoverErr = jobsRepository.RecoverInterruptedWork(tx)
		return recoverErr
	})
	if err != nil {
		t.Fatalf("RecoverInterruptedWork: %v", err)
	}
	if jobsReset != 0 || stepsReset != 0 {
		t.Fatalf("engine recovery requeued jobs=%d steps=%d, want 0", jobsReset, stepsReset)
	}
	jobStatus, stepStatus, _ := loadJobAndStepStatus(t, dbContext, running)
	if jobStatus != "failed" || stepStatus != "failed" {
		t.Fatalf("job/step after recovery = %s/%s, want failed", jobStatus, stepStatus)
	}
}
