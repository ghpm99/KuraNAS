package worker

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
)

var ErrStepSkipped = errors.New("step skipped")

type StepExecutor func(step jobs.StepModel) error

type JobScheduler struct {
	repository jobs.RepositoryInterface
	executors  map[StepType]StepExecutor

	queue   chan int
	queued  map[int]struct{}
	stopCh  chan struct{}
	stopWg  sync.WaitGroup
	started bool
	mu      sync.Mutex
	stepSem map[StepType]chan struct{}
}

func NewJobScheduler(repository jobs.RepositoryInterface, executors map[StepType]StepExecutor) *JobScheduler {
	if executors == nil {
		executors = map[StepType]StepExecutor{}
	}

	return &JobScheduler{
		repository: repository,
		executors:  executors,
		queue:      make(chan int, 256),
		queued:     map[int]struct{}{},
		stopCh:     make(chan struct{}),
		stepSem:    buildStepSemaphoreMap(),
	}
}

func (s *JobScheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return
	}

	s.started = true
	s.stopWg.Add(1)
	go s.loop()
}

func (s *JobScheduler) Stop() {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return
	}
	s.started = false
	close(s.stopCh)
	s.mu.Unlock()

	s.stopWg.Wait()
}

func (s *JobScheduler) Enqueue(jobID int) bool {
	if jobID <= 0 {
		return false
	}

	s.mu.Lock()
	if _, exists := s.queued[jobID]; exists {
		s.mu.Unlock()
		return true
	}
	s.queued[jobID] = struct{}{}
	s.mu.Unlock()

	select {
	case s.queue <- jobID:
		return true
	default:
		s.mu.Lock()
		delete(s.queued, jobID)
		s.mu.Unlock()
		return false
	}
}

func (s *JobScheduler) loop() {
	defer s.stopWg.Done()
	pollInterval := time.Duration(config.AppConfig.WorkerSchedulerPollMS) * time.Millisecond
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case jobID := <-s.queue:
			s.mu.Lock()
			delete(s.queued, jobID)
			s.mu.Unlock()
			_ = s.processJob(jobID)
		case <-pollTicker.C:
			s.scheduleQueuedJobs()
		}
	}
}

func (s *JobScheduler) scheduleQueuedJobs() {
	if s == nil || s.repository == nil {
		return
	}

	priorities := []string{
		string(JobPriorityHigh),
		string(JobPriorityNormal),
		string(JobPriorityLow),
	}
	for _, priority := range priorities {
		filter := jobs.JobFilter{}
		filter.Status.Set(string(JobStatusQueued))
		filter.Priority.Set(priority)

		queuedJobs, err := s.repository.ListJobs(filter, 1, 100)
		if err != nil {
			continue
		}

		for _, job := range queuedJobs.Items {
			_ = s.Enqueue(job.ID)
		}
	}
}

