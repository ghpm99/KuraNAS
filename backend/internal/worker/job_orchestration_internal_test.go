package worker

import (
	"database/sql"
	"encoding/json"
	"sort"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type inMemoryJobsRepository struct {
	jobsByID  map[string]jobs.JobModel
	stepsByID map[string]jobs.StepModel
}

func newInMemoryJobsRepository() *inMemoryJobsRepository {
	return &inMemoryJobsRepository{
		jobsByID:  map[string]jobs.JobModel{},
		stepsByID: map[string]jobs.StepModel{},
	}
}

func (r *inMemoryJobsRepository) GetDbContext() *database.DbContext { return nil }

func (r *inMemoryJobsRepository) CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error) {
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now().UTC()
	}
	r.jobsByID[job.ID] = job
	return job, nil
}

func (r *inMemoryJobsRepository) CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error) {
	if step.CreatedAt.IsZero() {
		step.CreatedAt = time.Now().UTC()
	}
	r.stepsByID[step.ID] = step
	return step, nil
}

func (r *inMemoryJobsRepository) GetJobByID(id string) (jobs.JobModel, error) {
	return r.jobsByID[id], nil
}

func (r *inMemoryJobsRepository) ListJobs(filter jobs.JobFilter, page int, pageSize int) (utils.PaginationResponse[jobs.JobModel], error) {
	items := make([]jobs.JobModel, 0, len(r.jobsByID))
	for _, job := range r.jobsByID {
		if filter.Status.HasValue && job.Status != filter.Status.Value {
			continue
		}
		if filter.Type.HasValue && job.Type != filter.Type.Value {
			continue
		}
		if filter.Priority.HasValue && job.Priority != filter.Priority.Value {
			continue
		}
		items = append(items, job)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return utils.PaginationResponse[jobs.JobModel]{
		Items: items,
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
	}, nil
}

func (r *inMemoryJobsRepository) GetStepsByJobID(jobID string) ([]jobs.StepModel, error) {
	steps := []jobs.StepModel{}
	for _, step := range r.stepsByID {
		if step.JobID == jobID {
			steps = append(steps, step)
		}
	}
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].CreatedAt.Before(steps[j].CreatedAt)
	})
	return steps, nil
}

func (r *inMemoryJobsRepository) UpdateJobStatus(tx *sql.Tx, id string, fromStatus string, toStatus string, startedAt *time.Time, endedAt *time.Time, lastError string) (bool, error) {
	job, exists := r.jobsByID[id]
	if !exists || job.Status != fromStatus {
		return false, nil
	}

	job.Status = toStatus
	if startedAt != nil {
		job.StartedAt = sql.NullTime{Valid: true, Time: *startedAt}
	}
	if endedAt != nil {
		job.EndedAt = sql.NullTime{Valid: true, Time: *endedAt}
	}
	job.LastError = lastError
	r.jobsByID[id] = job
	return true, nil
}

func (r *inMemoryJobsRepository) UpdateStepStatus(tx *sql.Tx, id string, fromStatus string, toStatus string, startedAt *time.Time, endedAt *time.Time, lastError string) (bool, error) {
	step, exists := r.stepsByID[id]
	if !exists || step.Status != fromStatus {
		return false, nil
	}

	step.Status = toStatus
	if startedAt != nil {
		step.StartedAt = sql.NullTime{Valid: true, Time: *startedAt}
	}
	if endedAt != nil {
		step.EndedAt = sql.NullTime{Valid: true, Time: *endedAt}
	}
	step.LastError = lastError
	r.stepsByID[id] = step
	return true, nil
}

func (r *inMemoryJobsRepository) UpdateStepExecution(tx *sql.Tx, id string, attempts int, lastError string, progress int, startedAt *time.Time, endedAt *time.Time) (bool, error) {
	step, exists := r.stepsByID[id]
	if !exists {
		return false, nil
	}
	step.Attempts = attempts
	step.LastError = lastError
	step.Progress = progress
	if startedAt != nil {
		step.StartedAt = sql.NullTime{Valid: true, Time: *startedAt}
	}
	if endedAt != nil {
		step.EndedAt = sql.NullTime{Valid: true, Time: *endedAt}
	}
	r.stepsByID[id] = step
	return true, nil
}

type recordingExecutor struct {
	executed []string
}

func (e *recordingExecutor) ExecuteStep(step domain.Step, context *WorkerContext) error {
	e.executed = append(e.executed, step.ID)
	return nil
}

type skippedExecutor struct {
	executed []string
}

func (e *skippedExecutor) ExecuteStep(step domain.Step, context *WorkerContext) error {
	e.executed = append(e.executed, step.ID)
	return newStepSkipped("already up-to-date")
}

