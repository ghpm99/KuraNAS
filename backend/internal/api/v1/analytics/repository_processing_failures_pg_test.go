package analytics

import (
	"database/sql"
	"fmt"
	"slices"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

type stepFixture struct {
	jobID       int
	stepType    string
	status      string
	payload     string
	endedOffset time.Duration
}

var processingFailuresBaseTime = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

func newProcessingFailuresRepository(t *testing.T) *Repository {
	t.Helper()
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_analytics_failures_it"))
	err := repo.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("TRUNCATE home_file, worker_job RESTART IDENTITY CASCADE")
		return execErr
	})
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
	return repo
}

func insertProcessingFile(t *testing.T, repo *Repository, name string) int {
	t.Helper()
	var fileID int
	err := repo.DbContext.ExecTx(func(tx *sql.Tx) error {
		return tx.QueryRow(
			`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type)
			 VALUES ($1, $2, '/lib', '.jpg', 10, NOW(), NOW(), 2) RETURNING id`,
			name, "/lib/"+name,
		).Scan(&fileID)
	})
	if err != nil {
		t.Fatalf("insert file: %v", err)
	}
	return fileID
}

func insertProcessingJob(t *testing.T, repo *Repository, scope string) int {
	t.Helper()
	var jobID int
	err := repo.DbContext.ExecTx(func(tx *sql.Tx) error {
		var scopeValue any
		if scope != "" {
			scopeValue = scope
		}
		return tx.QueryRow(
			`INSERT INTO worker_job (type, priority, scope, status) VALUES ('startup_scan', 'normal', $1, 'completed') RETURNING id`,
			scopeValue,
		).Scan(&jobID)
	})
	if err != nil {
		t.Fatalf("insert job: %v", err)
	}
	return jobID
}

func insertProcessingStep(t *testing.T, repo *Repository, fixture stepFixture) {
	t.Helper()
	err := repo.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(
			`INSERT INTO worker_step (job_id, type, status, payload, last_error, created_at, ended_at)
			 VALUES ($1, $2, $3, $4, 'raw internal failure', $5, $5)`,
			fixture.jobID, fixture.stepType, fixture.status, fixture.payload, processingFailuresBaseTime.Add(fixture.endedOffset),
		)
		return execErr
	})
	if err != nil {
		t.Fatalf("insert step: %v", err)
	}
}

func filePayload(fileID int) string {
	return fmt.Sprintf(`{"file_id": %d}`, fileID)
}

func failureNames(failures []ProcessingFailureModel) []string {
	names := make([]string, 0, len(failures))
	for _, failure := range failures {
		names = append(names, fmt.Sprintf("%s:%s", failure.Name, failure.Stage))
	}
	return names
}

func seedProcessingFailures(t *testing.T, repo *Repository) {
	t.Helper()
	failedFile := insertProcessingFile(t, repo, "failed.jpg")
	retriedFile := insertProcessingFile(t, repo, "retried.jpg")
	otherStageFile := insertProcessingFile(t, repo, "other-stage.pdf")
	insertProcessingFile(t, repo, "by-path.jpg")

	jobID := insertProcessingJob(t, repo, "")
	insertProcessingStep(t, repo, stepFixture{jobID, "thumbnail", "failed", filePayload(failedFile), 1 * time.Minute})
	insertProcessingStep(t, repo, stepFixture{jobID, "thumbnail", "failed", filePayload(retriedFile), 2 * time.Minute})
	insertProcessingStep(t, repo, stepFixture{jobID, "thumbnail", "completed", filePayload(retriedFile), 3 * time.Minute})
	insertProcessingStep(t, repo, stepFixture{jobID, "document_text_index", "failed", filePayload(otherStageFile), 4 * time.Minute})
	insertProcessingStep(t, repo, stepFixture{jobID, "metadata", "failed", `{"path": "/lib/by-path.jpg"}`, 5 * time.Minute})
	insertProcessingStep(t, repo, stepFixture{jobID, "scan_filesystem", "failed", `{}`, 6 * time.Minute})
}

func TestPostgres_ProcessingFailuresLatestAttemptOrderedNewestFirst(t *testing.T) {
	repo := newProcessingFailuresRepository(t)
	seedProcessingFailures(t, repo)

	page, err := repo.GetProcessingFailuresPage("", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"by-path.jpg:metadata", "other-stage.pdf:document_text", "failed.jpg:thumbnail"}
	if got := failureNames(page.Items); !slices.Equal(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
	if page.Items[0].FileID == 0 || page.Items[0].RawError != "raw internal failure" || !page.Items[0].FailedAt.Equal(processingFailuresBaseTime.Add(5*time.Minute)) {
		t.Fatalf("unexpected first failure: %+v", page.Items[0])
	}
	if page.Pagination.HasNext || page.Pagination.HasPrev {
		t.Fatalf("unexpected pagination: %+v", page.Pagination)
	}
}

func TestPostgres_ProcessingFailuresFiltersByStage(t *testing.T) {
	repo := newProcessingFailuresRepository(t)
	seedProcessingFailures(t, repo)

	page, err := repo.GetProcessingFailuresPage(ProcessingStageThumbnail, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := failureNames(page.Items); !slices.Equal(got, []string{"failed.jpg:thumbnail"}) {
		t.Fatalf("expected only the unretried thumbnail failure, got %v", got)
	}
}

func TestPostgres_ProcessingFailuresPaginatesWithHasNext(t *testing.T) {
	repo := newProcessingFailuresRepository(t)
	seedProcessingFailures(t, repo)

	firstPage, err := repo.GetProcessingFailuresPage("", 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	secondPage, err := repo.GetProcessingFailuresPage("", 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := failureNames(firstPage.Items); !slices.Equal(got, []string{"by-path.jpg:metadata", "other-stage.pdf:document_text"}) || !firstPage.Pagination.HasNext {
		t.Fatalf("unexpected first page: %v %+v", got, firstPage.Pagination)
	}
	if got := failureNames(secondPage.Items); !slices.Equal(got, []string{"failed.jpg:thumbnail"}) || secondPage.Pagination.HasNext || !secondPage.Pagination.HasPrev {
		t.Fatalf("unexpected second page: %v %+v", got, secondPage.Pagination)
	}
}

func TestPostgres_ProcessingFailuresIgnoresDeletedFiles(t *testing.T) {
	repo := newProcessingFailuresRepository(t)
	seedProcessingFailures(t, repo)
	err := repo.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`UPDATE home_file SET deleted_at = NOW() WHERE name = 'failed.jpg'`)
		return execErr
	})
	if err != nil {
		t.Fatalf("delete file: %v", err)
	}

	page, err := repo.GetProcessingFailuresPage("", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := failureNames(page.Items); slices.Contains(got, "failed.jpg:thumbnail") {
		t.Fatalf("deleted file must not appear, got %v", got)
	}
}