func (s *JobScheduler) processJob(jobID int) error {
	if s == nil || s.repository == nil {
		return fmt.Errorf("job scheduler repository is required")
	}

	jobModel, err := s.repository.GetJobByID(jobID)
	if err != nil {
		return fmt.Errorf("load job %d before process: %w", jobID, err)
	}
	if jobModel.CancelRequested || jobModel.Status == string(JobStatusCanceled) {
		endedAt := time.Now()
		_, _ = s.updateJobExecution(jobID, string(JobStatusCanceled), nil, &endedAt, nil, nil)
		_ = s.cancelQueuedSteps(jobID)
		return nil
	}

	now := time.Now()
	if _, err := s.updateJobExecution(jobID, string(JobStatusRunning), &now, nil, nil, nil); err != nil {
		return err
	}

	for {
		steps, err := s.repository.GetStepsByJobID(jobID)
		if err != nil {
			return fmt.Errorf("load steps for job %d: %w", jobID, err)
		}
		if len(steps) == 0 {
			break
		}

		if allStepsTerminal(steps) {
			break
		}

		executedAny := false

		sort.Slice(steps, func(i, j int) bool {
			return steps[i].ID < steps[j].ID
		})

		for _, step := range steps {
			if step.Status != string(StepStatusQueued) {
				continue
			}

			ready, readyErr := stepDependenciesSatisfied(step, steps)
			if readyErr != nil {
				return readyErr
			}
			if !ready {
				continue
			}

			if err := s.executeStep(step); err != nil {
				executedAny = true
				continue
			}
			executedAny = true
		}

		if !executedAny {
			if cancelErr := s.cancelQueuedSteps(jobID); cancelErr != nil {
				return cancelErr
			}
			break
		}
	}

	steps, err := s.repository.GetStepsByJobID(jobID)
	if err != nil {
		return fmt.Errorf("reload steps for job %d: %w", jobID, err)
	}

	status := resolveJobStatus(steps)
	finished := time.Now()
	_, err = s.updateJobExecution(jobID, string(status), nil, &finished, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *JobScheduler) executeStep(step jobs.StepModel) error {
	if step.MaxAttempts <= 0 {
		step.MaxAttempts = 1
	}

	started := time.Now()
	_, err := s.updateStepExecution(step.ID, string(StepStatusRunning), step.Progress, step.Attempts+1, &started, nil, nil)
	if err != nil {
		return err
	}

	executor := s.executors[StepType(step.Type)]
	if executor == nil {
		executor = func(step jobs.StepModel) error {
			return nil
		}
	}

	release := s.acquireStepSemaphore(StepType(step.Type))
	runErr := executor(step)
	if release != nil {
		release()
	}
	ended := time.Now()

	if runErr == nil {
		_, err = s.updateStepExecution(step.ID, string(StepStatusCompleted), 100, step.Attempts+1, nil, &ended, nil)
		return err
	}

	if errors.Is(runErr, ErrStepSkipped) {
		_, err = s.updateStepExecution(step.ID, string(StepStatusSkipped), 100, step.Attempts+1, nil, &ended, nil)
		return err
	}

	if step.Attempts+1 < step.MaxAttempts {
		runErrMessage := runErr.Error()
		_, updateErr := s.updateStepExecution(step.ID, string(StepStatusQueued), step.Progress, step.Attempts+1, nil, nil, &runErrMessage)
		if updateErr != nil {
			return updateErr
		}
		backoff := time.Duration(config.AppConfig.WorkerRetryBackoffMS) * time.Millisecond
		time.Sleep(backoff)
		return nil
	}

	runErrMessage := runErr.Error()
	_, updateErr := s.updateStepExecution(step.ID, string(StepStatusFailed), step.Progress, step.Attempts+1, nil, &ended, &runErrMessage)
	if updateErr != nil {
		return updateErr
	}

	return runErr
}

func (s *JobScheduler) updateJobExecution(jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error) {
	return s.withTx(func(tx *sql.Tx) (bool, error) {
		updated, err := s.repository.UpdateJobExecution(tx, jobID, status, startedAt, endedAt, cancelRequested, lastError)
		if err != nil {
			return false, fmt.Errorf("update job %d execution: %w", jobID, err)
		}
		return updated, nil
	})
}

func (s *JobScheduler) updateStepExecution(stepID int, status string, progress int, attempts int, startedAt *time.Time, endedAt *time.Time, lastError *string) (bool, error) {
	return s.withTx(func(tx *sql.Tx) (bool, error) {
		updated, err := s.repository.UpdateStepExecution(tx, stepID, status, progress, attempts, startedAt, endedAt, lastError)
		if err != nil {
			return false, fmt.Errorf("update step %d execution: %w", stepID, err)
		}
		return updated, nil
	})
}

func (s *JobScheduler) cancelQueuedSteps(jobID int) error {
	steps, err := s.repository.GetStepsByJobID(jobID)
	if err != nil {
		return fmt.Errorf("load steps for cancellation on job %d: %w", jobID, err)
	}

	for _, step := range steps {
		if step.Status != string(StepStatusQueued) {
			continue
		}

		ended := time.Now()
		if _, updateErr := s.updateStepExecution(step.ID, string(StepStatusCanceled), step.Progress, step.Attempts, nil, &ended, nil); updateErr != nil {
			return updateErr
		}
	}

	return nil
}

func (s *JobScheduler) withTx(fn func(*sql.Tx) (bool, error)) (bool, error) {
	dbContext := s.repository.GetDbContext()
	if dbContext == nil {
		return fn(nil)
	}

	var updated bool
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		result, callErr := fn(tx)
		if callErr != nil {
			return callErr
		}
		updated = result
		return nil
	})

	return updated, err
}

