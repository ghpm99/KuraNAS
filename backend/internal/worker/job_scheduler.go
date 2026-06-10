package worker

import (
	"nas-go/api/internal/worker/job"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/applog"
)

var ErrStepSkipped = errors.New("step skipped")

// stallHeartbeats is how many consecutive all-slots-busy / nothing-finishing
// heartbeats must pass before the scheduler is declared stalled.
const stallHeartbeats = 3

// errStepDeferred signals that a step timed out and the whole job was sent to
// the back of the queue. It is an internal control value, not a failure.
var errStepDeferred = errors.New("step deferred to back of queue")

type StepExecutor func(step jobs.StepModel) error

type JobScheduler struct {
	repository jobs.RepositoryInterface
	executors  map[job.StepType]StepExecutor

	queue   chan int
	queued  map[int]struct{}
	stopCh  chan struct{}
	stopWg  sync.WaitGroup
	started bool
	mu      sync.Mutex
	stepSem map[job.StepType]chan struct{}

	jobSem chan struct{}
	jobWg  sync.WaitGroup

	// onJobFinished, when set, is called once a job reaches a terminal status in
	// the normal finish path. It lets the composition root record audit/health
	// events and emit notifications without the scheduler depending on those
	// packages. Set it before Start(); it is read from job goroutines.
	onJobFinished func(jobID int, jobType string, status job.JobStatus)

	// onStall, when set, is called once per stall episode when every job slot
	// has been busy with no job finishing across several heartbeats — the
	// silent-freeze signature. Set it before Start().
	onStall func(runningJobs int)

	// finishedJobs counts terminal job finishes; the heartbeat watches it to
	// tell "busy but progressing" from "wedged". Accessed atomically.
	finishedJobs int64
}

// SetOnJobFinished registers a callback invoked when a job finishes with a
// terminal status. Call before Start().
func (s *JobScheduler) SetOnJobFinished(fn func(jobID int, jobType string, status job.JobStatus)) {
	s.onJobFinished = fn
}

// SetOnStall registers a callback invoked when the scheduler looks frozen (all
// slots busy, nothing finishing). Call before Start().
func (s *JobScheduler) SetOnStall(fn func(runningJobs int)) {
	s.onStall = fn
}

func NewJobScheduler(repository jobs.RepositoryInterface, executors map[job.StepType]StepExecutor) *JobScheduler {
	if executors == nil {
		executors = map[job.StepType]StepExecutor{}
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
	s.stopWg.Add(2)
	go s.runLoop()
	go s.runHeartbeat()
}

// runHeartbeat emits a periodic liveness record so a frozen scheduler is
// visible in the forensic log: if the heartbeat stops, the loop is stuck.
// Counts are read from the scheduler's in-memory state (cheap, no DB) and show
// running jobs, free slots and queue backlog — enough to tell "idle" from
// "wedged with work piled up".
func (s *JobScheduler) runHeartbeat() {
	defer s.stopWg.Done()

	interval := time.Duration(config.AppConfig.WorkerHeartbeatSeconds) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var stallTicks int
	lastFinished := atomic.LoadInt64(&s.finishedJobs)

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			tracked := len(s.queued)
			s.mu.Unlock()

			running := len(s.jobSem)
			slots := cap(s.jobSem)
			finished := atomic.LoadInt64(&s.finishedJobs)

			applog.Info("scheduler heartbeat",
				"running", running,
				"slots", slots,
				"queue_depth", len(s.queue),
				"tracked", tracked,
			)

			// Stall = every slot busy and no job finished since last tick. After
			// stallHeartbeats consecutive such ticks, raise the alarm once.
			if slots > 0 && running >= slots && finished == lastFinished {
				stallTicks++
			} else {
				stallTicks = 0
			}
			lastFinished = finished

			if stallTicks == stallHeartbeats {
				applog.Error("scheduler appears stalled",
					"running", running, "slots", slots, "queue_depth", len(s.queue),
					"heartbeats", stallTicks)
				if s.onStall != nil {
					s.onStall(running)
				}
			}
		}
	}
}

