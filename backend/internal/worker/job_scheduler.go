package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/utils"
)

type JobScheduler struct {
	repository            jobs.RepositoryInterface
	executor              StepAtomicExecutor
	workerContext         *WorkerContext
	runInTx               func(fn func(*sql.Tx) error) error
	maxJobsPerTick        int
	stepConcurrencyLimits map[domain.StepType]int
	retryBaseBackoff      time.Duration
	retryMaxBackoff       time.Duration

	retryMutex             sync.Mutex
	retryNotBeforeByStepID map[string]time.Time
	recoveryMutex          sync.Mutex
	hasRecoveredRunning    bool
}

type schedulerStep struct {
	Step            domain.Step
	DependsOnStepID []string
}

func NewJobScheduler(repository jobs.RepositoryInterface, executor StepAtomicExecutor, workerContext *WorkerContext) *JobScheduler {
	if executor == nil {
		executor = NewDefaultStepExecutor()
	}

	maxJobsPerTick := config.AppConfig.WorkerMaxJobsPerTick
	if maxJobsPerTick <= 0 {
		maxJobsPerTick = 50
	}

	scheduler := &JobScheduler{
		repository:             repository,
		executor:               executor,
		workerContext:          workerContext,
		maxJobsPerTick:         maxJobsPerTick,
		stepConcurrencyLimits:  buildStepConcurrencyLimits(),
		retryBaseBackoff:       time.Duration(config.AppConfig.WorkerRetryBaseBackoffMillis) * time.Millisecond,
		retryMaxBackoff:        time.Duration(config.AppConfig.WorkerRetryMaxBackoffMillis) * time.Millisecond,
		retryNotBeforeByStepID: map[string]time.Time{},
	}

	if scheduler.retryBaseBackoff <= 0 {
		scheduler.retryBaseBackoff = 500 * time.Millisecond
	}
	if scheduler.retryMaxBackoff <= 0 {
		scheduler.retryMaxBackoff = 30 * time.Second
	}

	if repository != nil {
		scheduler.runInTx = func(fn func(*sql.Tx) error) error {
			dbContext := repository.GetDbContext()
			if dbContext == nil {
				return fmt.Errorf("jobs db context is nil")
			}
			return dbContext.ExecTx(fn)
		}
	}

	return scheduler
}

func (s *JobScheduler) Start(pollInterval time.Duration) {
	if s == nil || s.repository == nil {
		return
	}
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		_ = s.RunOnce()
	}
}

func (s *JobScheduler) RunOnce() error {
	if s == nil || s.repository == nil || s.executor == nil {
		return nil
	}

	jobModels, err := s.listSchedulableJobs()
	if err != nil {
		return err
	}
	if err := s.recoverInterruptedRunningSteps(jobModels); err != nil {
		return err
	}

	jobModels, err = s.listSchedulableJobs()
	if err != nil {
		return err
	}

	runningByType, err := s.countRunningStepsByType(jobModels)
	if err != nil {
		return err
	}

	for _, jobModel := range jobModels {
		if runErr := s.runSingleJob(jobModel, runningByType); runErr != nil {
			continue
		}
	}

	return nil
}