func allStepsTerminal(steps []jobs.StepModel) bool {
	for _, step := range steps {
		if !stepStatusTerminal(step.Status) {
			return false
		}
	}

	return true
}

func stepDependenciesSatisfied(step jobs.StepModel, allSteps []jobs.StepModel) (bool, error) {
	if len(step.DependsOn) == 0 {
		return true, nil
	}

	dependencyIDs := []int{}
	if err := json.Unmarshal(step.DependsOn, &dependencyIDs); err != nil {
		return false, fmt.Errorf("invalid dependencies payload for step %d: %w", step.ID, err)
	}

	statusesByID := map[int]string{}
	for _, allStep := range allSteps {
		statusesByID[allStep.ID] = allStep.Status
	}

	for _, dependencyID := range dependencyIDs {
		dependencyStatus, exists := statusesByID[dependencyID]
		if !exists {
			return false, fmt.Errorf("step %d depends on missing step id %d", step.ID, dependencyID)
		}

		if dependencyStatus != string(StepStatusCompleted) && dependencyStatus != string(StepStatusSkipped) {
			return false, nil
		}
	}

	return true, nil
}

func stepStatusTerminal(status string) bool {
	switch StepStatus(status) {
	case StepStatusCompleted, StepStatusFailed, StepStatusCanceled, StepStatusSkipped:
		return true
	default:
		return false
	}
}

func resolveJobStatus(steps []jobs.StepModel) JobStatus {
	if len(steps) == 0 {
		return JobStatusCompleted
	}

	hasFailed := false
	hasSucceeded := false

	for _, step := range steps {
		switch StepStatus(step.Status) {
		case StepStatusFailed:
			hasFailed = true
		case StepStatusCompleted, StepStatusSkipped:
			hasSucceeded = true
		}
	}

	if hasFailed && hasSucceeded {
		return JobStatusPartialFail
	}
	if hasFailed {
		return JobStatusFailed
	}

	return JobStatusCompleted
}

func buildStepSemaphoreMap() map[StepType]chan struct{} {
	checksumLimit := config.AppConfig.WorkerConcurrencyChecksum
	if checksumLimit <= 0 {
		checksumLimit = 3
	}
	metadataLimit := config.AppConfig.WorkerConcurrencyMetadata
	if metadataLimit <= 0 {
		metadataLimit = 3
	}
	thumbnailLimit := config.AppConfig.WorkerConcurrencyThumbnail
	if thumbnailLimit <= 0 {
		thumbnailLimit = 2
	}

	return map[StepType]chan struct{}{
		StepTypeChecksum:  make(chan struct{}, checksumLimit),
		StepTypeMetadata:  make(chan struct{}, metadataLimit),
		StepTypeThumbnail: make(chan struct{}, thumbnailLimit),
	}
}

func (s *JobScheduler) acquireStepSemaphore(stepType StepType) func() {
	sem, exists := s.stepSem[stepType]
	if !exists || sem == nil {
		return nil
	}

	sem <- struct{}{}
	return func() {
		<-sem
	}
}
