package worker

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

	jobSem chan struct{}
	jobWg  sync.WaitGroup
}

func NewJobScheduler(repository jobs.RepositoryInterface, executors map[StepType]StepExecutor) *JobScheduler {
	if executors == nil {
		executors = map[StepType]StepExecutor{}
	}

	maxJobs := config.AppConfig.WorkerMaxConcurrentJobs
	if maxJobs <= 0 {
		maxJobs = 4
	}

	return &JobScheduler{
		repository: repository,
		executors:  executors,
		queue:      make(chan int, 1024),
		queued:     map[int]struct{}{},
		stopCh:     make(chan struct{}),
		stepSem:    buildStepSemaphoreMap(),
		jobSem:     make(chan struct{}, maxJobs),
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
	s.jobWg.Wait()
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

			s.jobSem <- struct{}{}
			s.jobWg.Add(1)
			go func(id int) {
				defer s.jobWg.Done()
				defer func() { <-s.jobSem }()
				if err := s.processJob(id); err != nil {
					log.Printf("[job=%d] processJob error: %v\n", id, err)
				}
			}(jobID)
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

		page := 1
		for {
			// Check if semaphore is full before loading more jobs
			if len(s.jobSem) >= cap(s.jobSem) {
				return
			}

			queuedJobs, err := s.repository.ListJobs(filter, page, 100)
			if err != nil {
				break
			}

			for _, job := range queuedJobs.Items {
				_ = s.Enqueue(job.ID)
			}

			if !queuedJobs.Pagination.HasNext {
				break
			}
			page++
		}
	}
}

func (s *JobScheduler) processJob(jobID int) error {
	if s == nil || s.repository == nil {
		return fmt.Errorf("job scheduler repository is required")
	}

	canceled, err := s.cancelIfRequested(jobID)
	if err != nil {
		return err
	}
	if canceled {
		return nil
	}

	now := time.Now()
	if _, err := s.updateJobExecution(jobID, string(JobStatusRunning), &now, nil, nil, nil); err != nil {
		return err
	}

	var firstStepErr error

	for {
		canceled, cancelErr := s.cancelIfRequested(jobID)
		if cancelErr != nil {
			return cancelErr
		}
		if canceled {
			return nil
		}

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

		sort.Slice(steps, func(i, j int) bool {
			return steps[i].ID < steps[j].ID
		})

		// Pre-parse all step dependencies (Fix 13)
		parsedDeps := parseAllStepDependencies(steps)
		statusByID := buildStatusMap(steps)

		readySteps := []jobs.StepModel{}
		for _, step := range steps {
			if step.Status != string(StepStatusQueued) {
				continue
			}
			if stepDependenciesSatisfied(step.ID, parsedDeps, statusByID) {
				readySteps = append(readySteps, step)
			}
		}

		if len(readySteps) == 0 {
			if cancelErr := s.cancelQueuedSteps(jobID); cancelErr != nil {
				return cancelErr
			}
			break
		}

		if len(readySteps) == 1 {
			if err := s.executeStep(readySteps[0]); err != nil {
				log.Printf("[job=%d step=%d type=%s] step error: %v\n", jobID, readySteps[0].ID, readySteps[0].Type, err)
				if firstStepErr == nil {
					firstStepErr = err
				}
			}
		} else {
			var wg sync.WaitGroup
			var errMu sync.Mutex
			wg.Add(len(readySteps))
			for _, rs := range readySteps {
				go func(step jobs.StepModel) {
					defer wg.Done()
					if err := s.executeStep(step); err != nil {
						log.Printf("[job=%d step=%d type=%s] step error: %v\n", jobID, step.ID, step.Type, err)
						errMu.Lock()
						if firstStepErr == nil {
							firstStepErr = err
						}
						errMu.Unlock()
					}
				}(rs)
			}
			wg.Wait()
		}
	}

	canceled, err = s.cancelIfRequested(jobID)
	if err != nil {
		return err
	}
	if canceled {
		return nil
	}

	finalSteps, err := s.repository.GetStepsByJobID(jobID)
	if err != nil {
		return fmt.Errorf("reload steps for job %d: %w", jobID, err)
	}

	status := resolveJobStatus(finalSteps)
	finished := time.Now()

	var lastError *string
	if firstStepErr != nil && status != JobStatusCompleted {
		errMsg := firstStepErr.Error()
		lastError = &errMsg
	}

	_, err = s.updateJobExecution(jobID, string(status), nil, &finished, nil, lastError)
	if err != nil {
		return err
	}

	return nil
}