func (s *JobScheduler) runSingleJob(jobModel jobs.JobModel, runningByType map[domain.StepType]int) error {
	steps, err := s.repository.GetStepsByJobID(jobModel.ID)
	if err != nil {
		return err
	}

	parsedSteps, parseErr := parseSchedulerSteps(steps)
	if parseErr != nil {
		return parseErr
	}

	jobDomain, jobErr := toDomainJob(jobModel)
	if jobErr != nil {
		return jobErr
	}

	if jobModel.CancelRequested {
		if s.workerContext != nil {
			s.workerContext.CancelJobExecution(jobDomain.ID)
		}
		return s.applyCancellation(jobDomain, parsedSteps)
	}

	nextStep := selectNextEligibleStep(parsedSteps)
	if nextStep == nil {
		return s.reconcileJobState(jobModel.ID, parsedSteps, "")
	}

	if s.isStepRetryBlocked(nextStep.Step.ID) {
		return nil
	}
	if s.isStepTypeAtCapacity(nextStep.Step.Type, runningByType) {
		return nil
	}

	if s.workerContext != nil {
		s.workerContext.EnsureJobExecutionContext(jobDomain.ID)
	}

	if transitionErr := s.transitionStepToRunning(jobDomain, nextStep.Step); transitionErr != nil {
		return transitionErr
	}

	nextStep.Step.Scope = jobDomain.Scope
	runningByType[nextStep.Step.Type]++

	execErr := s.executor.ExecuteStep(nextStep.Step, s.workerContext)
	if runningByType[nextStep.Step.Type] > 0 {
		runningByType[nextStep.Step.Type]--
	}

	if finalizeErr := s.finalizeStep(nextStep.Step, execErr); finalizeErr != nil {
		return finalizeErr
	}

	stepsAfterExecution, stepsErr := s.repository.GetStepsByJobID(jobModel.ID)
	if stepsErr != nil {
		return stepsErr
	}

	parsedAfterExecution, parseAfterErr := parseSchedulerSteps(stepsAfterExecution)
	if parseAfterErr != nil {
		return parseAfterErr
	}

	lastError := ""
	if execErr != nil && !isStepSkipped(execErr) {
		lastError = execErr.Error()
	}

	return s.reconcileJobState(jobModel.ID, parsedAfterExecution, lastError)
}

