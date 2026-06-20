package image

import (
	"database/sql"
	"errors"
	"testing"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
)

type fakeImageRepo struct {
	RepositoryInterface
	count    int
	countErr error
}

func (f *fakeImageRepo) CountPendingAIClassification(threshold float64) (int, error) {
	return f.count, f.countErr
}

type fakeJobEnqueuer struct {
	dbCtx         *database.DbContext
	activeJobs    []jobs.JobModel
	listErr       error
	createJobErr  error
	createStepErr error
	createdJobID  int
	stepCreated   bool
}

func (f *fakeJobEnqueuer) GetDbContext() *database.DbContext { return f.dbCtx }

func (f *fakeJobEnqueuer) CreateJob(tx *sql.Tx, j jobs.JobModel) (jobs.JobModel, error) {
	if f.createJobErr != nil {
		return jobs.JobModel{}, f.createJobErr
	}
	j.ID = 99
	f.createdJobID = j.ID
	return j, nil
}

func (f *fakeJobEnqueuer) CreateStep(tx *sql.Tx, s jobs.StepModel) (jobs.StepModel, error) {
	if f.createStepErr != nil {
		return jobs.StepModel{}, f.createStepErr
	}
	f.stepCreated = true
	return s, nil
}

func (f *fakeJobEnqueuer) ListJobs(filter jobs.JobFilter, page int, pageSize int) (utils.PaginationResponse[jobs.JobModel], error) {
	if f.listErr != nil {
		return utils.PaginationResponse[jobs.JobModel]{}, f.listErr
	}
	return utils.PaginationResponse[jobs.JobModel]{Items: f.activeJobs}, nil
}

func newSqlmockDbContext(t *testing.T) (*database.DbContext, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	return database.NewDbContext(db), mock, db
}

func TestGetPendingAIClassificationCount(t *testing.T) {
	svc := &Service{Repository: &fakeImageRepo{count: 5}}
	count, err := svc.GetPendingAIClassificationCount()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected 5, got %d", count)
	}
}

func TestGetPendingAIClassificationCount_Error(t *testing.T) {
	svc := &Service{Repository: &fakeImageRepo{countErr: errors.New("boom")}}
	if _, err := svc.GetPendingAIClassificationCount(); err == nil {
		t.Fatal("expected error")
	}
}

func TestEnqueueClassificationBackfill_Unavailable(t *testing.T) {
	svc := &Service{Repository: &fakeImageRepo{}}
	if _, err := svc.EnqueueClassificationBackfill(); !errors.Is(err, ErrBackfillUnavailable) {
		t.Fatalf("expected ErrBackfillUnavailable, got %v", err)
	}
}

func TestEnqueueClassificationBackfill_AlreadyActive(t *testing.T) {
	enq := &fakeJobEnqueuer{activeJobs: []jobs.JobModel{{ID: 42}}}
	svc := &Service{Repository: &fakeImageRepo{}, JobEnqueuer: enq}

	id, err := svc.EnqueueClassificationBackfill()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected existing job id 42, got %d", id)
	}
	if enq.stepCreated {
		t.Fatal("should not create a new job when one is already active")
	}
}

func TestEnqueueClassificationBackfill_Success(t *testing.T) {
	dbCtx, mock, db := newSqlmockDbContext(t)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectCommit()

	enq := &fakeJobEnqueuer{dbCtx: dbCtx}
	svc := &Service{Repository: &fakeImageRepo{}, JobEnqueuer: enq}

	id, err := svc.EnqueueClassificationBackfill()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 99 {
		t.Fatalf("expected new job id 99, got %d", id)
	}
	if !enq.stepCreated {
		t.Fatal("expected enumerate step to be created")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestEnqueueClassificationBackfill_ListError(t *testing.T) {
	enq := &fakeJobEnqueuer{listErr: errors.New("boom")}
	svc := &Service{Repository: &fakeImageRepo{}, JobEnqueuer: enq}
	if _, err := svc.EnqueueClassificationBackfill(); err == nil {
		t.Fatal("expected error")
	}
}

func TestEnqueueClassificationBackfill_CreateError(t *testing.T) {
	dbCtx, mock, db := newSqlmockDbContext(t)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()

	enq := &fakeJobEnqueuer{dbCtx: dbCtx, createJobErr: errors.New("boom")}
	svc := &Service{Repository: &fakeImageRepo{}, JobEnqueuer: enq}
	if _, err := svc.EnqueueClassificationBackfill(); err == nil {
		t.Fatal("expected error")
	}
}
