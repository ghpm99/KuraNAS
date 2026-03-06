package jobs

import (
	"database/sql"
	"fmt"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/jobs"
	"nas-go/api/pkg/utils"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(database *database.DbContext) *Repository {
	return &Repository{database}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
}

func (r *Repository) CreateJob(tx *sql.Tx, job JobModel) (JobModel, error) {
	_, err := tx.Exec(
		queries.InsertJobQuery,
		job.ID,
		job.Type,
		job.Priority,
		job.ScopeJSON,
		job.Status,
		job.CancelRequested,
		job.LastError,
	)
	if err != nil {
		return job, fmt.Errorf("CreateJob: %w", err)
	}

	return job, nil
}

func (r *Repository) CreateStep(tx *sql.Tx, step StepModel) (StepModel, error) {
	_, err := tx.Exec(
		queries.InsertStepQuery,
		step.ID,
		step.JobID,
		step.Type,
		step.Status,
		step.DependsOnJSON,
		step.Attempts,
		step.MaxAttempts,
		step.LastError,
		step.Progress,
		step.PayloadJSON,
	)
	if err != nil {
		return step, fmt.Errorf("CreateStep: %w", err)
	}

	return step, nil
}

func (r *Repository) GetJobByID(id string) (JobModel, error) {
	var job JobModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetJobByIDQuery, id).Scan(
			&job.ID,
			&job.Type,
			&job.Priority,
			&job.ScopeJSON,
			&job.Status,
			&job.CreatedAt,
			&job.StartedAt,
			&job.EndedAt,
			&job.CancelRequested,
			&job.LastError,
		)
	})
	if err != nil {
		return job, fmt.Errorf("GetJobByID: %w", err)
	}

	return job, nil
}

func (r *Repository) ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error) {
	paginationResponse := utils.PaginationResponse[JobModel]{
		Items: []JobModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		!filter.Status.HasValue,
		filter.Status.Value,
		!filter.Type.HasValue,
		filter.Type.Value,
		!filter.Priority.HasValue,
		filter.Priority.Value,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ListJobsQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var job JobModel
			if err := rows.Scan(
				&job.ID,
				&job.Type,
				&job.Priority,
				&job.ScopeJSON,
				&job.Status,
				&job.CreatedAt,
				&job.StartedAt,
				&job.EndedAt,
				&job.CancelRequested,
				&job.LastError,
			); err != nil {
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, job)
		}

		return nil
	})
	if err != nil {
		return paginationResponse, fmt.Errorf("ListJobs: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetStepsByJobID(jobID string) ([]StepModel, error) {
	steps := []StepModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetStepsByJobIDQuery, jobID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var step StepModel
			if err := rows.Scan(
				&step.ID,
				&step.JobID,
				&step.Type,
				&step.Status,
				&step.DependsOnJSON,
				&step.Attempts,
				&step.MaxAttempts,
				&step.LastError,
				&step.Progress,
				&step.PayloadJSON,
				&step.CreatedAt,
				&step.StartedAt,
				&step.EndedAt,
			); err != nil {
				return err
			}
			steps = append(steps, step)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("GetStepsByJobID: %w", err)
	}

	return steps, nil
}

func (r *Repository) UpdateJobStatus(
	tx *sql.Tx,
	id string,
	fromStatus string,
	toStatus string,
	startedAt *time.Time,
	endedAt *time.Time,
	lastError string,
) (bool, error) {
	result, err := tx.Exec(
		queries.UpdateJobStatusQuery,
		id,
		fromStatus,
		toStatus,
		startedAt,
		endedAt,
		lastError,
	)
	if err != nil {
		return false, fmt.Errorf("UpdateJobStatus: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("UpdateJobStatus rows affected: %w", err)
	}

	return rowsAffected == 1, nil
}

func (r *Repository) UpdateStepStatus(
	tx *sql.Tx,
	id string,
	fromStatus string,
	toStatus string,
	startedAt *time.Time,
	endedAt *time.Time,
	lastError string,
) (bool, error) {
	result, err := tx.Exec(
		queries.UpdateStepStatusQuery,
		id,
		fromStatus,
		toStatus,
		startedAt,
		endedAt,
		lastError,
	)
	if err != nil {
		return false, fmt.Errorf("UpdateStepStatus: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("UpdateStepStatus rows affected: %w", err)
	}

	return rowsAffected == 1, nil
}

func (r *Repository) UpdateStepExecution(
	tx *sql.Tx,
	id string,
	attempts int,
	lastError string,
	progress int,
	startedAt *time.Time,
	endedAt *time.Time,
) (bool, error) {
	result, err := tx.Exec(
		queries.UpdateStepExecutionQuery,
		id,
		attempts,
		lastError,
		progress,
		startedAt,
		endedAt,
	)
	if err != nil {
		return false, fmt.Errorf("UpdateStepExecution: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("UpdateStepExecution rows affected: %w", err)
	}

	return rowsAffected == 1, nil
}
