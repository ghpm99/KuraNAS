package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	jobsapi "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

// TestCreateJobSkipsDuplicatePendingPath verifies idempotency: a second job for
// a file that already has a pending job is skipped, so ~30k files cannot explode
// into millions of duplicate jobs.
func TestCreateJobSkipsDuplicatePendingPath(t *testing.T) {
	repo := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repo, nil)

	plan := PlannedJob{
		Type:     JobTypeFSEvent,
		Priority: JobPriorityLow,
		Scope:    JobScope{Path: "/data/a.jpg"},
		Steps: []PlannedStep{
			{Key: "persist", Type: StepTypePersist, MaxAttempts: 1},
		},
	}

	id1, err := orchestrator.CreateJob(plan)
	if err != nil {
		t.Fatalf("first CreateJob error: %v", err)
	}
	if id1 == 0 {
		t.Fatalf("expected first job to be created")
	}

	id2, err := orchestrator.CreateJob(plan)
	if err != nil {
		t.Fatalf("second CreateJob error: %v", err)
	}
	if id2 != 0 {
		t.Fatalf("expected duplicate to be skipped, got id %d", id2)
	}

	if len(repo.jobs) != 1 {
		t.Fatalf("expected exactly 1 job, got %d", len(repo.jobs))
	}

	// A different path is not deduped.
	other := plan
	other.Scope = JobScope{Path: "/data/b.jpg"}
	id3, err := orchestrator.CreateJob(other)
	if err != nil {
		t.Fatalf("CreateJob for other path error: %v", err)
	}
	if id3 == 0 {
		t.Fatalf("expected job for a different path to be created")
	}
}

// TestRecoverInterruptedWork verifies that jobs/steps stranded in 'running' are
// reset to 'queued' on recovery so the scheduler can reprocess them.
func TestRecoverInterruptedWork(t *testing.T) {
	repo := newFakeJobsRepository()
	scheduler := NewJobScheduler(repo, map[StepType]StepExecutor{})

	repo.jobs[1] = jobsapi.JobModel{ID: 1, Status: string(JobStatusRunning)}
	repo.jobs[2] = jobsapi.JobModel{ID: 2, Status: string(JobStatusCompleted)}
	repo.steps[10] = jobsapi.StepModel{ID: 10, JobID: 1, Status: string(StepStatusRunning)}
	repo.steps[11] = jobsapi.StepModel{ID: 11, JobID: 1, Status: string(StepStatusCompleted)}

	jobsReset, stepsReset, err := scheduler.RecoverInterruptedWork()
	if err != nil {
		t.Fatalf("RecoverInterruptedWork returned error: %v", err)
	}
	if jobsReset != 1 {
		t.Fatalf("expected 1 job reset, got %d", jobsReset)
	}
	if stepsReset != 1 {
		t.Fatalf("expected 1 step reset, got %d", stepsReset)
	}

	if repo.jobs[1].Status != string(JobStatusQueued) {
		t.Fatalf("expected running job requeued, got %s", repo.jobs[1].Status)
	}
	if repo.jobs[2].Status != string(JobStatusCompleted) {
		t.Fatalf("completed job must be untouched, got %s", repo.jobs[2].Status)
	}
	if repo.steps[10].Status != string(StepStatusQueued) {
		t.Fatalf("expected running step requeued, got %s", repo.steps[10].Status)
	}
	if repo.steps[11].Status != string(StepStatusCompleted) {
		t.Fatalf("completed step must be untouched, got %s", repo.steps[11].Status)
	}
}

