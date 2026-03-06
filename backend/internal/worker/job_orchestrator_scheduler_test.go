package worker

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	jobsapi "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type fakeJobsRepository struct {
	jobsapi.RepositoryInterface

	mu         sync.Mutex
	jobs       map[int]jobsapi.JobModel
	steps      map[int]jobsapi.StepModel
	nextJobID  int
	nextStepID int
}

func newFakeJobsRepository() *fakeJobsRepository {
	return &fakeJobsRepository{
		jobs:       map[int]jobsapi.JobModel{},
		steps:      map[int]jobsapi.StepModel{},
		nextJobID:  1,
		nextStepID: 1,
	}
}

func (r *fakeJobsRepository) GetDbContext() *database.DbContext {
	return nil
}

func (r *fakeJobsRepository) CreateJob(tx *sql.Tx, job jobsapi.JobModel) (jobsapi.JobModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	job.ID = r.nextJobID
	r.nextJobID++
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}

	r.jobs[job.ID] = job
	return job, nil
}

func (r *fakeJobsRepository) CreateStep(tx *sql.Tx, step jobsapi.StepModel) (jobsapi.StepModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	step.ID = r.nextStepID
	r.nextStepID++
	if step.CreatedAt.IsZero() {
		step.CreatedAt = time.Now()
	}

	r.steps[step.ID] = step
	return step, nil
}

func (r *fakeJobsRepository) GetJobByID(id int) (jobsapi.JobModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, exists := r.jobs[id]
	if !exists {
		return jobsapi.JobModel{}, sql.ErrNoRows
	}

	return job, nil
}

func (r *fakeJobsRepository) ListJobs(filter jobsapi.JobFilter, page int, pageSize int) (utils.PaginationResponse[jobsapi.JobModel], error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := utils.PaginationResponse[jobsapi.JobModel]{
		Items: []jobsapi.JobModel{},
		Pagination: utils.Pagination{
			Page: page, PageSize: pageSize,
		},
	}

	for _, job := range r.jobs {
		out.Items = append(out.Items, job)
	}
	sort.Slice(out.Items, func(i, j int) bool { return out.Items[i].ID < out.Items[j].ID })
	out.UpdatePagination()

	return out, nil
}

func (r *fakeJobsRepository) GetStepsByJobID(jobID int) ([]jobsapi.StepModel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	steps := []jobsapi.StepModel{}
	for _, step := range r.steps {
		if step.JobID == jobID {
			steps = append(steps, step)
		}
	}

	sort.Slice(steps, func(i, j int) bool { return steps[i].ID < steps[j].ID })
	return steps, nil
}

func (r *fakeJobsRepository) UpdateJobExecution(tx *sql.Tx, jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, exists := r.jobs[jobID]
	if !exists {
		return false, nil
	}

	job.Status = status
	if startedAt != nil {
		value := *startedAt
		job.StartedAt = &value
	}
	if endedAt != nil {
		value := *endedAt
		job.EndedAt = &value
	}
	if cancelRequested != nil {
		job.CancelRequested = *cancelRequested
	}
	if lastError != nil {
		job.LastError = *lastError
	}

	r.jobs[jobID] = job
	return true, nil
}

func (r *fakeJobsRepository) UpdateStepExecution(tx *sql.Tx, stepID int, status string, progress int, attempts int, startedAt *time.Time, endedAt *time.Time, lastError *string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	step, exists := r.steps[stepID]
	if !exists {
		return false, nil
	}

	step.Status = status
	step.Progress = progress
	step.Attempts = attempts
	if startedAt != nil {
		value := *startedAt
		step.StartedAt = &value
	}
	if endedAt != nil {
		value := *endedAt
		step.EndedAt = &value
	}
	if lastError != nil {
		step.LastError = *lastError
	}

	r.steps[stepID] = step
	return true, nil
}

