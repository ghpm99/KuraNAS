package worker

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jobs "nas-go/api/internal/api/v1/jobs"
	ollamaapi "nas-go/api/internal/api/v1/ollama"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type progressJobsRepo struct {
	progress []int
}

func (r *progressJobsRepo) GetDbContext() *database.DbContext { return database.NewDbContext(nil) }
func (r *progressJobsRepo) CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error) {
	return job, nil
}
func (r *progressJobsRepo) CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error) {
	return step, nil
}
func (r *progressJobsRepo) GetJobByID(id int) (jobs.JobModel, error) { return jobs.JobModel{}, nil }
func (r *progressJobsRepo) ListJobs(filter jobs.JobFilter, page int, pageSize int) (utils.PaginationResponse[jobs.JobModel], error) {
	return utils.PaginationResponse[jobs.JobModel]{}, nil
}
func (r *progressJobsRepo) GetStepsByJobID(jobID int) ([]jobs.StepModel, error) { return nil, nil }
func (r *progressJobsRepo) UpdateJobExecution(tx *sql.Tx, jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error) {
	return true, nil
}
func (r *progressJobsRepo) UpdateStepExecution(tx *sql.Tx, stepID int, status string, progress int, attempts int, startedAt *time.Time, endedAt *time.Time, lastError *string) (bool, error) {
	r.progress = append(r.progress, progress)
	return true, nil
}
func (r *progressJobsRepo) DeferStepForTimeout(tx *sql.Tx, stepID int, attempts int, lastError string) (bool, error) {
	return true, nil
}
func (r *progressJobsRepo) RequeueJob(tx *sql.Tx, jobID int) (bool, error) { return true, nil }

func TestExecuteOllamaPullStepReportsProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pull" {
			t.Fatalf("expected /api/pull, got %s", r.URL.Path)
		}
		flusher, _ := w.(http.Flusher)
		lines := []string{
			`{"status":"pulling manifest"}`,
			`{"status":"downloading","total":100,"completed":10}`,
			`{"status":"downloading","total":100,"completed":100}`,
			`{"status":"success"}`,
		}
		for _, line := range lines {
			w.Write([]byte(line + "\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}
	}))
	defer server.Close()

	repo := &progressJobsRepo{}
	ctx := &WorkerContext{JobsRepository: repo}

	payload, _ := json.Marshal(ollamaapi.PullStepPayload{Model: "llama3.1", BaseURL: server.URL})
	step := jobs.StepModel{ID: 1, Payload: payload}

	if err := executeOllamaPullStep(ctx, step); err != nil {
		t.Fatalf("executeOllamaPullStep error: %v", err)
	}
	if len(repo.progress) == 0 {
		t.Fatalf("expected progress to be reported")
	}
	if repo.progress[len(repo.progress)-1] != 100 {
		t.Fatalf("expected final progress 100, got %v", repo.progress)
	}
}

func TestExecuteOllamaPullStepError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":"model not found"}` + "\n"))
	}))
	defer server.Close()

	payload, _ := json.Marshal(ollamaapi.PullStepPayload{Model: "missing", BaseURL: server.URL})
	if err := executeOllamaPullStep(&WorkerContext{JobsRepository: &progressJobsRepo{}}, jobs.StepModel{Payload: payload}); err == nil {
		t.Fatalf("expected error when pull stream reports an error")
	}
}

func TestExecuteOllamaPullStepInvalidPayload(t *testing.T) {
	if err := executeOllamaPullStep(&WorkerContext{}, jobs.StepModel{Payload: []byte("not json")}); err == nil {
		t.Fatalf("expected error on invalid payload")
	}
	emptyModel, _ := json.Marshal(ollamaapi.PullStepPayload{BaseURL: "http://x"})
	if err := executeOllamaPullStep(&WorkerContext{}, jobs.StepModel{Payload: emptyModel}); err == nil {
		t.Fatalf("expected error on empty model")
	}
}