func (s *JobScheduler) executeStep(step jobs.StepModel) error {
	if step.MaxAttempts <= 0 {
		step.MaxAttempts = 1
	}

	executor := s.executors[StepType(step.Type)]

	// Acquire semaphore before marking as running (Fix 2)
	var release func()
	if executor != nil {
		release = s.acquireStepSemaphore(StepType(step.Type))
	}

	started := time.Now()
	_, err := s.updateStepExecution(step.ID, string(StepStatusRunning), step.Progress, step.Attempts+1, &started, nil, nil)
	if err != nil {
		if release != nil {
			release()
		}
		return err
	}

	var runErr error
	if executor == nil {
		runErr = fmt.Errorf("step executor is not configured for type %q", step.Type)
	} else {
		runErr = executor(step)
		if release != nil {
			release()
		}
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
		// Apply retry backoff (Fix 3)
		backoff := time.Duration(config.AppConfig.WorkerRetryBackoffMS) * time.Millisecond * time.Duration(step.Attempts+1)
		if backoff > 0 {
			time.Sleep(backoff)
		}
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

func (s *JobScheduler) cancelIfRequested(jobID int) (bool, error) {
	jobModel, err := s.repository.GetJobByID(jobID)
	if err != nil {
		return false, fmt.Errorf("load job %d before cancellation check: %w", jobID, err)
	}
	if !jobModel.CancelRequested && jobModel.Status != string(JobStatusCanceled) {
		return false, nil
	}

	endedAt := time.Now()
	if _, err := s.updateJobExecution(jobID, string(JobStatusCanceled), nil, &endedAt, nil, nil); err != nil {
		return false, err
	}
	if err := s.cancelQueuedSteps(jobID); err != nil {
		return false, err
	}
	return true, nil
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

// parseAllStepDependencies pre-parses all step DependsOn JSON into a map (Fix 13).
func parseAllStepDependencies(steps []jobs.StepModel) map[int][]int {
	result := make(map[int][]int, len(steps))
	for _, step := range steps {
		if len(step.DependsOn) == 0 {
			result[step.ID] = nil
			continue
		}
		var deps []int
		if err := json.Unmarshal(step.DependsOn, &deps); err != nil {
			result[step.ID] = nil
			continue
		}
		result[step.ID] = deps
	}
	return result
}

func buildStatusMap(steps []jobs.StepModel) map[int]string {
	m := make(map[int]string, len(steps))
	for _, step := range steps {
		m[step.ID] = step.Status
	}
	return m
}

func stepDependenciesSatisfied(stepID int, parsedDeps map[int][]int, statusByID map[int]string) bool {
	deps := parsedDeps[stepID]
	if len(deps) == 0 {
		return true
	}

	for _, depID := range deps {
		depStatus, exists := statusByID[depID]
		if !exists {
			return false
		}
		if depStatus != string(StepStatusCompleted) && depStatus != string(StepStatusSkipped) {
			return false
		}
	}

	return true
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
	allCanceled := true

	for _, step := range steps {
		switch StepStatus(step.Status) {
		case StepStatusFailed:
			hasFailed = true
			allCanceled = false
		case StepStatusCompleted, StepStatusSkipped:
			hasSucceeded = true
			allCanceled = false
		case StepStatusCanceled:
			// remains allCanceled = true
		default:
			allCanceled = false
		}
	}

	if allCanceled {
		return JobStatusCanceled
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
