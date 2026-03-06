package jobs

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"nas-go/api/pkg/utils"
)

var (
	ErrInvalidJobID = errors.New("invalid job id")
	ErrJobNotFound  = errors.New("job not found")
)

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func (s *Service) GetJobByID(id int) (JobDto, error) {
	if id <= 0 {
		return JobDto{}, ErrInvalidJobID
	}

	jobModel, err := s.Repository.GetJobByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JobDto{}, ErrJobNotFound
		}
		return JobDto{}, err
	}

	stepModels, err := s.Repository.GetStepsByJobID(id)
	if err != nil {
		return JobDto{}, err
	}

	return toJobDto(jobModel, stepModels)
}

func (s *Service) ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobDto], error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	jobsPagination, err := s.Repository.ListJobs(filter, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[JobDto]{}, err
	}

	response := utils.PaginationResponse[JobDto]{
		Items: make([]JobDto, 0, len(jobsPagination.Items)),
		Pagination: utils.Pagination{
			Page:     jobsPagination.Pagination.Page,
			PageSize: jobsPagination.Pagination.PageSize,
			HasNext:  jobsPagination.Pagination.HasNext,
			HasPrev:  jobsPagination.Pagination.HasPrev,
		},
	}

	for _, model := range jobsPagination.Items {
		stepModels, stepErr := s.Repository.GetStepsByJobID(model.ID)
		if stepErr != nil {
			return utils.PaginationResponse[JobDto]{}, stepErr
		}

		dto, mapErr := toJobDto(model, stepModels)
		if mapErr != nil {
			return utils.PaginationResponse[JobDto]{}, mapErr
		}

		response.Items = append(response.Items, dto)
	}

	return response, nil
}

func (s *Service) GetStepsByJobID(jobID int) ([]StepDto, error) {
	if jobID <= 0 {
		return nil, ErrInvalidJobID
	}

	_, err := s.Repository.GetJobByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}

	stepModels, err := s.Repository.GetStepsByJobID(jobID)
	if err != nil {
		return nil, err
	}

	steps := make([]StepDto, 0, len(stepModels))
	for _, stepModel := range stepModels {
		stepDto, mapErr := toStepDto(stepModel)
		if mapErr != nil {
			return nil, mapErr
		}
		steps = append(steps, stepDto)
	}

	return steps, nil
}

func (s *Service) CancelJob(jobID int) error {
	if jobID <= 0 {
		return ErrInvalidJobID
	}

	jobModel, err := s.Repository.GetJobByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrJobNotFound
		}
		return err
	}

	if jobModel.Status == "completed" || jobModel.Status == "failed" || jobModel.Status == "canceled" {
		return nil
	}

	endedAt := time.Now()
	cancelRequested := true
	dbContext := s.Repository.GetDbContext()
	if dbContext == nil {
		_, err = s.Repository.UpdateJobExecution(nil, jobID, "canceled", nil, &endedAt, &cancelRequested, nil)
	} else {
		err = dbContext.ExecTx(func(tx *sql.Tx) error {
			_, updateErr := s.Repository.UpdateJobExecution(tx, jobID, "canceled", nil, &endedAt, &cancelRequested, nil)
			return updateErr
		})
	}
	if err != nil {
		return err
	}

	return nil
}

func toJobDto(jobModel JobModel, stepModels []StepModel) (JobDto, error) {
	scope, err := decodeJSONMap(jobModel.Scope)
	if err != nil {
		return JobDto{}, fmt.Errorf("decode job scope: %w", err)
	}

	return JobDto{
		ID:              jobModel.ID,
		Type:            jobModel.Type,
		Priority:        jobModel.Priority,
		Scope:           scope,
		Status:          jobModel.Status,
		Progress:        calculateProgress(stepModels),
		CreatedAt:       jobModel.CreatedAt,
		StartedAt:       jobModel.StartedAt,
		EndedAt:         jobModel.EndedAt,
		CancelRequested: jobModel.CancelRequested,
		LastError:       jobModel.LastError,
	}, nil
}

func toStepDto(stepModel StepModel) (StepDto, error) {
	dependsOn := []int{}
	if len(stepModel.DependsOn) > 0 {
		if err := json.Unmarshal(stepModel.DependsOn, &dependsOn); err != nil {
			return StepDto{}, fmt.Errorf("decode step dependencies: %w", err)
		}
	}

	payload, err := decodeJSONAny(stepModel.Payload)
	if err != nil {
		return StepDto{}, fmt.Errorf("decode step payload: %w", err)
	}

	return StepDto{
		ID:          stepModel.ID,
		JobID:       stepModel.JobID,
		Type:        stepModel.Type,
		Status:      stepModel.Status,
		DependsOn:   dependsOn,
		Attempts:    stepModel.Attempts,
		MaxAttempts: stepModel.MaxAttempts,
		LastError:   stepModel.LastError,
		Progress:    stepModel.Progress,
		Payload:     payload,
		CreatedAt:   stepModel.CreatedAt,
		StartedAt:   stepModel.StartedAt,
		EndedAt:     stepModel.EndedAt,
	}, nil
}

func calculateProgress(stepModels []StepModel) JobProgressDto {
	progress := JobProgressDto{TotalSteps: len(stepModels)}

	if len(stepModels) == 0 {
		progress.Progress = 0
		return progress
	}

	for _, step := range stepModels {
		switch step.Status {
		case "completed":
			progress.CompletedSteps++
		case "running":
			progress.RunningSteps++
		case "failed":
			progress.FailedSteps++
		case "skipped":
			progress.SkippedSteps++
		case "canceled":
			progress.CanceledSteps++
		}
	}

	completedLike := progress.CompletedSteps + progress.SkippedSteps + progress.CanceledSteps
	progress.Progress = int(float64(completedLike) / float64(progress.TotalSteps) * 100)

	return progress
}

func decodeJSONMap(payload []byte) (JobScopeDto, error) {
	if len(payload) == 0 {
		return JobScopeDto{}, nil
	}

	result := JobScopeDto{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func decodeJSONAny(payload []byte) (any, error) {
	if len(payload) == 0 {
		return nil, nil
	}

	var result any
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}

	return result, nil
}