// TestProcessJobDefersOnTimeout verifies that a step timing out does not fail
// the job: the step returns to the queue with timeout_count bumped and the job
// is sent to the back of the line (queued + next_attempt_at set), never failed.
func TestProcessJobDefersOnTimeout(t *testing.T) {
	repo := newFakeJobsRepository()
	scheduler := NewJobScheduler(repo, map[StepType]StepExecutor{
		StepTypeScanFilesystem: func(step jobsapi.StepModel) error {
			return context.DeadlineExceeded
		},
	})

	orchestrator := NewJobOrchestrator(repo, scheduler)
	jobID, err := orchestrator.CreateJob(PlannedJob{
		Type:     JobTypeStartupScan,
		Priority: JobPriorityLow,
		Scope:    JobScope{Root: "/data"},
		Steps: []PlannedStep{
			{Key: "scan", Type: StepTypeScanFilesystem, MaxAttempts: 1},
		},
	})
	if err != nil {
		t.Fatalf("unexpected create job error: %v", err)
	}

	if err := scheduler.processJob(jobID); err != nil {
		t.Fatalf("processJob returned error: %v", err)
	}

	job, err := repo.GetJobByID(jobID)
	if err != nil {
		t.Fatalf("GetJobByID returned error: %v", err)
	}
	if job.Status != string(JobStatusQueued) {
		t.Fatalf("expected job requeued, got %s", job.Status)
	}
	if job.NextAttemptAt == nil {
		t.Fatalf("expected next_attempt_at set to send job to back of queue")
	}

	steps, err := repo.GetStepsByJobID(jobID)
	if err != nil {
		t.Fatalf("GetStepsByJobID returned error: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}
	if steps[0].Status != string(StepStatusQueued) {
		t.Fatalf("expected step requeued, got %s", steps[0].Status)
	}
	if steps[0].TimeoutCount != 1 {
		t.Fatalf("expected timeout_count=1, got %d", steps[0].TimeoutCount)
	}
}

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

func (r *fakeJobsRepository) DeferStepForTimeout(tx *sql.Tx, stepID int, attempts int, lastError string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	step, exists := r.steps[stepID]
	if !exists {
		return false, nil
	}

	step.Status = string(StepStatusQueued)
	step.Attempts = attempts
	step.TimeoutCount++
	step.StartedAt = nil
	step.LastError = lastError

	r.steps[stepID] = step
	return true, nil
}

func (r *fakeJobsRepository) HasPendingJobForPath(path string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if path == "" {
		return false, nil
	}
	for _, job := range r.jobs {
		if job.Status != string(JobStatusQueued) && job.Status != string(JobStatusRunning) {
			continue
		}
		var scope JobScope
		if len(job.Scope) > 0 {
			_ = json.Unmarshal(job.Scope, &scope)
		}
		if scope.Path == path {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeJobsRepository) RecoverInterruptedWork(tx *sql.Tx) (int64, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var jobsReset, stepsReset int64
	for id, step := range r.steps {
		if step.Status == string(StepStatusRunning) {
			step.Status = string(StepStatusQueued)
			step.StartedAt = nil
			r.steps[id] = step
			stepsReset++
		}
	}
	for id, job := range r.jobs {
		if job.Status == string(JobStatusRunning) {
			job.Status = string(JobStatusQueued)
			job.StartedAt = nil
			r.jobs[id] = job
			jobsReset++
		}
	}
	return jobsReset, stepsReset, nil
}

func (r *fakeJobsRepository) RequeueJob(tx *sql.Tx, jobID int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, exists := r.jobs[jobID]
	if !exists {
		return false, nil
	}

	job.Status = string(JobStatusQueued)
	job.StartedAt = nil
	job.EndedAt = nil
	now := time.Now()
	job.NextAttemptAt = &now

	r.jobs[jobID] = job
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

func TestJobSchedulerFailsStepWhenExecutorMissing(t *testing.T) {
	fakeRepository := newFakeJobsRepository()
	scheduler := NewJobScheduler(fakeRepository, map[StepType]StepExecutor{})
	orchestrator := NewJobOrchestrator(fakeRepository, scheduler)

	jobID, err := orchestrator.CreateJob(PlannedJob{
		Type:     JobTypeFSEvent,
		Priority: JobPriorityNormal,
		Steps: []PlannedStep{
			{
				Key:  "checksum",
				Type: StepTypeChecksum,
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
	if job.Status != string(JobStatusFailed) {
		t.Fatalf("expected failed job, got %s", job.Status)
	}

	steps, err := fakeRepository.GetStepsByJobID(jobID)
	if err != nil {
		t.Fatalf("GetStepsByJobID returned error: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected one step, got %d", len(steps))
	}
	if steps[0].Status != string(StepStatusFailed) {
		t.Fatalf("expected failed step, got %s", steps[0].Status)
	}
	if !strings.Contains(steps[0].LastError, "not configured") {
		t.Fatalf("expected missing executor error in step, got %q", steps[0].LastError)
	}
}

func TestJobSchedulerKeepsCanceledStatusWhenCancellationIsRequestedDuringExecution(t *testing.T) {
	fakeRepository := newFakeJobsRepository()
	jobID := 0

	scheduler := NewJobScheduler(fakeRepository, map[StepType]StepExecutor{
		StepTypeScanFilesystem: func(step jobsapi.StepModel) error {
			fakeRepository.mu.Lock()
			job := fakeRepository.jobs[jobID]
			job.CancelRequested = true
			fakeRepository.jobs[jobID] = job
			fakeRepository.mu.Unlock()
			return nil
		},
		StepTypeChecksum: func(step jobsapi.StepModel) error {
			return nil
		},
	})

	orchestrator := NewJobOrchestrator(fakeRepository, scheduler)
	createdJobID, err := orchestrator.CreateJob(PlannedJob{
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
	jobID = createdJobID

	if err := scheduler.processJob(jobID); err != nil {
		t.Fatalf("unexpected processJob error: %v", err)
	}

	job, err := fakeRepository.GetJobByID(jobID)
	if err != nil {
		t.Fatalf("GetJobByID returned error: %v", err)
	}
	if job.Status != string(JobStatusCanceled) {
		t.Fatalf("expected canceled job, got %s", job.Status)
	}

	steps, err := fakeRepository.GetStepsByJobID(jobID)
	if err != nil {
		t.Fatalf("GetStepsByJobID returned error: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("expected two steps, got %d", len(steps))
	}
	if steps[0].Status != string(StepStatusCompleted) {
		t.Fatalf("expected first step completed, got %s", steps[0].Status)
	}
	if steps[1].Status != string(StepStatusCanceled) {
		t.Fatalf("expected queued dependent step canceled, got %s", steps[1].Status)
	}
}
