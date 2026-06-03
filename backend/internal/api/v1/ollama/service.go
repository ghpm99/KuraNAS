package ollama

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/database"
)

// Job and step type identifiers. Declared as literals (matching the worker's
// StepType/JobType values) to avoid importing the worker package and creating
// an import cycle.
const (
	pullJobType  = "ollama_pull"
	pullStepType = "ollama_model_pull"
)

type Service struct {
	client   *Client
	baseURL  func() string
	jobsRepo jobs.RepositoryInterface
}

func NewService(baseURL func() string, jobsRepo jobs.RepositoryInterface) *Service {
	return &Service{
		client:   NewClient(baseURL),
		baseURL:  baseURL,
		jobsRepo: jobsRepo,
	}
}

// GetStatus probes the daemon for reachability, version and installed models.
// It never returns an error: an unreachable daemon yields Reachable=false.
func (s *Service) GetStatus(ctx context.Context) StatusDto {
	status := StatusDto{BaseURL: s.baseURL(), Models: []ModelDto{}}

	version, err := s.client.Version(ctx)
	if err != nil {
		return status
	}
	status.Reachable = true
	status.Version = version

	if models, listErr := s.client.ListModels(ctx); listErr == nil {
		status.Models = models
	}
	return status
}

func (s *Service) ListModels(ctx context.Context) ([]ModelDto, error) {
	return s.client.ListModels(ctx)
}

func (s *Service) DeleteModel(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidModelName
	}
	return s.client.DeleteModel(ctx, name)
}

// PullModel enqueues a background job that downloads a model. The actual
// streaming download is performed by the worker, which reports progress on the
// job's step. Returns the job id so the caller can track it.
func (s *Service) PullModel(name string) (int, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, ErrInvalidModelName
	}
	if s.jobsRepo == nil {
		return 0, ErrJobsUnavailable
	}

	stepPayload, err := json.Marshal(PullStepPayload{Model: name, BaseURL: s.baseURL()})
	if err != nil {
		return 0, err
	}
	scope, err := json.Marshal(map[string]any{"model": name})
	if err != nil {
		return 0, err
	}

	var jobID int
	err = database.ExecOptionalTx(s.jobsRepo.GetDbContext(), func(tx *sql.Tx) error {
		job, createErr := s.jobsRepo.CreateJob(tx, jobs.JobModel{
			Type:            pullJobType,
			Priority:        "normal",
			Scope:           scope,
			Status:          "queued",
			CancelRequested: false,
		})
		if createErr != nil {
			return createErr
		}

		_, stepErr := s.jobsRepo.CreateStep(tx, jobs.StepModel{
			JobID:       job.ID,
			Type:        pullStepType,
			Status:      "queued",
			DependsOn:   []byte("[]"),
			Attempts:    0,
			MaxAttempts: 1,
			Progress:    0,
			Payload:     stepPayload,
		})
		if stepErr != nil {
			return stepErr
		}

		jobID = job.ID
		return nil
	})
	if err != nil {
		return 0, err
	}

	return jobID, nil
}
