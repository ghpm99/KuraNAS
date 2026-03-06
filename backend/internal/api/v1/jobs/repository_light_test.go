package jobs

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/jobs"

	"github.com/DATA-DOG/go-sqlmock"
)

func newJobRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestJobRepositoryReadPaths(t *testing.T) {
	repo, mock, db := newJobRepoWithMock(t)
	defer db.Close()

	now := time.Now()

	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetJobByIDQuery)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "priority", "scope", "status", "created_at", "started_at", "ended_at", "cancel_requested", "last_error"}).
			AddRow(10, "startup_scan", "low", []byte(`{"root":"/mnt/data"}`), "queued", now, nil, nil, false, ""))
	mock.ExpectRollback()

	job, err := repo.GetJobByID(10)
	if err != nil {
		t.Fatalf("GetJobByID returned error: %v", err)
	}
	if job.ID != 10 {
		t.Fatalf("expected job id 10, got %d", job.ID)
	}

	filter := JobFilter{}
	filter.Status.Set("running")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListJobsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "priority", "scope", "status", "created_at", "started_at", "ended_at", "cancel_requested", "last_error"}).
			AddRow(11, "upload_process", "high", []byte(`{"path":"/x"}`), "running", now, now, nil, false, "").
			AddRow(12, "fs_event", "normal", []byte(`{"path":"/y"}`), "queued", now, nil, nil, false, ""))
	mock.ExpectRollback()

	jobs, err := repo.ListJobs(filter, 1, 10)
	if err != nil {
		t.Fatalf("ListJobs returned error: %v", err)
	}
	if len(jobs.Items) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs.Items))
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetStepsByJobIDQuery)).
		WithArgs(11).
		WillReturnRows(sqlmock.NewRows([]string{"id", "job_id", "type", "status", "depends_on", "attempts", "max_attempts", "last_error", "progress", "payload", "created_at", "started_at", "ended_at"}).
			AddRow(1, 11, "checksum", "queued", []byte(`[2]`), 0, 3, "", 0, []byte(`{"file_id":9}`), now, nil, nil))
	mock.ExpectRollback()

	steps, err := repo.GetStepsByJobID(11)
	if err != nil {
		t.Fatalf("GetStepsByJobID returned error: %v", err)
	}
	if len(steps) != 1 || steps[0].Type != "checksum" {
		t.Fatalf("unexpected steps result: %+v", steps)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestJobRepositoryWritePaths(t *testing.T) {
	repo, mock, db := newJobRepoWithMock(t)
	defer db.Close()

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertJobQuery)).
		WithArgs("startup_scan", "low", []byte(`{"root":"/mnt/data"}`), "queued", false, "").
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "priority", "scope", "status", "created_at", "started_at", "ended_at", "cancel_requested", "last_error"}).
			AddRow(100, "startup_scan", "low", []byte(`{"root":"/mnt/data"}`), "queued", now, nil, nil, false, ""))

	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertStepQuery)).
		WithArgs(100, "scan_filesystem", "queued", []byte(`[]`), 0, 3, "", 0, []byte(`{"root":"/mnt/data"}`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "job_id", "type", "status", "depends_on", "attempts", "max_attempts", "last_error", "progress", "payload", "created_at", "started_at", "ended_at"}).
			AddRow(1000, 100, "scan_filesystem", "queued", []byte(`[]`), 0, 3, "", 0, []byte(`{"root":"/mnt/data"}`), now, nil, nil))

	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateJobExecutionQuery)).
		WithArgs("running", sqlmock.AnyArg(), nil, nil, nil, 100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateStepExecutionQuery)).
		WithArgs("running", 25, 1, sqlmock.AnyArg(), nil, nil, 1000).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		job, createJobErr := repo.CreateJob(tx, JobModel{
			Type:            "startup_scan",
			Priority:        "low",
			Scope:           []byte(`{"root":"/mnt/data"}`),
			Status:          "queued",
			CancelRequested: false,
			LastError:       "",
		})
		if createJobErr != nil {
			return createJobErr
		}

		step, createStepErr := repo.CreateStep(tx, StepModel{
			JobID:       job.ID,
			Type:        "scan_filesystem",
			Status:      "queued",
			DependsOn:   []byte(`[]`),
			Attempts:    0,
			MaxAttempts: 3,
			LastError:   "",
			Progress:    0,
			Payload:     []byte(`{"root":"/mnt/data"}`),
		})
		if createStepErr != nil {
			return createStepErr
		}

		started := time.Now()
		updatedJob, updateJobErr := repo.UpdateJobExecution(tx, job.ID, "running", &started, nil, nil, nil)
		if updateJobErr != nil {
			return updateJobErr
		}
		if !updatedJob {
			t.Fatalf("expected job update to affect one row")
		}

		updatedStep, updateStepErr := repo.UpdateStepExecution(tx, step.ID, "running", 25, 1, &started, nil, nil)
		if updateStepErr != nil {
			return updateStepErr
		}
		if !updatedStep {
			t.Fatalf("expected step update to affect one row")
		}

		return nil
	})
	if err != nil {
		t.Fatalf("write transaction failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestJobRepositoryUpdateErrors(t *testing.T) {
	repo, mock, db := newJobRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateJobExecutionQuery)).
		WithArgs("failed", nil, nil, nil, sqlmock.AnyArg(), 1).
		WillReturnError(errors.New("job update failed"))
	mock.ExpectRollback()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		lastError := "boom"
		_, callErr := repo.UpdateJobExecution(tx, 1, "failed", nil, nil, nil, &lastError)
		return callErr
	})
	if err == nil {
		t.Fatalf("expected update job error")
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateStepExecutionQuery)).
		WithArgs("failed", 100, 3, nil, nil, sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectRollback()

	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		lastError := "step boom"
		_, callErr := repo.UpdateStepExecution(tx, 2, "failed", 100, 3, nil, nil, &lastError)
		return callErr
	})
	if err == nil {
		t.Fatalf("expected update step error for multiple rows")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
