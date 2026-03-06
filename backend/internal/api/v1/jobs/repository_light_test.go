package jobs

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/jobs"

	"github.com/DATA-DOG/go-sqlmock"
)

func newJobsRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestJobsRepositoryCreateAndUpdate(t *testing.T) {
	repo, mock, db := newJobsRepoWithMock(t)
	defer db.Close()

	now := time.Now()

	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context")
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.InsertJobQuery)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.InsertStepQuery)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateJobStatusQuery)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateStepStatusQuery)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateStepExecutionQuery)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := repo.CreateJob(tx, JobModel{
			ID:              "job-1",
			Type:            "startup_scan",
			Priority:        2,
			ScopeJSON:       "{}",
			Status:          "queued",
			CancelRequested: false,
			LastError:       "",
		})
		if err != nil {
			return err
		}

		_, err = repo.CreateStep(tx, StepModel{
			ID:            "step-1",
			JobID:         "job-1",
			Type:          "scan_filesystem",
			Status:        "queued",
			DependsOnJSON: "[]",
			Attempts:      0,
			MaxAttempts:   3,
			LastError:     "",
			Progress:      0,
			PayloadJSON:   "{}",
		})
		if err != nil {
			return err
		}

		ok, err := repo.UpdateJobStatus(tx, "job-1", "queued", "running", &now, nil, "")
		if err != nil {
			return err
		}
		if !ok {
			t.Fatalf("expected UpdateJobStatus to update one row")
		}

		ok, err = repo.UpdateStepStatus(tx, "step-1", "queued", "running", &now, nil, "")
		if err != nil {
			return err
		}
		if !ok {
			t.Fatalf("expected UpdateStepStatus to update one row")
		}

		ok, err = repo.UpdateStepExecution(tx, "step-1", 1, "", 20, &now, nil)
		if err != nil {
			return err
		}
		if !ok {
			t.Fatalf("expected UpdateStepExecution to update one row")
		}

		return nil
	})
	if err != nil {
		t.Fatalf("ExecTx failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestJobsRepositoryReadPaths(t *testing.T) {
	repo, mock, db := newJobsRepoWithMock(t)
	defer db.Close()

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetJobByIDQuery)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "type", "priority", "parent_job_id", "scope_json", "status", "created_at", "started_at", "ended_at", "cancel_requested", "last_error",
		}).AddRow("job-1", "startup_scan", 2, nil, "{}", "queued", now, nil, nil, false, ""))
	mock.ExpectRollback()

	job, err := repo.GetJobByID("job-1")
	if err != nil {
		t.Fatalf("GetJobByID failed: %v", err)
	}
	if job.ID != "job-1" {
		t.Fatalf("unexpected job id: %s", job.ID)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListJobsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "type", "priority", "parent_job_id", "scope_json", "status", "created_at", "started_at", "ended_at", "cancel_requested", "last_error",
		}).AddRow("job-1", "startup_scan", 2, nil, "{}", "queued", now, nil, nil, false, ""))
	mock.ExpectRollback()

	list, err := repo.ListJobs(JobFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("unexpected items count: %d", len(list.Items))
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetStepsByJobIDQuery)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "job_id", "type", "status", "depends_on_json", "attempts", "max_attempts", "last_error", "progress", "payload_json", "created_at", "started_at", "ended_at",
		}).AddRow("step-1", "job-1", "scan_filesystem", "queued", "[]", 0, 3, "", 0, "{}", now, nil, nil))
	mock.ExpectRollback()

	steps, err := repo.GetStepsByJobID("job-1")
	if err != nil {
		t.Fatalf("GetStepsByJobID failed: %v", err)
	}
	if len(steps) != 1 || steps[0].ID != "step-1" {
		t.Fatalf("unexpected steps result: %+v", steps)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestJobsRepositoryErrorPaths(t *testing.T) {
	repo, mock, db := newJobsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.InsertJobQuery)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := repo.CreateJob(tx, JobModel{ID: "job-1"})
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "CreateJob") {
		t.Fatalf("expected wrapped CreateJob error, got %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListJobsQuery)).
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()

	_, err = repo.ListJobs(JobFilter{}, 1, 10)
	if err == nil || !strings.Contains(err.Error(), "ListJobs") {
		t.Fatalf("expected wrapped ListJobs error, got %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateStepExecutionQuery)).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected failed")))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback()

	_, err = repo.UpdateStepExecution(tx, "step-1", 1, "", 10, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "rows affected") {
		t.Fatalf("expected rows affected error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