// runLoop owns the scheduler loop's lifetime: it runs loop() under panic
// recovery and restarts it if it panics, so a single bad iteration cannot
// permanently freeze job scheduling. It returns (releasing stopWg) only on a
// normal stop via stopCh.
func (s *JobScheduler) runLoop() {
	defer s.stopWg.Done()
	for {
		if !applog.RunGuarded("scheduler-loop", s.loop) {
			return
		}
		select {
		case <-s.stopCh:
			return
		default:
			applog.Warn("scheduler loop restarting after panic")
		}
	}
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

			// Acquire a job slot, but stay responsive to Stop() while all slots
			// are busy: a stuck job must never make the loop deaf to shutdown.
			select {
			case s.jobSem <- struct{}{}:
			case <-s.stopCh:
				return
			}
			s.jobWg.Add(1)
			go func(id int) {
				defer s.jobWg.Done()
				defer func() { <-s.jobSem }()
				// Recover so a panic inside a step executor is logged with its
				// stack and only fails this job, instead of crashing the whole
				// process (which would also leak the jobSem slot via os.Exit).
				applog.Recover(fmt.Sprintf("job-%d", id), func() {
					if err := s.processJob(id); err != nil {
						applog.Error("processJob error", "job_id", id, "error", err.Error())
					}
				})
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
		string(job.JobPriorityHigh),
		string(job.JobPriorityNormal),
		string(job.JobPriorityLow),
	}
	for _, priority := range priorities {
		filter := jobs.JobFilter{}
		filter.Status.Set(string(job.JobStatusQueued))
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
	if _, err := s.updateJobExecution(jobID, string(job.JobStatusRunning), &now, nil, nil, nil); err != nil {
		return err
	}

	jobType := ""
	if jobModel, jobErr := s.repository.GetJobByID(jobID); jobErr == nil {
		jobType = jobModel.Type
	}
	applog.Info("job started", "job_id", jobID, "type", jobType)

	var firstStepErr error
	deferred := false

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
			if step.Status != string(job.StepStatusQueued) {
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
				if errors.Is(err, errStepDeferred) {
					deferred = true
				} else if firstStepErr == nil {
					// executeStep already logged the failure with full context.
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
						if errors.Is(err, errStepDeferred) {
							errMu.Lock()
							deferred = true
							errMu.Unlock()
							return
						}
						// executeStep already logged the failure with full context.
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

		if deferred {
			break
		}
	}

	if deferred {
		if _, err := s.requeueJob(jobID); err != nil {
			return err
		}
		applog.Warn("job deferred to back of queue after step timeout",
			"job_id", jobID, "type", jobType, "duration_ms", time.Since(now).Milliseconds())
		return nil
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
	if firstStepErr != nil && status != job.JobStatusCompleted {
		errMsg := firstStepErr.Error()
		lastError = &errMsg
	}

	_, err = s.updateJobExecution(jobID, string(status), nil, &finished, nil, lastError)
	if err != nil {
		return err
	}

	logJobFinished(jobID, jobType, status, finished.Sub(now), finalSteps, lastError)

	atomic.AddInt64(&s.finishedJobs, 1)
	if s.onJobFinished != nil {
		s.onJobFinished(jobID, jobType, status)
	}

	return nil
}

// logJobFinished records the job outcome forensically: a failed or partially
// failed job lands at ERROR (so it surfaces on a grep for failures), a healthy
// one at INFO, both carrying duration and a per-status step breakdown.
func logJobFinished(jobID int, jobType string, status job.JobStatus, duration time.Duration, steps []jobs.StepModel, lastError *string) {
	completed, failed, skipped, canceled := 0, 0, 0, 0
	for _, step := range steps {
		switch job.StepStatus(step.Status) {
		case job.StepStatusCompleted:
			completed++
		case job.StepStatusFailed:
			failed++
		case job.StepStatusSkipped:
			skipped++
		case job.StepStatusCanceled:
			canceled++
		}
	}

	args := []any{
		"job_id", jobID,
		"type", jobType,
		"status", string(status),
		"duration_ms", duration.Milliseconds(),
		"steps_completed", completed,
		"steps_failed", failed,
		"steps_skipped", skipped,
		"steps_canceled", canceled,
	}
	if lastError != nil {
		args = append(args, "error", *lastError)
	}

	if status == job.JobStatusFailed || status == job.JobStatusPartialFail {
		applog.Error("job finished", args...)
		return
	}
	applog.Info("job finished", args...)
}

func (s *JobScheduler) executeStep(step jobs.StepModel) error {
	if step.MaxAttempts <= 0 {
		step.MaxAttempts = 1
	}

	executor := s.executors[job.StepType(step.Type)]

	// Acquire semaphore before marking as running (Fix 2)
	var release func()
	if executor != nil {
		release = s.acquireStepSemaphore(job.StepType(step.Type))
	}

	started := time.Now()
	_, err := s.updateStepExecution(step.ID, string(job.StepStatusRunning), step.Progress, step.Attempts+1, &started, nil, nil)
	if err != nil {
		if release != nil {
			release()
		}
		return err
	}

	applog.Debug("step started",
		"job_id", step.JobID, "step_id", step.ID, "type", step.Type, "attempt", step.Attempts+1)

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
	durationMs := ended.Sub(started).Milliseconds()

	if runErr == nil {
		applog.Debug("step completed",
			"job_id", step.JobID, "step_id", step.ID, "type", step.Type, "duration_ms", durationMs)
		_, err = s.updateStepExecution(step.ID, string(job.StepStatusCompleted), 100, step.Attempts+1, nil, &ended, nil)
		return err
	}

	if errors.Is(runErr, ErrStepSkipped) {
		applog.Debug("step skipped",
			"job_id", step.JobID, "step_id", step.ID, "type", step.Type)
		_, err = s.updateStepExecution(step.ID, string(job.StepStatusSkipped), 100, step.Attempts+1, nil, &ended, nil)
		return err
	}

	// Timeout: don't fail the step. Return it to the queue (without consuming the
	// retry budget) and bump timeout_count, then signal the job to go to the back
	// of the line so the next file gets a turn — like stepping out of the picolé
	// line to think about the flavor. Recurring offenders surface in analytics.
	if errors.Is(runErr, context.DeadlineExceeded) {
		if _, deferErr := s.deferStepForTimeout(step.ID, step.Attempts, runErr.Error()); deferErr != nil {
			return deferErr
		}
		applog.Warn("step timed out, requeued at back of line",
			"job_id", step.JobID, "step_id", step.ID, "type", step.Type, "duration_ms", durationMs)
		return errStepDeferred
	}

	if step.Attempts+1 < step.MaxAttempts {
		runErrMessage := runErr.Error()
		_, updateErr := s.updateStepExecution(step.ID, string(job.StepStatusQueued), step.Progress, step.Attempts+1, nil, nil, &runErrMessage)
		if updateErr != nil {
			return updateErr
		}
		applog.Warn("step failed, retry scheduled",
			"job_id", step.JobID, "step_id", step.ID, "type", step.Type,
			"attempt", step.Attempts+1, "max_attempts", step.MaxAttempts, "error", runErrMessage)
		// Apply retry backoff (Fix 3)
		backoff := time.Duration(config.AppConfig.WorkerRetryBackoffMS) * time.Millisecond * time.Duration(step.Attempts+1)
		if backoff > 0 {
			time.Sleep(backoff)
		}
		return nil
	}

	runErrMessage := runErr.Error()
	_, updateErr := s.updateStepExecution(step.ID, string(job.StepStatusFailed), step.Progress, step.Attempts+1, nil, &ended, &runErrMessage)
	if updateErr != nil {
		return updateErr
	}

	applog.Error("step failed permanently",
		"job_id", step.JobID, "step_id", step.ID, "type", step.Type,
		"attempts", step.Attempts+1, "error", runErrMessage)
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

func (s *JobScheduler) deferStepForTimeout(stepID int, attempts int, lastError string) (bool, error) {
	return s.withTx(func(tx *sql.Tx) (bool, error) {
		updated, err := s.repository.DeferStepForTimeout(tx, stepID, attempts, lastError)
		if err != nil {
			return false, fmt.Errorf("defer step %d for timeout: %w", stepID, err)
		}
		return updated, nil
	})
}

// RecoverInterruptedWork resets jobs/steps stranded in 'running' back to
// 'queued'. Call once on startup before Start() so orphaned work from a previous
// run is reprocessed. Returns the number of jobs and steps reset.
func (s *JobScheduler) RecoverInterruptedWork() (int64, int64, error) {
	var jobsReset, stepsReset int64
	_, err := s.withTx(func(tx *sql.Tx) (bool, error) {
		j, st, recoverErr := s.repository.RecoverInterruptedWork(tx)
		if recoverErr != nil {
			return false, recoverErr
		}
		jobsReset, stepsReset = j, st
		return false, nil
	})
	return jobsReset, stepsReset, err
}

func (s *JobScheduler) requeueJob(jobID int) (bool, error) {
	return s.withTx(func(tx *sql.Tx) (bool, error) {
		updated, err := s.repository.RequeueJob(tx, jobID)
		if err != nil {
			return false, fmt.Errorf("requeue job %d: %w", jobID, err)
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
		if step.Status != string(job.StepStatusQueued) {
			continue
		}

		ended := time.Now()
		if _, updateErr := s.updateStepExecution(step.ID, string(job.StepStatusCanceled), step.Progress, step.Attempts, nil, &ended, nil); updateErr != nil {
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
	if !jobModel.CancelRequested && jobModel.Status != string(job.JobStatusCanceled) {
		return false, nil
	}

	endedAt := time.Now()
	if _, err := s.updateJobExecution(jobID, string(job.JobStatusCanceled), nil, &endedAt, nil, nil); err != nil {
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
		if depStatus != string(job.StepStatusCompleted) && depStatus != string(job.StepStatusSkipped) {
			return false
		}
	}

	return true
}

func stepStatusTerminal(status string) bool {
	switch job.StepStatus(status) {
	case job.StepStatusCompleted, job.StepStatusFailed, job.StepStatusCanceled, job.StepStatusSkipped:
		return true
	default:
		return false
	}
}

func resolveJobStatus(steps []jobs.StepModel) job.JobStatus {
	if len(steps) == 0 {
		return job.JobStatusCompleted
	}

	hasFailed := false
	hasSucceeded := false
	allCanceled := true

	for _, step := range steps {
		switch job.StepStatus(step.Status) {
		case job.StepStatusFailed:
			hasFailed = true
			allCanceled = false
		case job.StepStatusCompleted, job.StepStatusSkipped:
			hasSucceeded = true
			allCanceled = false
		case job.StepStatusCanceled:
			// remains allCanceled = true
		default:
			allCanceled = false
		}
	}

	if allCanceled {
		return job.JobStatusCanceled
	}
	if hasFailed && hasSucceeded {
		return job.JobStatusPartialFail
	}
	if hasFailed {
		return job.JobStatusFailed
	}

	return job.JobStatusCompleted
}

func buildStepSemaphoreMap() map[job.StepType]chan struct{} {
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

	return map[job.StepType]chan struct{}{
		job.StepTypeChecksum:  make(chan struct{}, checksumLimit),
		job.StepTypeMetadata:  make(chan struct{}, metadataLimit),
		job.StepTypeThumbnail: make(chan struct{}, thumbnailLimit),
	}
}

func (s *JobScheduler) acquireStepSemaphore(stepType job.StepType) func() {
	sem, exists := s.stepSem[stepType]
	if !exists || sem == nil {
		return nil
	}

	sem <- struct{}{}
	return func() {
		<-sem
	}
}
