package jobs

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
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
