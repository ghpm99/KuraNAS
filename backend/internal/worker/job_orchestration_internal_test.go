package worker

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
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

func (r *inMemoryJobsRepository) RequestJobCancel(tx *sql.Tx, id string) (bool, error) {
	job, exists := r.jobsByID[id]
	if !exists || job.CancelRequested {
		return false, nil
	}
	job.CancelRequested = true
	r.jobsByID[id] = job
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
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}

	stepByType := map[string]jobs.StepModel{}
	for _, step := range steps {
		stepByType[step.Type] = step
	}

	scanStep := stepByType[string(domain.StepTypeScanFilesystem)]
	diffStep := stepByType[string(domain.StepTypeDiffAgainstDB)]
	markDeletedStep := stepByType[string(domain.StepTypeMarkDeleted)]
	if scanStep.ID == "" || diffStep.ID == "" || markDeletedStep.ID == "" {
		t.Fatalf("expected scan, diff and mark_deleted steps to exist")
	}

	dependsOn := []string{}
	if err := json.Unmarshal([]byte(diffStep.DependsOnJSON), &dependsOn); err != nil {
		t.Fatalf("failed to parse dependencies: %v", err)
	}
	if len(dependsOn) != 1 || dependsOn[0] != scanStep.ID {
		t.Fatalf("expected diff step to depend on scan step id, got %#v", dependsOn)
	}

	dependsOn = []string{}
	if err := json.Unmarshal([]byte(markDeletedStep.DependsOnJSON), &dependsOn); err != nil {
		t.Fatalf("failed to parse mark_deleted dependencies: %v", err)
	}
	if len(dependsOn) != 1 || dependsOn[0] != diffStep.ID {
		t.Fatalf("expected mark_deleted step to depend on diff step id, got %#v", dependsOn)
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
	if completedCount != 1 || queuedCount != 2 {
		t.Fatalf("expected one completed and two queued after first tick, got completed=%d queued=%d", completedCount, queuedCount)
	}

	if err := scheduler.RunOnce(); err != nil {
		t.Fatalf("second RunOnce failed: %v", err)
	}
	if len(executor.executed) != 2 {
		t.Fatalf("expected two executed steps after second tick, got %d", len(executor.executed))
	}

	if err := scheduler.RunOnce(); err != nil {
		t.Fatalf("third RunOnce failed: %v", err)
	}
	if len(executor.executed) != 3 {
		t.Fatalf("expected three executed steps after third tick, got %d", len(executor.executed))
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

	if err := scheduler.RunOnce(); err != nil {
		t.Fatalf("third RunOnce failed: %v", err)
	}

	storedJob, _ := repo.GetJobByID(job.ID)
	if storedJob.Status != string(domain.JobStatusCompleted) {
		t.Fatalf("expected completed job after skipped steps, got %s", storedJob.Status)
	}
}

func TestExecuteMarkDeletedStepMarksMissingActiveFiles(t *testing.T) {
	tmpDir := t.TempDir()
	presentPath := filepath.Join(tmpDir, "present.txt")
	if err := os.WriteFile(presentPath, []byte("ok"), 0644); err != nil {
		t.Fatalf("failed to create present file: %v", err)
	}

	missingPath := filepath.Join(tmpDir, "missing.txt")
	updatedIDs := []int{}
	service := &workerFilesServiceMock{
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					{ID: 1, Name: "missing.txt", Path: missingPath},
					{ID: 2, Name: "present.txt", Path: presentPath},
				},
				Pagination: utils.Pagination{Page: 1, PageSize: pageSize, HasNext: false},
			}, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			if !file.DeletedAt.HasValue {
				t.Fatalf("expected deleted_at to be set for file id=%d", file.ID)
			}
			updatedIDs = append(updatedIDs, file.ID)
			return true, nil
		},
	}

	executor := NewDefaultStepExecutor()
	err := executor.ExecuteStep(domain.Step{
		ID:    "mark-step",
		JobID: "job-1",
		Type:  domain.StepTypeMarkDeleted,
		Scope: domain.NewRootScopePayload(tmpDir),
	}, &WorkerContext{FilesService: service})
	if err != nil {
		t.Fatalf("mark_deleted execution failed: %v", err)
	}

	if len(updatedIDs) != 1 || updatedIDs[0] != 1 {
		t.Fatalf("expected only missing file to be marked deleted, got ids=%v", updatedIDs)
	}
}