func TestJobOrchestratorCreateAndSchedule(t *testing.T) {
	fakeRepository := newFakeJobsRepository()

	var executed []StepType
	scheduler := NewJobScheduler(fakeRepository, map[StepType]StepExecutor{
		StepTypeScanFilesystem: func(step jobsapi.StepModel) error {
			executed = append(executed, StepType(step.Type))
			return nil
		},
		StepTypeChecksum: func(step jobsapi.StepModel) error {
			executed = append(executed, StepType(step.Type))
			return nil
		},
	})

	orchestrator := NewJobOrchestrator(fakeRepository, scheduler)
	jobID, err := orchestrator.CreateJob(PlannedJob{
		Type:     JobTypeStartupScan,
		Priority: JobPriorityLow,
		Scope:    JobScope{Root: "/data"},
		Steps: []PlannedStep{
			{
				Key:  "scan",
				Type: StepTypeScanFilesystem,
			},
			{
				Key:       "checksum",
				Type:      StepTypeChecksum,
				DependsOn: []string{"scan"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected create job error: %v", err)
	}

	if err := scheduler.processJob(jobID); err != nil {
		t.Fatalf("processJob returned error: %v", err)
	}

	job, err := fakeRepository.GetJobByID(jobID)
	if err != nil {
		t.Fatalf("GetJobByID returned error: %v", err)
	}
	if job.Status != string(JobStatusCompleted) {
		t.Fatalf("expected completed job, got %s", job.Status)
	}

	steps, err := fakeRepository.GetStepsByJobID(jobID)
	if err != nil {
		t.Fatalf("GetStepsByJobID returned error: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}

	expectedOrder := []StepType{StepTypeScanFilesystem, StepTypeChecksum}
	if len(executed) != len(expectedOrder) {
		t.Fatalf("unexpected execution count. expected=%d got=%d", len(expectedOrder), len(executed))
	}
	for index := range expectedOrder {
		if executed[index] != expectedOrder[index] {
			t.Fatalf("unexpected execution order at index %d: expected=%s got=%s", index, expectedOrder[index], executed[index])
		}
	}

	dependencies := []int{}
	if err := json.Unmarshal(steps[1].DependsOn, &dependencies); err != nil {
		t.Fatalf("invalid dependency json: %v", err)
	}
	if len(dependencies) != 1 || dependencies[0] != steps[0].ID {
		t.Fatalf("unexpected dependency relation: %+v", dependencies)
	}
}

func TestJobOrchestratorValidatesInvalidDependencies(t *testing.T) {
	fakeRepository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(fakeRepository, NewJobScheduler(fakeRepository, nil))

	_, err := orchestrator.CreateJob(PlannedJob{
		Type:     JobTypeStartupScan,
		Priority: JobPriorityLow,
		Steps: []PlannedStep{
			{
				Key:       "checksum",
				Type:      StepTypeChecksum,
				DependsOn: []string{"scan"},
			},
		},
	})
	if err == nil {
		t.Fatalf("expected invalid dependency error")
	}
}

func TestJobSchedulerMarksPartialFailure(t *testing.T) {
	fakeRepository := newFakeJobsRepository()

	scheduler := NewJobScheduler(fakeRepository, map[StepType]StepExecutor{
		StepTypeScanFilesystem: func(step jobsapi.StepModel) error {
			return nil
		},
		StepTypeChecksum: func(step jobsapi.StepModel) error {
			return errors.New("checksum failed")
		},
	})

	orchestrator := NewJobOrchestrator(fakeRepository, scheduler)
	jobID, err := orchestrator.CreateJob(PlannedJob{
		Type:     JobTypeStartupScan,
		Priority: JobPriorityLow,
		Steps: []PlannedStep{
			{
				Key:  "scan",
				Type: StepTypeScanFilesystem,
			},
			{
				Key:       "checksum",
				Type:      StepTypeChecksum,
				DependsOn: []string{"scan"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if err := scheduler.processJob(jobID); err != nil {
		t.Fatalf("unexpected processJob error: %v", err)
	}

	job, err := fakeRepository.GetJobByID(jobID)
	if err != nil {
		t.Fatalf("GetJobByID returned error: %v", err)
	}
	if job.Status != string(JobStatusPartialFail) {
		t.Fatalf("expected partial_fail status, got %s", job.Status)
	}
}
