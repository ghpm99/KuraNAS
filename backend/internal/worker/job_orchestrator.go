package worker

import (
	"database/sql"
	"encoding/json"
	"fmt"

	jobs "nas-go/api/internal/api/v1/jobs"
)

type PlannedStep struct {
	Key         string
	Type        StepType
	DependsOn   []string
	Payload     []byte
	MaxAttempts int
}

type PlannedJob struct {
	Type     JobType
	Priority JobPriority
	Scope    JobScope
	Steps    []PlannedStep
}

type JobOrchestrator struct {
	repository jobs.RepositoryInterface
	scheduler  *JobScheduler
}

func NewJobOrchestrator(repository jobs.RepositoryInterface, scheduler *JobScheduler) *JobOrchestrator {
	return &JobOrchestrator{repository: repository, scheduler: scheduler}
}

func (o *JobOrchestrator) CreateJob(plan PlannedJob) (int, error) {
	if err := plan.Validate(); err != nil {
		return 0, err
	}

	if o == nil || o.repository == nil {
		return 0, fmt.Errorf("job orchestrator repository is required")
	}

	var createdJob jobs.JobModel

	err := o.withTx(func(tx *sql.Tx) error {
		scopeJSON, scopeErr := json.Marshal(plan.Scope)
		if scopeErr != nil {
			return fmt.Errorf("marshal scope: %w", scopeErr)
		}

		created, createErr := o.repository.CreateJob(tx, jobs.JobModel{
			Type:            string(plan.Type),
			Priority:        string(plan.Priority),
			Scope:           scopeJSON,
			Status:          string(JobStatusQueued),
			CancelRequested: false,
			LastError:       "",
		})
		if createErr != nil {
			return fmt.Errorf("create job: %w", createErr)
		}
		createdJob = created

		stepIDByKey := map[string]int{}
		for _, step := range plan.Steps {
			dependsOn := make([]int, 0, len(step.DependsOn))
			for _, dependencyKey := range step.DependsOn {
				dependencyID, exists := stepIDByKey[dependencyKey]
				if !exists {
					return fmt.Errorf("step %q depends on unknown step %q", step.Key, dependencyKey)
				}
				dependsOn = append(dependsOn, dependencyID)
			}

			dependsOnJSON, marshalErr := json.Marshal(dependsOn)
			if marshalErr != nil {
				return fmt.Errorf("marshal step dependencies: %w", marshalErr)
			}

			maxAttempts := step.MaxAttempts
			if maxAttempts <= 0 {
				maxAttempts = 1
			}

			createdStep, createStepErr := o.repository.CreateStep(tx, jobs.StepModel{
				JobID:       createdJob.ID,
				Type:        string(step.Type),
				Status:      string(StepStatusQueued),
				DependsOn:   dependsOnJSON,
				Attempts:    0,
				MaxAttempts: maxAttempts,
				LastError:   "",
				Progress:    0,
				Payload:     step.Payload,
			})
			if createStepErr != nil {
				return fmt.Errorf("create step %q: %w", step.Key, createStepErr)
			}

			stepIDByKey[step.Key] = createdStep.ID
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	if o.scheduler != nil {
		o.scheduler.Enqueue(createdJob.ID)
	}

	return createdJob.ID, nil
}

func (o *JobOrchestrator) withTx(fn func(*sql.Tx) error) error {
	dbContext := o.repository.GetDbContext()
	if dbContext == nil {
		return fn(nil)
	}
	return dbContext.ExecTx(fn)
}

func (p PlannedJob) Validate() error {
	if !p.Type.IsValid() {
		return fmt.Errorf("invalid job type: %q", p.Type)
	}
	if !p.Priority.IsValid() {
		return fmt.Errorf("invalid job priority: %q", p.Priority)
	}
	if len(p.Steps) == 0 {
		return fmt.Errorf("planned job requires at least one step")
	}

	seenStepKeys := map[string]struct{}{}
	for _, step := range p.Steps {
		if step.Key == "" {
			return fmt.Errorf("planned step key is required")
		}
		if _, exists := seenStepKeys[step.Key]; exists {
			return fmt.Errorf("duplicate planned step key: %q", step.Key)
		}
		if !step.Type.IsValid() {
			return fmt.Errorf("invalid step type for key %q: %q", step.Key, step.Type)
		}

		seenStepKeys[step.Key] = struct{}{}
	}

	return nil
}