func TestExecuteDiffAgainstDBStepReactivatesDeletedRecord(t *testing.T) {
	tmpDir := t.TempDir()
	reactivatedPath := filepath.Join(tmpDir, "reactivated.txt")
	if err := os.WriteFile(reactivatedPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create reactivated file: %v", err)
	}
	info, err := os.Stat(reactivatedPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	repo := newInMemoryJobsRepository()
	orchestrator := NewJobOrchestrator(repo, NewDefaultJobPlanner())
	orchestrator.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	updateCalls := 0
	service := &workerFilesServiceMock{
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					{
						ID:        77,
						Name:      info.Name(),
						Path:      reactivatedPath,
						UpdatedAt: info.ModTime(),
						CreatedAt: info.ModTime(),
						Size:      info.Size(),
						Type:      files.File,
						Format:    filepath.Ext(info.Name()),
						DeletedAt: utils.Optional[time.Time]{HasValue: true, Value: time.Now().UTC()},
					},
				},
				Pagination: utils.Pagination{Page: 1, PageSize: pageSize, HasNext: false},
			}, nil
		},
		getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
			return files.FileDto{
				ID:        77,
				Name:      name,
				Path:      path,
				UpdatedAt: info.ModTime(),
				CreatedAt: info.ModTime(),
				Size:      info.Size(),
				Type:      files.File,
				Format:    filepath.Ext(name),
				DeletedAt: utils.Optional[time.Time]{HasValue: true, Value: time.Now().UTC()},
			}, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			updateCalls++
			return true, nil
		},
	}

	executor := NewDefaultStepExecutor()
	err = executor.ExecuteStep(domain.Step{
		ID:    "diff-step",
		JobID: "job-2",
		Type:  domain.StepTypeDiffAgainstDB,
		Scope: domain.NewRootScopePayload(tmpDir),
	}, &WorkerContext{
		FilesService: service,
		Orchestrator: orchestrator,
	})
	if err != nil {
		t.Fatalf("diff_against_db execution failed: %v", err)
	}

	if updateCalls != 0 {
		t.Fatalf("expected diff step to enqueue follow-up job instead of mutating records directly, got update_calls=%d", updateCalls)
	}
	if len(repo.jobsByID) == 0 {
		t.Fatalf("expected incremental child jobs to be enqueued")
	}
	foundReactivatedPath := false
	for _, job := range repo.jobsByID {
		if job.ScopeJSON == "" {
			continue
		}
		scope := domain.ScopePayload{}
		if err := json.Unmarshal([]byte(job.ScopeJSON), &scope); err != nil {
			continue
		}
		if scope.File != nil && scope.File.Path == reactivatedPath {
			foundReactivatedPath = true
			break
		}
	}
	if !foundReactivatedPath {
		t.Fatalf("expected one child job scoped to reactivated file path")
	}
}

func TestExecuteDiffAgainstDBStepSkipsUnchangedEntries(t *testing.T) {
	tmpDir := t.TempDir()
	unchangedPath := filepath.Join(tmpDir, "unchanged.txt")
	if err := os.WriteFile(unchangedPath, []byte("same"), 0644); err != nil {
		t.Fatalf("failed to create unchanged file: %v", err)
	}

	rootInfo, err := os.Stat(tmpDir)
	if err != nil {
		t.Fatalf("stat root failed: %v", err)
	}
	info, err := os.Stat(unchangedPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	createCalls := 0
	updateCalls := 0
	repo := newInMemoryJobsRepository()
	orchestrator := NewJobOrchestrator(repo, NewDefaultJobPlanner())
	orchestrator.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	service := &workerFilesServiceMock{
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					{
						ID:        9,
						Name:      rootInfo.Name(),
						Path:      tmpDir,
						UpdatedAt: rootInfo.ModTime(),
						CreatedAt: rootInfo.ModTime(),
						Size:      rootInfo.Size(),
						Type:      files.Directory,
						DeletedAt: utils.Optional[time.Time]{HasValue: false},
					},
					{
						ID:        10,
						Name:      info.Name(),
						Path:      unchangedPath,
						UpdatedAt: info.ModTime(),
						CreatedAt: info.ModTime(),
						Size:      info.Size(),
						Type:      files.File,
						Format:    filepath.Ext(info.Name()),
						DeletedAt: utils.Optional[time.Time]{HasValue: false},
					},
				},
				Pagination: utils.Pagination{Page: 1, PageSize: pageSize, HasNext: false},
			}, nil
		},
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			createCalls++
			return fileDto, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			updateCalls++
			return true, nil
		},
	}

	executor := NewDefaultStepExecutor()
	err = executor.ExecuteStep(domain.Step{
		ID:    "diff-step",
		JobID: "job-unchanged",
		Type:  domain.StepTypeDiffAgainstDB,
		Scope: domain.NewRootScopePayload(tmpDir),
	}, &WorkerContext{
		FilesService: service,
		Orchestrator: orchestrator,
	})
	if err != nil {
		t.Fatalf("diff_against_db execution failed: %v", err)
	}

	if createCalls != 0 || updateCalls != 0 {
		t.Fatalf("expected unchanged entry to skip fan-out, create=%d update=%d", createCalls, updateCalls)
	}
	if len(repo.jobsByID) != 0 {
		t.Fatalf("expected unchanged entries to avoid child-job fan-out, got %d jobs", len(repo.jobsByID))
	}
}