func (s *JobScheduler) applyCancellation(job domain.Job, steps []schedulerStep) error {
	if s.runInTx == nil {
		return fmt.Errorf("scheduler transaction is not configured")
	}

	endedAt := time.Now().UTC()
	cancelMessage := "job cancellation requested"

	hasRunningStep := false
	for _, step := range steps {
		if step.Step.Status == domain.StepStatusRunning {
			hasRunningStep = true
			break
		}
	}

	err := s.runInTx(func(tx *sql.Tx) error {
		for _, step := range steps {
			if step.Step.Status != domain.StepStatusQueued {
				continue
			}

			updatedExecution, err := s.repository.UpdateStepExecution(tx, step.Step.ID, step.Step.Attempts, cancelMessage, 100, nil, &endedAt)
			if err != nil {
				return err
			}
			if !updatedExecution {
				return fmt.Errorf("step %s execution metadata was not updated for cancellation", step.Step.ID)
			}

			updatedStatus, err := s.repository.UpdateStepStatus(
				tx,
				step.Step.ID,
				string(domain.StepStatusQueued),
				string(domain.StepStatusCanceled),
				nil,
				&endedAt,
				cancelMessage,
			)
			if err != nil {
				return err
			}
			if !updatedStatus {
				return fmt.Errorf("step %s status was not updated to canceled", step.Step.ID)
			}
		}

		if hasRunningStep {
			return nil
		}

		updated, err := s.repository.UpdateJobStatus(tx, job.ID, string(domain.JobStatusRunning), string(domain.JobStatusCanceled), nil, &endedAt, cancelMessage)
		if err != nil {
			return err
		}
		if updated {
			return nil
		}

		updated, err = s.repository.UpdateJobStatus(tx, job.ID, string(domain.JobStatusQueued), string(domain.JobStatusCanceled), nil, &endedAt, cancelMessage)
		if err != nil {
			return err
		}
		if !updated {
			return fmt.Errorf("job %s was not updated to canceled", job.ID)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if !hasRunningStep && s.workerContext != nil {
		s.workerContext.ReleaseJobExecution(job.ID)
	}

	return nil
}

func (s *JobScheduler) listSchedulableJobs() ([]jobs.JobModel, error) {
	queuedJobs, err := s.repository.ListJobs(jobs.JobFilter{
		Status: utils.Optional[string]{HasValue: true, Value: string(domain.JobStatusQueued)},
	}, 1, s.maxJobsPerTick)
	if err != nil {
		return nil, err
	}

	runningJobs, err := s.repository.ListJobs(jobs.JobFilter{
		Status: utils.Optional[string]{HasValue: true, Value: string(domain.JobStatusRunning)},
	}, 1, s.maxJobsPerTick)
	if err != nil {
		return nil, err
	}

	jobsToProcess := append([]jobs.JobModel{}, queuedJobs.Items...)
	jobsToProcess = append(jobsToProcess, runningJobs.Items...)

	sort.SliceStable(jobsToProcess, func(i, j int) bool {
		if jobsToProcess[i].Priority == jobsToProcess[j].Priority {
			return jobsToProcess[i].CreatedAt.Before(jobsToProcess[j].CreatedAt)
		}
		return jobsToProcess[i].Priority > jobsToProcess[j].Priority
	})

	return jobsToProcess, nil
}

func (s *JobScheduler) countRunningStepsByType(jobModels []jobs.JobModel) (map[domain.StepType]int, error) {
	counts := map[domain.StepType]int{}
	for _, jobModel := range jobModels {
		steps, err := s.repository.GetStepsByJobID(jobModel.ID)
		if err != nil {
			return nil, err
		}

		parsedSteps, err := parseSchedulerSteps(steps)
		if err != nil {
			return nil, err
		}

		for _, step := range parsedSteps {
			if step.Step.Status == domain.StepStatusRunning {
				counts[step.Step.Type]++
			}
		}
	}
	return counts, nil
}

func (s *JobScheduler) transitionStepToRunning(job domain.Job, step domain.Step) error {
	if s.runInTx == nil {
		return fmt.Errorf("scheduler transaction is not configured")
	}

	now := time.Now().UTC()

	return s.runInTx(func(tx *sql.Tx) error {
		if job.Status == domain.JobStatusQueued {
			updated, err := s.repository.UpdateJobStatus(
				tx,
				job.ID,
				string(domain.JobStatusQueued),
				string(domain.JobStatusRunning),
				&now,
				nil,
				"",
			)
			if err != nil {
				return err
			}
			if !updated {
				return fmt.Errorf("job %s was not updated to running", job.ID)
			}
		}

		updated, err := s.repository.UpdateStepStatus(
			tx,
			step.ID,
			string(domain.StepStatusQueued),
			string(domain.StepStatusRunning),
			&now,
			nil,
			"",
		)
		if err != nil {
			return err
		}
		if !updated {
			return fmt.Errorf("step %s was not updated to running", step.ID)
		}

		return nil
	})
}

func (s *JobScheduler) finalizeStep(step domain.Step, execErr error) error {
	if s.runInTx == nil {
		return fmt.Errorf("scheduler transaction is not configured")
	}

	attempts := step.Attempts + 1
	maxAttempts := s.resolveMaxAttempts(step)
	progress := 100
	lastError := ""
	toStatus := domain.StepStatusCompleted
	endedAt := time.Now().UTC()
	endedAtValue := &endedAt

	if isStepSkipped(execErr) {
		toStatus = domain.StepStatusSkipped
	} else if isStepCanceled(execErr) {
		toStatus = domain.StepStatusCanceled
		lastError = execErr.Error()
	} else if execErr != nil {
		progress = 0
		lastError = execErr.Error()

		if attempts < maxAttempts && isTransientExecutionError(execErr) {
			toStatus = domain.StepStatusQueued
			endedAtValue = nil
			s.setStepRetryNotBefore(step.ID, s.nextRetryAt(attempts))
		} else {
			toStatus = domain.StepStatusFailed
			s.clearStepRetryNotBefore(step.ID)
		}
	} else {
		s.clearStepRetryNotBefore(step.ID)
	}

	if toStatus == domain.StepStatusCanceled || toStatus == domain.StepStatusSkipped || toStatus == domain.StepStatusCompleted {
		s.clearStepRetryNotBefore(step.ID)
	}

	return s.runInTx(func(tx *sql.Tx) error {
		updatedExecution, err := s.repository.UpdateStepExecution(tx, step.ID, attempts, lastError, progress, nil, endedAtValue)
		if err != nil {
			return err
		}
		if !updatedExecution {
			return fmt.Errorf("step %s execution metadata was not updated", step.ID)
		}

		updatedStatus, err := s.repository.UpdateStepStatus(
			tx,
			step.ID,
			string(domain.StepStatusRunning),
			string(toStatus),
			nil,
			endedAtValue,
			lastError,
		)
		if err != nil {
			return err
		}
		if !updatedStatus {
			if toStatus == domain.StepStatusQueued || toStatus == domain.StepStatusCanceled {
				return nil
			}
			return fmt.Errorf("step %s status was not updated to %s", step.ID, toStatus)
		}

		return nil
	})
}

func (s *JobScheduler) reconcileJobState(jobID string, steps []schedulerStep, lastError string) error {
	if s.runInTx == nil {
		return fmt.Errorf("scheduler transaction is not configured")
	}

	jobStatus := calculateJobStatusFromSteps(steps)
	if jobStatus == domain.JobStatusRunning || jobStatus == domain.JobStatusQueued {
		return nil
	}

	endedAt := time.Now().UTC()

	err := s.runInTx(func(tx *sql.Tx) error {
		updated, err := s.repository.UpdateJobStatus(tx, jobID, string(domain.JobStatusRunning), string(jobStatus), nil, &endedAt, lastError)
		if err != nil {
			return err
		}
		if updated {
			return nil
		}

		updated, err = s.repository.UpdateJobStatus(tx, jobID, string(domain.JobStatusQueued), string(jobStatus), nil, &endedAt, lastError)
		if err != nil {
			return err
		}
		if !updated {
			return fmt.Errorf("job %s status was not updated to %s", jobID, jobStatus)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.workerContext != nil {
		s.workerContext.ReleaseJobExecution(jobID)
	}

	return nil
}

func (s *JobScheduler) resolveMaxAttempts(step domain.Step) int {
	if step.MaxAttempts > 0 {
		return step.MaxAttempts
	}
	if config.AppConfig.WorkerRetryDefaultMaxAttempts > 0 {
		return config.AppConfig.WorkerRetryDefaultMaxAttempts
	}
	return 3
}

func (s *JobScheduler) isStepTypeAtCapacity(stepType domain.StepType, runningByType map[domain.StepType]int) bool {
	limit := s.stepConcurrencyLimits[stepType]
	if limit <= 0 {
		limit = config.AppConfig.WorkerStepConcurrencyDefault
	}
	if limit <= 0 {
		limit = 1
	}
	return runningByType[stepType] >= limit
}

func (s *JobScheduler) isStepRetryBlocked(stepID string) bool {
	s.retryMutex.Lock()
	defer s.retryMutex.Unlock()

	notBefore, exists := s.retryNotBeforeByStepID[stepID]
	if !exists {
		return false
	}

	if time.Now().UTC().After(notBefore) || time.Now().UTC().Equal(notBefore) {
		delete(s.retryNotBeforeByStepID, stepID)
		return false
	}

	return true
}

func (s *JobScheduler) setStepRetryNotBefore(stepID string, retryAt time.Time) {
	s.retryMutex.Lock()
	defer s.retryMutex.Unlock()
	s.retryNotBeforeByStepID[stepID] = retryAt
}

func (s *JobScheduler) clearStepRetryNotBefore(stepID string) {
	s.retryMutex.Lock()
	defer s.retryMutex.Unlock()
	delete(s.retryNotBeforeByStepID, stepID)
}

func (s *JobScheduler) nextRetryAt(attempts int) time.Time {
	if attempts < 1 {
		attempts = 1
	}

	delay := s.retryBaseBackoff
	for i := 1; i < attempts; i++ {
		delay *= 2
		if delay >= s.retryMaxBackoff {
			delay = s.retryMaxBackoff
			break
		}
	}

	if delay > s.retryMaxBackoff {
		delay = s.retryMaxBackoff
	}

	return time.Now().UTC().Add(delay)
}

func isTransientExecutionError(err error) bool {
	if err == nil {
		return false
	}
	if isTransientStepError(err) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, syscall.EBUSY) || errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}

	temporaryErr, ok := err.(interface{ Temporary() bool })
	if ok && temporaryErr.Temporary() {
		return true
	}

	normalizedErr := strings.ToLower(err.Error())
	transientHints := []string{
		"timeout",
		"temporary",
		"temporarily",
		"try again",
		"resource busy",
		"deadlock",
		"connection reset",
		"broken pipe",
	}
	for _, hint := range transientHints {
		if strings.Contains(normalizedErr, hint) {
			return true
		}
	}

	return false
}

func buildStepConcurrencyLimits() map[domain.StepType]int {
	defaultLimit := config.AppConfig.WorkerStepConcurrencyDefault
	if defaultLimit <= 0 {
		defaultLimit = 1
	}

	limit := func(value int) int {
		if value <= 0 {
			return defaultLimit
		}
		return value
	}

	return map[domain.StepType]int{
		domain.StepTypeScanFilesystem: limit(config.AppConfig.WorkerStepConcurrencyScanFilesystem),
		domain.StepTypeDiffAgainstDB:  limit(config.AppConfig.WorkerStepConcurrencyDiffAgainstDB),
		domain.StepTypeMetadata:       limit(config.AppConfig.WorkerStepConcurrencyMetadata),
		domain.StepTypeChecksum:       limit(config.AppConfig.WorkerStepConcurrencyChecksum),
		domain.StepTypePersist:        limit(config.AppConfig.WorkerStepConcurrencyPersist),
		domain.StepTypeThumbnail:      limit(config.AppConfig.WorkerStepConcurrencyThumbnail),
		domain.StepTypePlaylistIndex:  limit(config.AppConfig.WorkerStepConcurrencyPlaylistIndex),
		domain.StepTypeMarkDeleted:    limit(config.AppConfig.WorkerStepConcurrencyMarkDeleted),
	}
}

func parseSchedulerSteps(stepModels []jobs.StepModel) ([]schedulerStep, error) {
	steps := make([]schedulerStep, 0, len(stepModels))

	for _, stepModel := range stepModels {
		dependsOn := []string{}
		if stepModel.DependsOnJSON != "" {
			if err := json.Unmarshal([]byte(stepModel.DependsOnJSON), &dependsOn); err != nil {
				return nil, fmt.Errorf("parse step dependencies: %w", err)
			}
		}

		status := domain.StepStatus(stepModel.Status)
		stepType := domain.StepType(stepModel.Type)

		steps = append(steps, schedulerStep{
			Step: domain.Step{
				ID:          stepModel.ID,
				JobID:       stepModel.JobID,
				Type:        stepType,
				Status:      status,
				Attempts:    stepModel.Attempts,
				MaxAttempts: stepModel.MaxAttempts,
				Error:       stepModel.LastError,
				CreatedAt:   stepModel.CreatedAt,
				StartedAt:   parseNullTime(stepModel.StartedAt),
				FinishedAt:  parseNullTime(stepModel.EndedAt),
			},
			DependsOnStepID: dependsOn,
		})
	}

	return steps, nil
}

func selectNextEligibleStep(steps []schedulerStep) *schedulerStep {
	for _, step := range steps {
		if step.Step.Status == domain.StepStatusRunning {
			return nil
		}
	}

	statusByStepID := map[string]domain.StepStatus{}
	for _, step := range steps {
		statusByStepID[step.Step.ID] = step.Step.Status
	}

	for i := range steps {
		if steps[i].Step.Status != domain.StepStatusQueued {
			continue
		}

		if hasPendingDependencies(statusByStepID, steps[i].DependsOnStepID) {
			continue
		}

		return &steps[i]
	}

	return nil
}

func hasPendingDependencies(statusByStepID map[string]domain.StepStatus, dependencies []string) bool {
	for _, dependencyID := range dependencies {
		dependencyStatus, exists := statusByStepID[dependencyID]
		if !exists {
			return true
		}
		if dependencyStatus != domain.StepStatusCompleted && dependencyStatus != domain.StepStatusSkipped {
			return true
		}
	}
	return false
}

func calculateJobStatusFromSteps(steps []schedulerStep) domain.JobStatus {
	if len(steps) == 0 {
		return domain.JobStatusCompleted
	}

	hasQueued := false
	hasRunning := false
	hasFailed := false
	hasCanceled := false
	hasNonFailedTerminal := false

	for _, step := range steps {
		switch step.Step.Status {
		case domain.StepStatusFailed:
			hasFailed = true
		case domain.StepStatusQueued:
			hasQueued = true
		case domain.StepStatusRunning:
			hasRunning = true
		case domain.StepStatusCanceled:
			hasCanceled = true
			hasNonFailedTerminal = true
		case domain.StepStatusCompleted, domain.StepStatusSkipped:
			hasNonFailedTerminal = true
		}
	}

	if hasRunning {
		return domain.JobStatusRunning
	}
	if hasFailed {
		if hasNonFailedTerminal {
			return domain.JobStatusPartialFail
		}
		return domain.JobStatusFailed
	}
	if hasQueued {
		return domain.JobStatusQueued
	}
	if hasCanceled {
		return domain.JobStatusCanceled
	}

	return domain.JobStatusCompleted
}

func (s *JobScheduler) recoverInterruptedRunningSteps(jobModels []jobs.JobModel) error {
	s.recoveryMutex.Lock()
	if s.hasRecoveredRunning {
		s.recoveryMutex.Unlock()
		return nil
	}
	s.hasRecoveredRunning = true
	s.recoveryMutex.Unlock()

	if s.runInTx == nil {
		return fmt.Errorf("scheduler transaction is not configured")
	}

	const recoveryMessage = "step returned to queued after scheduler restart"

	for _, jobModel := range jobModels {
		if domain.JobStatus(jobModel.Status) != domain.JobStatusRunning {
			continue
		}

		steps, err := s.repository.GetStepsByJobID(jobModel.ID)
		if err != nil {
			return err
		}
		parsedSteps, err := parseSchedulerSteps(steps)
		if err != nil {
			return err
		}

		hasRunningStep := false
		for _, step := range parsedSteps {
			if step.Step.Status == domain.StepStatusRunning {
				hasRunningStep = true
				break
			}
		}
		if !hasRunningStep {
			continue
		}

		if err := s.runInTx(func(tx *sql.Tx) error {
			for _, step := range parsedSteps {
				if step.Step.Status != domain.StepStatusRunning {
					continue
				}

				updatedExecution, execErr := s.repository.UpdateStepExecution(
					tx,
					step.Step.ID,
					step.Step.Attempts,
					recoveryMessage,
					0,
					nil,
					nil,
				)
				if execErr != nil {
					return execErr
				}
				if !updatedExecution {
					return fmt.Errorf("step %s execution metadata was not updated during recovery", step.Step.ID)
				}

				updatedStatus, statusErr := s.repository.UpdateStepStatus(
					tx,
					step.Step.ID,
					string(domain.StepStatusRunning),
					string(domain.StepStatusQueued),
					nil,
					nil,
					recoveryMessage,
				)
				if statusErr != nil {
					return statusErr
				}
				if !updatedStatus {
					return fmt.Errorf("step %s was not requeued during recovery", step.Step.ID)
				}
			}

			updatedJob, updateJobErr := s.repository.UpdateJobStatus(
				tx,
				jobModel.ID,
				string(domain.JobStatusRunning),
				string(domain.JobStatusQueued),
				nil,
				nil,
				recoveryMessage,
			)
			if updateJobErr != nil {
				return updateJobErr
			}
			if !updatedJob {
				return fmt.Errorf("job %s was not requeued during recovery", jobModel.ID)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func toDomainJob(job jobs.JobModel) (domain.Job, error) {
	scope := domain.ScopePayload{}
	if job.ScopeJSON != "" {
		if err := json.Unmarshal([]byte(job.ScopeJSON), &scope); err != nil {
			return domain.Job{}, fmt.Errorf("parse job scope: %w", err)
		}
	}

	return domain.Job{
		ID:         job.ID,
		Type:       domain.JobType(job.Type),
		Status:     domain.JobStatus(job.Status),
		Priority:   domain.JobPriority(job.Priority),
		Scope:      scope,
		Error:      job.LastError,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.CreatedAt,
		StartedAt:  parseNullTime(job.StartedAt),
		FinishedAt: parseNullTime(job.EndedAt),
	}, nil
}

func parseNullTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}
