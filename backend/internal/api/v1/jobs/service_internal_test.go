package jobs

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type jobsRepoMock struct {
	getJobByIDFn    func(id string) (JobModel, error)
	listJobsFn      func(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error)
	getStepsByJobFn func(jobID string) ([]StepModel, error)
	requestCancelFn func(tx *sql.Tx, id string) (bool, error)
}

func (m *jobsRepoMock) GetDbContext() *database.DbContext { return nil }
func (m *jobsRepoMock) CreateJob(tx *sql.Tx, job JobModel) (JobModel, error) {
	return JobModel{}, errors.New("not used")
}
func (m *jobsRepoMock) CreateStep(tx *sql.Tx, step StepModel) (StepModel, error) {
	return StepModel{}, errors.New("not used")
}
func (m *jobsRepoMock) GetJobByID(id string) (JobModel, error) {
	if m.getJobByIDFn != nil {
		return m.getJobByIDFn(id)
	}
	return JobModel{}, errors.New("not used")
}
func (m *jobsRepoMock) ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error) {
	if m.listJobsFn != nil {
		return m.listJobsFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[JobModel]{}, errors.New("not used")
}
func (m *jobsRepoMock) GetStepsByJobID(jobID string) ([]StepModel, error) {
	if m.getStepsByJobFn != nil {
		return m.getStepsByJobFn(jobID)
	}
	return nil, errors.New("not used")
}
func (m *jobsRepoMock) UpdateJobStatus(tx *sql.Tx, id string, fromStatus string, toStatus string, startedAt *time.Time, endedAt *time.Time, lastError string) (bool, error) {
	return false, errors.New("not used")
}
func (m *jobsRepoMock) UpdateStepStatus(tx *sql.Tx, id string, fromStatus string, toStatus string, startedAt *time.Time, endedAt *time.Time, lastError string) (bool, error) {
	return false, errors.New("not used")
}
func (m *jobsRepoMock) UpdateStepExecution(tx *sql.Tx, id string, attempts int, lastError string, progress int, startedAt *time.Time, endedAt *time.Time) (bool, error) {
	return false, errors.New("not used")
}
func (m *jobsRepoMock) RequestJobCancel(tx *sql.Tx, id string) (bool, error) {
	if m.requestCancelFn != nil {
		return m.requestCancelFn(tx, id)
	}
	return false, errors.New("not used")
}

func TestJobsServiceGetJobByIDAggregatesProgress(t *testing.T) {
	repo := &jobsRepoMock{
		getJobByIDFn: func(id string) (JobModel, error) {
			return JobModel{ID: id, Type: "startup_scan", Status: "running", Priority: 2, CreatedAt: time.Now()}, nil
		},
		getStepsByJobFn: func(jobID string) ([]StepModel, error) {
			return []StepModel{
				{ID: "s1", JobID: jobID, Status: "queued", Progress: 0},
				{ID: "s2", JobID: jobID, Status: "running", Progress: 50},
				{ID: "s3", JobID: jobID, Status: "completed", Progress: 100},
			}, nil
		},
	}

	svc := NewService(repo)
	result, err := svc.GetJobByID("job-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Progress.Progress != 50 {
		t.Fatalf("expected progress 50, got %d", result.Progress.Progress)
	}
	if result.Progress.TotalSteps != 3 || result.Progress.CompletedSteps != 1 || result.Progress.RunningSteps != 1 || result.Progress.QueuedSteps != 1 {
		t.Fatalf("unexpected progress summary: %+v", result.Progress)
	}
}

func TestJobsServiceListJobsAndValidation(t *testing.T) {
	repo := &jobsRepoMock{
		listJobsFn: func(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error) {
			return utils.PaginationResponse[JobModel]{
				Items:      []JobModel{{ID: "job-1", Status: "failed", CreatedAt: time.Now()}},
				Pagination: utils.Pagination{Page: page, PageSize: pageSize},
			}, nil
		},
		getStepsByJobFn: func(jobID string) ([]StepModel, error) {
			return []StepModel{{ID: "s1", JobID: jobID, Status: "failed", Progress: 20}}, nil
		},
	}
	svc := NewService(repo)

	page, err := svc.ListJobs(JobFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page.Items))
	}
	if page.Items[0].Progress.Progress != 100 {
		t.Fatalf("expected terminal failed step to aggregate as 100, got %d", page.Items[0].Progress.Progress)
	}

	if _, err := svc.ListJobs(JobFilter{}, 0, 10); !errors.Is(err, ErrInvalidPage) {
		t.Fatalf("expected ErrInvalidPage, got %v", err)
	}
	if _, err := svc.ListJobs(JobFilter{}, 1, 0); !errors.Is(err, ErrInvalidPageSize) {
		t.Fatalf("expected ErrInvalidPageSize, got %v", err)
	}
}

func TestJobsServiceGetStepsByJobIDNotFound(t *testing.T) {
	repo := &jobsRepoMock{
		getJobByIDFn: func(id string) (JobModel, error) {
			return JobModel{}, sql.ErrNoRows
		},
	}

	svc := NewService(repo)
	_, err := svc.GetStepsByJobID("missing")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestJobsServiceCancelJob(t *testing.T) {
	repo := &jobsRepoMock{
		getJobByIDFn: func(id string) (JobModel, error) {
			return JobModel{
				ID:              id,
				Type:            "startup_scan",
				Status:          "running",
				Priority:        2,
				CreatedAt:       time.Now().UTC(),
				CancelRequested: false,
			}, nil
		},
		requestCancelFn: func(tx *sql.Tx, id string) (bool, error) {
			return true, nil
		},
		getStepsByJobFn: func(jobID string) ([]StepModel, error) {
			return []StepModel{}, nil
		},
	}

	svc := NewService(repo)
	result, err := svc.CancelJob("job-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != "job-1" {
		t.Fatalf("expected canceled job id, got %s", result.ID)
	}
}

func TestJobsServiceCancelJobNotAllowed(t *testing.T) {
	repo := &jobsRepoMock{
		getJobByIDFn: func(id string) (JobModel, error) {
			return JobModel{
				ID:        id,
				Type:      "upload_process",
				Status:    "running",
				CreatedAt: time.Now().UTC(),
			}, nil
		},
	}

	svc := NewService(repo)
	_, err := svc.CancelJob("job-1")
	if !errors.Is(err, ErrJobCancelNotAllowed) {
		t.Fatalf("expected ErrJobCancelNotAllowed, got %v", err)
	}
}