func TestJobOrchestratorCreateJobPersistsStepsWithDependencies(t *testing.T) {
	repo := newInMemoryJobsRepository()
	orchestrator := NewJobOrchestrator(repo, NewDefaultJobPlanner())
	orchestrator.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	job, err := orchestrator.CreateJob(
		domain.JobTypeStartupScan,
		domain.JobPriorityNormal,
		domain.NewRootScopePayload("/tmp"),
	)
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}

	storedJob, ok := repo.jobsByID[job.ID]
	if !ok {
		t.Fatalf("expected job to be persisted")
	}
	if storedJob.Status != string(domain.JobStatusQueued) {
		t.Fatalf("expected queued status, got %s", storedJob.Status)
	}

	steps, err := repo.GetStepsByJobID(job.ID)
	if err != nil {
		t.Fatalf("GetStepsByJobID failed: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}

	stepByType := map[string]jobs.StepModel{}
	for _, step := range steps {
		stepByType[step.Type] = step
	}

	scanStep := stepByType[string(domain.StepTypeScanFilesystem)]
	diffStep := stepByType[string(domain.StepTypeDiffAgainstDB)]
	if scanStep.ID == "" || diffStep.ID == "" {
		t.Fatalf("expected scan and diff steps to exist")
	}

	dependsOn := []string{}
	if err := json.Unmarshal([]byte(diffStep.DependsOnJSON), &dependsOn); err != nil {
		t.Fatalf("failed to parse dependencies: %v", err)
	}
	if len(dependsOn) != 1 || dependsOn[0] != scanStep.ID {
		t.Fatalf("expected diff step to depend on scan step id, got %#v", dependsOn)
	}
}

func TestSchedulerRespectsStepDependencies(t *testing.T) {
	repo := newInMemoryJobsRepository()
	orchestrator := NewJobOrchestrator(repo, NewDefaultJobPlanner())
	orchestrator.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	job, err := orchestrator.CreateJob(
		domain.JobTypeStartupScan,
		domain.JobPriorityNormal,
		domain.NewRootScopePayload("/tmp"),
	)
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}

	executor := &recordingExecutor{}
	scheduler := NewJobScheduler(repo, executor, &WorkerContext{})
	scheduler.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	if err := scheduler.RunOnce(); err != nil {
		t.Fatalf("RunOnce failed: %v", err)
	}
	if len(executor.executed) != 1 {
		t.Fatalf("expected one executed step in first tick, got %d", len(executor.executed))
	}

	stepsAfterFirstTick, _ := repo.GetStepsByJobID(job.ID)
	completedCount := 0
	queuedCount := 0
	for _, step := range stepsAfterFirstTick {
		switch step.Status {
		case string(domain.StepStatusCompleted):
			completedCount++
		case string(domain.StepStatusQueued):
			queuedCount++
		}
	}
	if completedCount != 1 || queuedCount != 1 {
		t.Fatalf("expected one completed and one queued after first tick, got completed=%d queued=%d", completedCount, queuedCount)
	}

	if err := scheduler.RunOnce(); err != nil {
		t.Fatalf("second RunOnce failed: %v", err)
	}
	if len(executor.executed) != 2 {
		t.Fatalf("expected two executed steps after second tick, got %d", len(executor.executed))
	}

	storedJob, _ := repo.GetJobByID(job.ID)
	if storedJob.Status != string(domain.JobStatusCompleted) {
		t.Fatalf("expected completed job after all steps, got %s", storedJob.Status)
	}
}

func TestSelectNextEligibleStepHonorsDependencies(t *testing.T) {
	steps := []schedulerStep{
		{
			Step: domain.Step{ID: "step-a", Status: domain.StepStatusQueued},
		},
		{
			Step:            domain.Step{ID: "step-b", Status: domain.StepStatusQueued},
			DependsOnStepID: []string{"step-a"},
		},
	}

	next := selectNextEligibleStep(steps)
	if next == nil || next.Step.ID != "step-a" {
		t.Fatalf("expected step-a to be selected first, got %#v", next)
	}

	steps[0].Step.Status = domain.StepStatusCompleted
	next = selectNextEligibleStep(steps)
	if next == nil || next.Step.ID != "step-b" {
		t.Fatalf("expected step-b to be selected after dependency completion, got %#v", next)
	}
}

func TestSchedulerTreatsSkippedStepAsSuccessfulCompletion(t *testing.T) {
	repo := newInMemoryJobsRepository()
	orchestrator := NewJobOrchestrator(repo, NewDefaultJobPlanner())
	orchestrator.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	job, err := orchestrator.CreateJob(
		domain.JobTypeStartupScan,
		domain.JobPriorityNormal,
		domain.NewRootScopePayload("/tmp"),
	)
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}

	executor := &skippedExecutor{}
	scheduler := NewJobScheduler(repo, executor, &WorkerContext{})
	scheduler.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	if err := scheduler.RunOnce(); err != nil {
		t.Fatalf("RunOnce failed: %v", err)
	}

	steps, err := repo.GetStepsByJobID(job.ID)
	if err != nil {
		t.Fatalf("GetStepsByJobID failed: %v", err)
	}

	skippedCount := 0
	for _, step := range steps {
		if step.Status == string(domain.StepStatusSkipped) {
			skippedCount++
		}
	}
	if skippedCount != 1 {
		t.Fatalf("expected one skipped step after first tick, got %d", skippedCount)
	}

	if err := scheduler.RunOnce(); err != nil {
		t.Fatalf("second RunOnce failed: %v", err)
	}

	storedJob, _ := repo.GetJobByID(job.ID)
	if storedJob.Status != string(domain.JobStatusCompleted) {
		t.Fatalf("expected completed job after skipped steps, got %s", storedJob.Status)
	}
}
