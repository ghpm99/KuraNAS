package jobs

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
)

type jobsServiceRepoMock struct {
	RepositoryInterface
	getJobByIDFn    func(id int) (JobModel, error)
	listJobsFn      func(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error)
	getStepsByJobFn func(jobID int) ([]StepModel, error)
	updateJobExecFn func(tx *sql.Tx, jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error)
}

func (m *jobsServiceRepoMock) GetJobByID(id int) (JobModel, error) {
	if m.getJobByIDFn != nil {
		return m.getJobByIDFn(id)
	}
	return JobModel{}, nil
}

func (m *jobsServiceRepoMock) ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error) {
	if m.listJobsFn != nil {
		return m.listJobsFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[JobModel]{}, nil
}

func (m *jobsServiceRepoMock) GetStepsByJobID(jobID int) ([]StepModel, error) {
	if m.getStepsByJobFn != nil {
		return m.getStepsByJobFn(jobID)
	}
	return nil, nil
}

func (m *jobsServiceRepoMock) UpdateJobExecution(tx *sql.Tx, jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error) {
	if m.updateJobExecFn != nil {
		return m.updateJobExecFn(tx, jobID, status, startedAt, endedAt, cancelRequested, lastError)
	}
	return true, nil
}

func (m *jobsServiceRepoMock) GetDbContext() *database.DbContext { return nil }

func TestJobsServiceGetJobByID(t *testing.T) {
	now := time.Now()

	service := NewService(&jobsServiceRepoMock{
		getJobByIDFn: func(id int) (JobModel, error) {
			return JobModel{ID: id, Type: "startup_scan", Priority: "low", Scope: []byte(`{"root":"/data"}`), Status: "running", CreatedAt: now}, nil
		},
		getStepsByJobFn: func(jobID int) ([]StepModel, error) {
			return []StepModel{
				{ID: 1, JobID: jobID, Status: "completed"},
				{ID: 2, JobID: jobID, Status: "running"},
				{ID: 3, JobID: jobID, Status: "skipped"},
			}, nil
		},
	})

	job, err := service.GetJobByID(12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID != 12 {
		t.Fatalf("expected id 12, got %d", job.ID)
	}
	if job.Progress.TotalSteps != 3 {
		t.Fatalf("expected total steps 3, got %d", job.Progress.TotalSteps)
	}
	if job.Progress.Progress != 66 {
		t.Fatalf("expected progress 66, got %d", job.Progress.Progress)
	}
}

func TestJobsServiceErrors(t *testing.T) {
	service := NewService(&jobsServiceRepoMock{})
	if _, err := service.GetJobByID(0); !errors.Is(err, ErrInvalidJobID) {
		t.Fatalf("expected ErrInvalidJobID, got %v", err)
	}

	service = NewService(&jobsServiceRepoMock{
		getJobByIDFn: func(id int) (JobModel, error) { return JobModel{}, sql.ErrNoRows },
	})
	if _, err := service.GetJobByID(1); !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestJobsServiceListAndSteps(t *testing.T) {
	now := time.Now()

	service := NewService(&jobsServiceRepoMock{
		listJobsFn: func(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error) {
			return utils.PaginationResponse[JobModel]{
				Items:      []JobModel{{ID: 1, Type: "upload_process", Priority: "high", Scope: []byte(`{"path":"/x"}`), Status: "queued", CreatedAt: now}},
				Pagination: utils.Pagination{Page: 1, PageSize: 20},
			}, nil
		},
		getJobByIDFn: func(id int) (JobModel, error) {
			return JobModel{ID: id, Type: "upload_process", Priority: "high", Scope: []byte(`{"path":"/x"}`), Status: "queued", CreatedAt: now}, nil
		},
		getStepsByJobFn: func(jobID int) ([]StepModel, error) {
			return []StepModel{{ID: 1, JobID: jobID, Type: "checksum", Status: "queued", DependsOn: []byte(`[2]`), Payload: []byte(`{"file_id":5}`), CreatedAt: now}}, nil
		},
	})

	jobs, err := service.ListJobs(JobFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("ListJobs returned error: %v", err)
	}
	if len(jobs.Items) != 1 {
		t.Fatalf("expected one job, got %d", len(jobs.Items))
	}

	steps, err := service.GetStepsByJobID(1)
	if err != nil {
		t.Fatalf("GetStepsByJobID returned error: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected one step, got %d", len(steps))
	}
	if len(steps[0].DependsOn) != 1 || steps[0].DependsOn[0] != 2 {
		t.Fatalf("unexpected dependencies: %+v", steps[0].DependsOn)
	}
}

func TestJobsServiceCancelJob(t *testing.T) {
	called := false
	service := NewService(&jobsServiceRepoMock{
		getJobByIDFn: func(id int) (JobModel, error) {
			return JobModel{ID: id, Status: "running"}, nil
		},
		updateJobExecFn: func(tx *sql.Tx, jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error) {
			called = true
			if status != "canceled" {
				t.Fatalf("expected canceled status, got %s", status)
			}
			return true, nil
		},
	})

	if err := service.CancelJob(10); err != nil {
		t.Fatalf("unexpected cancel error: %v", err)
	}
	if !called {
		t.Fatalf("expected UpdateJobExecution to be called")
	}
}

func TestJobsServiceAdditionalBranches(t *testing.T) {
	t.Run("GetJobByID returns repository error", func(t *testing.T) {
		expectedErr := errors.New("repo failed")
		service := NewService(&jobsServiceRepoMock{
			getJobByIDFn: func(id int) (JobModel, error) { return JobModel{}, expectedErr },
		})

		_, err := service.GetJobByID(1)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("GetJobByID returns step error", func(t *testing.T) {
		expectedErr := errors.New("steps failed")
		service := NewService(&jobsServiceRepoMock{
			getJobByIDFn:    func(id int) (JobModel, error) { return JobModel{ID: id, Scope: []byte(`{}`)}, nil },
			getStepsByJobFn: func(jobID int) ([]StepModel, error) { return nil, expectedErr },
		})

		_, err := service.GetJobByID(1)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected step error, got %v", err)
		}
	})

	t.Run("GetJobByID returns mapping error", func(t *testing.T) {
		service := NewService(&jobsServiceRepoMock{
			getJobByIDFn:    func(id int) (JobModel, error) { return JobModel{ID: id, Scope: []byte(`{`)}, nil },
			getStepsByJobFn: func(jobID int) ([]StepModel, error) { return nil, nil },
		})

		if _, err := service.GetJobByID(1); err == nil {
			t.Fatalf("expected invalid scope mapping error")
		}
	})

	t.Run("ListJobs normalizes pagination and returns mapping error", func(t *testing.T) {
		calledPage := 0
		calledSize := 0
		service := NewService(&jobsServiceRepoMock{
			listJobsFn: func(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error) {
				calledPage = page
				calledSize = pageSize
				return utils.PaginationResponse[JobModel]{
					Items: []JobModel{{ID: 1, Scope: []byte(`{`), Status: "queued"}},
				}, nil
			},
			getStepsByJobFn: func(jobID int) ([]StepModel, error) { return nil, nil },
		})

		if _, err := service.ListJobs(JobFilter{}, 0, 0); err == nil {
			t.Fatalf("expected job mapping error")
		}
		if calledPage != 1 || calledSize != 20 {
			t.Fatalf("expected normalized pagination 1/20, got %d/%d", calledPage, calledSize)
		}
	})

	t.Run("GetStepsByJobID handles invalid and missing jobs", func(t *testing.T) {
		service := NewService(&jobsServiceRepoMock{})
		if _, err := service.GetStepsByJobID(0); !errors.Is(err, ErrInvalidJobID) {
			t.Fatalf("expected invalid id error, got %v", err)
		}

		service = NewService(&jobsServiceRepoMock{
			getJobByIDFn: func(id int) (JobModel, error) { return JobModel{}, sql.ErrNoRows },
		})
		if _, err := service.GetStepsByJobID(2); !errors.Is(err, ErrJobNotFound) {
			t.Fatalf("expected job not found error, got %v", err)
		}
	})

	t.Run("GetStepsByJobID returns repository and mapping errors", func(t *testing.T) {
		stepErr := errors.New("step lookup failed")
		service := NewService(&jobsServiceRepoMock{
			getJobByIDFn:    func(id int) (JobModel, error) { return JobModel{ID: id}, nil },
			getStepsByJobFn: func(jobID int) ([]StepModel, error) { return nil, stepErr },
		})
		if _, err := service.GetStepsByJobID(1); !errors.Is(err, stepErr) {
			t.Fatalf("expected step repository error, got %v", err)
		}

		service = NewService(&jobsServiceRepoMock{
			getJobByIDFn: func(id int) (JobModel, error) { return JobModel{ID: id}, nil },
			getStepsByJobFn: func(jobID int) ([]StepModel, error) {
				return []StepModel{{ID: 1, JobID: jobID, DependsOn: []byte(`[`)}}, nil
			},
		})
		if _, err := service.GetStepsByJobID(1); err == nil {
			t.Fatalf("expected step mapping error")
		}
	})

	t.Run("CancelJob handles terminal states, missing jobs and transaction path", func(t *testing.T) {
		service := NewService(&jobsServiceRepoMock{})
		if err := service.CancelJob(0); !errors.Is(err, ErrInvalidJobID) {
			t.Fatalf("expected invalid id error, got %v", err)
		}

		service = NewService(&jobsServiceRepoMock{
			getJobByIDFn: func(id int) (JobModel, error) { return JobModel{}, sql.ErrNoRows },
		})
		if err := service.CancelJob(1); !errors.Is(err, ErrJobNotFound) {
			t.Fatalf("expected not found error, got %v", err)
		}

		called := false
		service = NewService(&jobsServiceRepoMock{
			getJobByIDFn: func(id int) (JobModel, error) { return JobModel{ID: id, Status: "completed"}, nil },
			updateJobExecFn: func(tx *sql.Tx, jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error) {
				called = true
				return true, nil
			},
		})
		if err := service.CancelJob(1); err != nil {
			t.Fatalf("expected terminal job cancel to succeed, got %v", err)
		}
		if called {
			t.Fatalf("expected terminal jobs to skip update")
		}

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock db: %v", err)
		}
		defer db.Close()
		mock.ExpectBegin()
		mock.ExpectCommit()

		execTxCalled := false
		service = &Service{Repository: &jobsServiceRepoMockWithContext{
			jobsServiceRepoMock: jobsServiceRepoMock{
				getJobByIDFn: func(id int) (JobModel, error) { return JobModel{ID: id, Status: "running"}, nil },
				updateJobExecFn: func(tx *sql.Tx, jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error) {
					if tx == nil {
						t.Fatalf("expected transaction to be provided")
					}
					if cancelRequested == nil || !*cancelRequested {
						t.Fatalf("expected cancelRequested=true")
					}
					execTxCalled = true
					return true, nil
				},
			},
			dbContext: database.NewDbContext(db),
		}}

		if err := service.CancelJob(9); err != nil {
			t.Fatalf("expected cancel with db context to succeed, got %v", err)
		}
		if !execTxCalled {
			t.Fatalf("expected transactional cancel path to run")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sqlmock expectations: %v", err)
		}
	})

	t.Run("helpers decode and calculate progress branches", func(t *testing.T) {
		if scope, err := decodeJSONMap(nil); err != nil || len(scope) != 0 {
			t.Fatalf("expected empty scope without error, got %#v / %v", scope, err)
		}
		if payload, err := decodeJSONAny(nil); err != nil || payload != nil {
			t.Fatalf("expected nil empty payload without error, got %#v / %v", payload, err)
		}
		if _, err := decodeJSONMap([]byte(`{`)); err == nil {
			t.Fatalf("expected invalid scope json error")
		}
		if _, err := decodeJSONAny([]byte(`{`)); err == nil {
			t.Fatalf("expected invalid payload json error")
		}

		progress := calculateProgress([]StepModel{
			{Status: "completed"},
			{Status: "running"},
			{Status: "failed"},
			{Status: "skipped"},
			{Status: "canceled"},
		})
		if progress.CompletedSteps != 1 || progress.RunningSteps != 1 || progress.FailedSteps != 1 || progress.SkippedSteps != 1 || progress.CanceledSteps != 1 {
			t.Fatalf("unexpected progress counters: %+v", progress)
		}
		if progress.Progress != 60 {
			t.Fatalf("expected progress 60, got %d", progress.Progress)
		}
	})
}

type jobsServiceRepoMockWithContext struct {
	jobsServiceRepoMock
	dbContext *database.DbContext
}

func (m *jobsServiceRepoMockWithContext) GetDbContext() *database.DbContext {
	return m.dbContext
}
