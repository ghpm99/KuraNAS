package ingest

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/roots"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
)

func newRetryService(t *testing.T, dbContext *database.DbContext) *DownloadsService {
	t.Helper()
	roots.Set([]roots.Root{{ID: 2, Path: filepath.Clean(t.TempDir()), Label: "Videos", Enabled: true}})
	t.Cleanup(roots.Reset)
	downloadRepository := NewDownloadRepository(dbContext)
	fetchService := NewService(jobs.NewRepository(dbContext), downloadRepository, installedProbe{})
	return NewDownloadsService(downloadRepository, fetchService)
}

func countRows(t *testing.T, dbContext *database.DbContext, table string) int {
	t.Helper()
	var total int
	if err := dbContext.GetDatabase().QueryRow(`SELECT count(*) FROM ` + table).Scan(&total); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return total
}

func TestRetryDownload_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_ingest_download_retry_it")
	service := newRetryService(t, dbContext)
	clearDownloadsAndFiles(t, dbContext)
	now := time.Now().UTC().Truncate(time.Second)

	failedID := seedDownloadRow(t, dbContext, seededDownload{state: "failed", createdAt: now})
	canceledID := seedDownloadRow(t, dbContext, seededDownload{state: "canceled", createdAt: now})
	completedID := seedDownloadRow(t, dbContext, seededDownload{state: "completed", createdAt: now})
	runningID := seedDownloadRow(t, dbContext, seededDownload{state: "running", createdAt: now})
	queuedID := seedDownloadRow(t, dbContext, seededDownload{state: "queued", createdAt: now})

	t.Run("failed creates a new queued record with the same data and a job", func(t *testing.T) {
		jobsBefore := countRows(t, dbContext, "worker_job")
		retried, err := service.RetryDownload(failedID)
		if err != nil {
			t.Fatalf("RetryDownload: %v", err)
		}
		if retried.ID == failedID || retried.State != "queued" || retried.Progress != 0 {
			t.Fatalf("unexpected new record: %+v", retried)
		}
		if retried.URL != "https://x.test/v" || retried.Preset != "video_1080" ||
			retried.TargetRootID != 2 || retried.TargetRootName != "Videos" || retried.Subfolder != "Shows" {
			t.Fatalf("data not copied: %+v", retried)
		}
		if retried.JobID == nil || retried.Error != "" || retried.EndedAt != nil {
			t.Fatalf("expected a job and clean record: %+v", retried)
		}
		if countRows(t, dbContext, "worker_job") != jobsBefore+1 {
			t.Fatalf("expected exactly one new job")
		}
		original, err := NewDownloadRepository(dbContext).GetDownload(failedID)
		if err != nil || original.State != "failed" {
			t.Fatalf("original changed: %+v err=%v", original, err)
		}
	})

	t.Run("canceled can be retried", func(t *testing.T) {
		if _, err := service.RetryDownload(canceledID); err != nil {
			t.Fatalf("RetryDownload: %v", err)
		}
	})

	for name, downloadID := range map[string]int{"completed": completedID, "running": runningID, "queued": queuedID} {
		t.Run(name+" is a conflict", func(t *testing.T) {
			downloadsBefore := countRows(t, dbContext, "ingest_download")
			if _, err := service.RetryDownload(downloadID); !errors.Is(err, ErrDownloadNotRetryable) {
				t.Fatalf("err = %v, want ErrDownloadNotRetryable", err)
			}
			if countRows(t, dbContext, "ingest_download") != downloadsBefore {
				t.Fatalf("a record was created")
			}
		})
	}

	t.Run("missing is not found", func(t *testing.T) {
		if _, err := service.RetryDownload(987654); !errors.Is(err, ErrDownloadNotFound) {
			t.Fatalf("err = %v, want ErrDownloadNotFound", err)
		}
	})

	t.Run("destination root no longer enabled", func(t *testing.T) {
		roots.Set([]roots.Root{{ID: 9, Path: filepath.Clean(t.TempDir()), Label: "Other", Enabled: true}})
		if _, err := service.RetryDownload(failedID); !errors.Is(err, ErrInvalidTarget) {
			t.Fatalf("err = %v, want ErrInvalidTarget", err)
		}
	})
}
