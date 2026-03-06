package jobs

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/utils"
)

var (
	ErrJobNotFound         = errors.New("job not found")
	ErrInvalidPage         = errors.New("invalid page")
	ErrInvalidPageSize     = errors.New("invalid page size")
	ErrJobCancelNotAllowed = errors.New("job cancel not allowed")
)

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func (s *Service) GetJobByID(id string) (JobSummaryDto, error) {
	job, err := s.Repository.GetJobByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JobSummaryDto{}, ErrJobNotFound
		}
		return JobSummaryDto{}, fmt.Errorf("GetJobByID: %w", err)
	}

	steps, err := s.Repository.GetStepsByJobID(id)
	if err != nil {
		return JobSummaryDto{}, fmt.Errorf("GetJobByID steps: %w", err)
	}

	return toJobSummaryDto(job, steps), nil
}

func (s *Service) ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobSummaryDto], error) {
	if page < 1 {
		return utils.PaginationResponse[JobSummaryDto]{}, ErrInvalidPage
	}
	if pageSize < 1 {
		return utils.PaginationResponse[JobSummaryDto]{}, ErrInvalidPageSize
	}

	jobsPage, err := s.Repository.ListJobs(filter, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[JobSummaryDto]{}, fmt.Errorf("ListJobs: %w", err)
	}

	items := make([]JobSummaryDto, 0, len(jobsPage.Items))
	for _, job := range jobsPage.Items {
		steps, stepsErr := s.Repository.GetStepsByJobID(job.ID)
		if stepsErr != nil {
			return utils.PaginationResponse[JobSummaryDto]{}, fmt.Errorf("ListJobs steps for %s: %w", job.ID, stepsErr)
		}

		items = append(items, toJobSummaryDto(job, steps))
	}

	return utils.PaginationResponse[JobSummaryDto]{
		Items:      items,
		Pagination: jobsPage.Pagination,
	}, nil
}

func (s *Service) GetStepsByJobID(jobID string) ([]StepDto, error) {
	_, err := s.Repository.GetJobByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("GetStepsByJobID job: %w", err)
	}

	steps, err := s.Repository.GetStepsByJobID(jobID)
	if err != nil {
		return nil, fmt.Errorf("GetStepsByJobID: %w", err)
	}

	result := make([]StepDto, 0, len(steps))
	for _, step := range steps {
		result = append(result, toStepDto(step))
	}

	return result, nil
}

func (s *Service) CancelJob(id string) (JobSummaryDto, error) {
	job, err := s.Repository.GetJobByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JobSummaryDto{}, ErrJobNotFound
		}
		return JobSummaryDto{}, fmt.Errorf("CancelJob get job: %w", err)
	}

	jobType := domain.JobType(job.Type)
	if jobType != domain.JobTypeStartupScan && jobType != domain.JobTypeReindexFolder {
		return JobSummaryDto{}, ErrJobCancelNotAllowed
	}

	jobStatus := domain.JobStatus(job.Status)
	if jobStatus != domain.JobStatusQueued && jobStatus != domain.JobStatusRunning {
		return JobSummaryDto{}, ErrJobCancelNotAllowed
	}

	if !job.CancelRequested {
		dbContext := s.Repository.GetDbContext()
		if dbContext != nil {
			txErr := dbContext.ExecTx(func(tx *sql.Tx) error {
				updated, cancelErr := s.Repository.RequestJobCancelCascade(tx, id)
				if cancelErr != nil {
					return cancelErr
				}
				if !updated {
					return fmt.Errorf("CancelJob request was not persisted")
				}
				return nil
			})
			if txErr != nil {
				return JobSummaryDto{}, fmt.Errorf("CancelJob request: %w", txErr)
			}
		} else {
			updated, cancelErr := s.Repository.RequestJobCancelCascade(nil, id)
			if cancelErr != nil {
				return JobSummaryDto{}, fmt.Errorf("CancelJob request: %w", cancelErr)
			}
			if !updated {
				return JobSummaryDto{}, fmt.Errorf("CancelJob request was not persisted")
			}
		}
	}

	return s.GetJobByID(id)
}

func toJobSummaryDto(job JobModel, steps []StepModel) JobSummaryDto {
	var parentJobID *string
	if job.ParentJobID.Valid {
		parentJobID = &job.ParentJobID.String
	}

	return JobSummaryDto{
		ID:              job.ID,
		Type:            job.Type,
		Priority:        job.Priority,
		ParentJobID:     parentJobID,
		ScopeJSON:       job.ScopeJSON,
		Status:          job.Status,
		CreatedAt:       job.CreatedAt,
		StartedAt:       parseNullTime(job.StartedAt),
		EndedAt:         parseNullTime(job.EndedAt),
		CancelRequested: job.CancelRequested,
		LastError:       job.LastError,
		Progress:        buildProgressSummary(steps),
	}
}

func toStepDto(step StepModel) StepDto {
	return StepDto{
		ID:            step.ID,
		JobID:         step.JobID,
		Type:          step.Type,
		Status:        step.Status,
		DependsOnJSON: step.DependsOnJSON,
		Attempts:      step.Attempts,
		MaxAttempts:   step.MaxAttempts,
		LastError:     step.LastError,
		Progress:      clampProgress(step.Progress),
		PayloadJSON:   step.PayloadJSON,
		CreatedAt:     step.CreatedAt,
		StartedAt:     parseNullTime(step.StartedAt),
		EndedAt:       parseNullTime(step.EndedAt),
	}
}

func buildProgressSummary(steps []StepModel) JobProgressSummaryDto {
	summary := JobProgressSummaryDto{TotalSteps: len(steps)}
	if len(steps) == 0 {
		return summary
	}

	total := 0
	for _, step := range steps {
		status := domain.StepStatus(step.Status)
		switch status {
		case domain.StepStatusQueued:
			summary.QueuedSteps++
		case domain.StepStatusRunning:
			summary.RunningSteps++
		case domain.StepStatusCompleted:
			summary.CompletedSteps++
		case domain.StepStatusFailed:
			summary.FailedSteps++
		case domain.StepStatusCanceled:
			summary.CanceledSteps++
		case domain.StepStatusSkipped:
			summary.SkippedSteps++
		}

		total += getStepProgressForAggregation(step)
	}

	summary.Progress = int(math.Round(float64(total) / float64(len(steps))))
	return summary
}

func getStepProgressForAggregation(step StepModel) int {
	status := domain.StepStatus(step.Status)
	if status == domain.StepStatusCompleted || status == domain.StepStatusSkipped || status == domain.StepStatusFailed || status == domain.StepStatusCanceled {
		return 100
	}

	return clampProgress(step.Progress)
}

func clampProgress(progress int) int {
	if progress < 0 {
		return 0
	}
	if progress > 100 {
		return 100
	}

	return progress
}

func parseNullTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	t := value.Time
	return &t
}