func TestExecuteMarkDeletedStepRespectsScopePath(t *testing.T) {
	tmpDir := t.TempDir()
	scopedDir := filepath.Join(tmpDir, "scoped")
	otherDir := filepath.Join(tmpDir, "other")
	if err := os.MkdirAll(scopedDir, 0755); err != nil {
		t.Fatalf("failed to create scoped dir: %v", err)
	}
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatalf("failed to create other dir: %v", err)
	}

	existingScopedFile := filepath.Join(scopedDir, "existing.txt")
	if err := os.WriteFile(existingScopedFile, []byte("ok"), 0644); err != nil {
		t.Fatalf("failed to create scoped file: %v", err)
	}

	updatedIDs := []int{}
	service := &workerFilesServiceMock{
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					{ID: 1, Path: existingScopedFile, Type: files.File, DeletedAt: utils.Optional[time.Time]{HasValue: false}},
					{ID: 2, Path: filepath.Join(scopedDir, "missing.txt"), Type: files.File, DeletedAt: utils.Optional[time.Time]{HasValue: false}},
					{ID: 3, Path: filepath.Join(otherDir, "outside-missing.txt"), Type: files.File, DeletedAt: utils.Optional[time.Time]{HasValue: false}},
				},
				Pagination: utils.Pagination{Page: 1, PageSize: pageSize, HasNext: false},
			}, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			if file.DeletedAt.HasValue {
				updatedIDs = append(updatedIDs, file.ID)
			}
			return true, nil
		},
	}

	executor := NewDefaultStepExecutor()
	err := executor.ExecuteStep(domain.Step{
		ID:    "mark-deleted-scoped",
		JobID: "job-mark-deleted-scoped",
		Type:  domain.StepTypeMarkDeleted,
		Scope: domain.NewPathScopePayload(scopedDir),
	}, &WorkerContext{FilesService: service})
	if err != nil {
		t.Fatalf("mark_deleted execution failed: %v", err)
	}

	if len(updatedIDs) != 1 || updatedIDs[0] != 2 {
		t.Fatalf("expected only scoped missing entry to be marked deleted, got ids=%v", updatedIDs)
	}
}

func TestSchedulerRecoversInterruptedRunningStep(t *testing.T) {
	repo := newInMemoryJobsRepository()
	startedAt := time.Now().UTC().Add(-2 * time.Minute)
	repo.jobsByID["job-running"] = jobs.JobModel{
		ID:        "job-running",
		Type:      string(domain.JobTypeStartupScan),
		Status:    string(domain.JobStatusRunning),
		Priority:  int(domain.JobPriorityLow),
		CreatedAt: startedAt,
	}
	repo.stepsByID["step-running"] = jobs.StepModel{
		ID:          "step-running",
		JobID:       "job-running",
		Type:        string(domain.StepTypeScanFilesystem),
		Status:      string(domain.StepStatusRunning),
		Attempts:    1,
		MaxAttempts: 3,
		CreatedAt:   startedAt,
		StartedAt:   sql.NullTime{Valid: true, Time: startedAt},
	}

	scheduler := NewJobScheduler(repo, &recordingExecutor{}, &WorkerContext{})
	scheduler.runInTx = func(fn func(*sql.Tx) error) error { return fn(nil) }

	schedulableJobs, err := scheduler.listSchedulableJobs()
	if err != nil {
		t.Fatalf("listSchedulableJobs failed: %v", err)
	}
	if err := scheduler.recoverInterruptedRunningSteps(schedulableJobs); err != nil {
		t.Fatalf("recoverInterruptedRunningSteps failed: %v", err)
	}

	job := repo.jobsByID["job-running"]
	if job.Status != string(domain.JobStatusQueued) {
		t.Fatalf("expected running job to be requeued after restart recovery, got %s", job.Status)
	}

	step := repo.stepsByID["step-running"]
	if step.Status != string(domain.StepStatusQueued) {
		t.Fatalf("expected running step to be requeued after restart recovery, got %s", step.Status)
	}
	if step.LastError == "" {
		t.Fatalf("expected recovery message to be persisted")
	}
}

func TestCalculateJobStatusFromStepsReturnsPartialFail(t *testing.T) {
	steps := []schedulerStep{
		{Step: domain.Step{ID: "failed", Status: domain.StepStatusFailed}},
		{Step: domain.Step{ID: "completed", Status: domain.StepStatusCompleted}},
	}

	status := calculateJobStatusFromSteps(steps)
	if status != domain.JobStatusPartialFail {
		t.Fatalf("expected partial_fail, got %s", status)
	}
}
