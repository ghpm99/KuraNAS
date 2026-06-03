package ollama

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type fakeJobsRepo struct {
	lastJob  jobs.JobModel
	lastStep jobs.StepModel
}

func (r *fakeJobsRepo) GetDbContext() *database.DbContext { return database.NewDbContext(nil) }
func (r *fakeJobsRepo) CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error) {
	job.ID = 42
	r.lastJob = job
	return job, nil
}
func (r *fakeJobsRepo) CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error) {
	step.ID = 7
	r.lastStep = step
	return step, nil
}
func (r *fakeJobsRepo) GetJobByID(id int) (jobs.JobModel, error) { return jobs.JobModel{}, nil }
func (r *fakeJobsRepo) ListJobs(filter jobs.JobFilter, page int, pageSize int) (utils.PaginationResponse[jobs.JobModel], error) {
	return utils.PaginationResponse[jobs.JobModel]{}, nil
}
func (r *fakeJobsRepo) GetStepsByJobID(jobID int) ([]jobs.StepModel, error) { return nil, nil }
func (r *fakeJobsRepo) UpdateJobExecution(tx *sql.Tx, jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error) {
	return true, nil
}
func (r *fakeJobsRepo) UpdateStepExecution(tx *sql.Tx, stepID int, status string, progress int, attempts int, startedAt *time.Time, endedAt *time.Time, lastError *string) (bool, error) {
	return true, nil
}
func (r *fakeJobsRepo) DeferStepForTimeout(tx *sql.Tx, stepID int, attempts int, lastError string) (bool, error) {
	return true, nil
}
func (r *fakeJobsRepo) RequeueJob(tx *sql.Tx, jobID int) (bool, error) { return true, nil }
func (r *fakeJobsRepo) RecoverInterruptedWork(tx *sql.Tx) (int64, int64, error) {
	return 0, 0, nil
}

func staticBaseURL(url string) func() string {
	return func() string { return url }
}

func TestGetStatusReachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			w.Write([]byte(`{"version":"0.5.0"}`))
		case "/api/tags":
			w.Write([]byte(`{"models":[{"name":"llama3.1:latest","size":123,"digest":"abc","details":{"parameter_size":"8B"}}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	svc := NewService(staticBaseURL(server.URL), nil)
	status := svc.GetStatus(context.Background())

	if !status.Reachable || status.Version != "0.5.0" {
		t.Fatalf("expected reachable 0.5.0, got %+v", status)
	}
	if len(status.Models) != 1 || status.Models[0].Name != "llama3.1:latest" {
		t.Fatalf("unexpected models: %+v", status.Models)
	}
}

func TestGetStatusUnreachable(t *testing.T) {
	svc := NewService(staticBaseURL("http://127.0.0.1:0"), nil)
	status := svc.GetStatus(context.Background())
	if status.Reachable {
		t.Fatalf("expected unreachable daemon")
	}
}

func TestDeleteModelNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := NewService(staticBaseURL(server.URL), nil)
	if err := svc.DeleteModel(context.Background(), "missing"); err != ErrModelNotFound {
		t.Fatalf("expected ErrModelNotFound, got %v", err)
	}
}

func TestDeleteModelEmptyName(t *testing.T) {
	svc := NewService(staticBaseURL("http://x"), nil)
	if err := svc.DeleteModel(context.Background(), "  "); err != ErrInvalidModelName {
		t.Fatalf("expected ErrInvalidModelName, got %v", err)
	}
}

func TestPullModelEnqueuesJob(t *testing.T) {
	repo := &fakeJobsRepo{}
	svc := NewService(staticBaseURL("http://nas:11434"), repo)

	jobID, err := svc.PullModel("llama3.1")
	if err != nil {
		t.Fatalf("PullModel error: %v", err)
	}
	if jobID != 42 {
		t.Fatalf("expected job id 42, got %d", jobID)
	}
	if repo.lastJob.Type != pullJobType {
		t.Fatalf("expected job type %s, got %s", pullJobType, repo.lastJob.Type)
	}
	if repo.lastStep.Type != pullStepType {
		t.Fatalf("expected step type %s, got %s", pullStepType, repo.lastStep.Type)
	}
}

func TestPullModelValidations(t *testing.T) {
	if _, err := NewService(staticBaseURL("http://x"), &fakeJobsRepo{}).PullModel("  "); err != ErrInvalidModelName {
		t.Fatalf("expected ErrInvalidModelName, got %v", err)
	}
	if _, err := NewService(staticBaseURL("http://x"), nil).PullModel("llama3.1"); err != ErrJobsUnavailable {
		t.Fatalf("expected ErrJobsUnavailable, got %v", err)
	}
}
