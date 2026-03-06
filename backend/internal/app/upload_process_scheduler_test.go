package app

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
)

type uploadSchedulerJobsRepoStub struct {
	dbContext *database.DbContext
	jobs      []jobs.JobModel
	steps     []jobs.StepModel
}

func (r *uploadSchedulerJobsRepoStub) GetDbContext() *database.DbContext { return r.dbContext }
func (r *uploadSchedulerJobsRepoStub) CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error) {
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now().UTC()
	}
	r.jobs = append(r.jobs, job)
	return job, nil
}
func (r *uploadSchedulerJobsRepoStub) CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error) {
	if step.CreatedAt.IsZero() {
		step.CreatedAt = time.Now().UTC()
	}
	r.steps = append(r.steps, step)
	return step, nil
}
func (r *uploadSchedulerJobsRepoStub) GetJobByID(id string) (jobs.JobModel, error) {
	return jobs.JobModel{}, sql.ErrNoRows
}
func (r *uploadSchedulerJobsRepoStub) ListJobs(filter jobs.JobFilter, page int, pageSize int) (utils.PaginationResponse[jobs.JobModel], error) {
	return utils.PaginationResponse[jobs.JobModel]{}, nil
}
func (r *uploadSchedulerJobsRepoStub) GetStepsByJobID(jobID string) ([]jobs.StepModel, error) {
	result := make([]jobs.StepModel, 0)
	for _, step := range r.steps {
		if step.JobID == jobID {
			result = append(result, step)
		}
	}
	return result, nil
}
func (r *uploadSchedulerJobsRepoStub) UpdateJobStatus(tx *sql.Tx, id string, fromStatus string, toStatus string, startedAt *time.Time, endedAt *time.Time, lastError string) (bool, error) {
	return false, nil
}
func (r *uploadSchedulerJobsRepoStub) UpdateStepStatus(tx *sql.Tx, id string, fromStatus string, toStatus string, startedAt *time.Time, endedAt *time.Time, lastError string) (bool, error) {
	return false, nil
}
func (r *uploadSchedulerJobsRepoStub) UpdateStepExecution(tx *sql.Tx, id string, attempts int, lastError string, progress int, startedAt *time.Time, endedAt *time.Time) (bool, error) {
	return false, nil
}
func (r *uploadSchedulerJobsRepoStub) RequestJobCancel(tx *sql.Tx, id string) (bool, error) {
	return false, nil
}
func (r *uploadSchedulerJobsRepoStub) RequestJobCancelCascade(tx *sql.Tx, id string) (bool, error) {
	return false, nil
}

func TestUploadProcessSchedulerCreatesUploadJobsAndSteps(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectCommit()

	repo := &uploadSchedulerJobsRepoStub{
		dbContext: database.NewDbContext(db),
	}
	scheduler := newUploadProcessScheduler(repo)
	if scheduler == nil {
		t.Fatalf("expected scheduler instance")
	}

	result, err := scheduler.ScheduleUploadProcess([]string{"/tmp/picture.jpg", "/tmp/movie.mp4"})
	if err != nil {
		t.Fatalf("expected upload scheduling to succeed, got %v", err)
	}
	if result.JobID == "" {
		t.Fatalf("expected non-empty root job id")
	}
	if len(result.Jobs) != 2 {
		t.Fatalf("expected one job reference per uploaded file, got %d", len(result.Jobs))
	}
	if len(repo.jobs) != 2 {
		t.Fatalf("expected two jobs persisted, got %d", len(repo.jobs))
	}

	for _, job := range repo.jobs {
		if job.Type != string(domain.JobTypeUploadProcess) {
			t.Fatalf("expected upload_process job type, got %s", job.Type)
		}
		if job.Priority != int(domain.JobPriorityHigh) {
			t.Fatalf("expected high priority upload job, got %d", job.Priority)
		}
	}

	stepsByType := map[string]int{}
	for _, step := range repo.steps {
		stepsByType[step.Type]++
	}
	if stepsByType[string(domain.StepTypeChecksum)] != 2 {
		t.Fatalf("expected checksum step for each upload, got %d", stepsByType[string(domain.StepTypeChecksum)])
	}
	if stepsByType[string(domain.StepTypePersist)] != 2 {
		t.Fatalf("expected persist step for each upload, got %d", stepsByType[string(domain.StepTypePersist)])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected sqlmock expectations: %v", err)
	}
}
