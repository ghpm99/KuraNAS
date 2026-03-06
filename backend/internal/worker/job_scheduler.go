package worker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/utils"
)

type JobScheduler struct {
	repository     jobs.RepositoryInterface
	executor       StepAtomicExecutor
	workerContext  *WorkerContext
	runInTx        func(fn func(*sql.Tx) error) error
	maxJobsPerTick int
}

type schedulerStep struct {
	Step            domain.Step
	DependsOnStepID []string
}

func NewJobScheduler(repository jobs.RepositoryInterface, executor StepAtomicExecutor, workerContext *WorkerContext) *JobScheduler {
	if executor == nil {
		executor = NewDefaultStepExecutor()
	}

	scheduler := &JobScheduler{
		repository:     repository,
		executor:       executor,
		workerContext:  workerContext,
		maxJobsPerTick: 50,
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

	for _, jobModel := range jobModels {
		if runErr := s.runSingleJob(jobModel); runErr != nil {
			continue
		}
	}

	return nil
}

func (s *JobScheduler) runSingleJob(jobModel jobs.JobModel) error {
	steps, err := s.repository.GetStepsByJobID(jobModel.ID)
	if err != nil {
		return err
	}

	parsedSteps, parseErr := parseSchedulerSteps(steps)
	if parseErr != nil {
		return parseErr
	}

	nextStep := selectNextEligibleStep(parsedSteps)
	if nextStep == nil {
		return s.reconcileJobState(jobModel.ID, parsedSteps, "")
	}

	jobDomain, jobErr := toDomainJob(jobModel)
	if jobErr != nil {
		return jobErr
	}

	if transitionErr := s.transitionStepToRunning(jobDomain, nextStep.Step); transitionErr != nil {
		return transitionErr
	}

	nextStep.Step.Scope = jobDomain.Scope

	execErr := s.executor.ExecuteStep(nextStep.Step, s.workerContext)
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

	endedAt := time.Now().UTC()
	attempts := step.Attempts + 1
	progress := 100
	lastError := ""
	toStatus := string(domain.StepStatusCompleted)
	if execErr != nil && !isStepSkipped(execErr) {
		progress = 0
		lastError = execErr.Error()
		toStatus = string(domain.StepStatusFailed)
	} else if isStepSkipped(execErr) {
		toStatus = string(domain.StepStatusSkipped)
	}

	return s.runInTx(func(tx *sql.Tx) error {
		updatedExecution, err := s.repository.UpdateStepExecution(tx, step.ID, attempts, lastError, progress, nil, &endedAt)
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
			toStatus,
			nil,
			&endedAt,
			lastError,
		)
		if err != nil {
			return err
		}
		if !updatedStatus {
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

	return s.runInTx(func(tx *sql.Tx) error {
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

	allDone := true
	for _, step := range steps {
		switch step.Step.Status {
		case domain.StepStatusFailed, domain.StepStatusCanceled:
			return domain.JobStatusFailed
		case domain.StepStatusQueued:
			allDone = false
		case domain.StepStatusRunning:
			return domain.JobStatusRunning
		case domain.StepStatusCompleted, domain.StepStatusSkipped:
			continue
		default:
			allDone = false
		}
	}

	if allDone {
		return domain.JobStatusCompleted
	}

	return domain.JobStatusQueued
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
