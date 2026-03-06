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
	var startedAt sql.NullTime
	var endedAt sql.NullTime

	err := tx.QueryRow(
		queries.InsertJobQuery,
		job.Type,
		job.Priority,
		job.Scope,
		job.Status,
		job.CancelRequested,
		job.LastError,
	).Scan(
		&job.ID,
		&job.Type,
		&job.Priority,
		&job.Scope,
		&job.Status,
		&job.CreatedAt,
		&startedAt,
		&endedAt,
		&job.CancelRequested,
		&job.LastError,
	)
	if err != nil {
		return job, fmt.Errorf("CreateJob: %w", err)
	}

	if startedAt.Valid {
		value := startedAt.Time
		job.StartedAt = &value
	}
	if endedAt.Valid {
		value := endedAt.Time
		job.EndedAt = &value
	}

	return job, nil
}

func (r *Repository) CreateStep(tx *sql.Tx, step StepModel) (StepModel, error) {
	var startedAt sql.NullTime
	var endedAt sql.NullTime

	err := tx.QueryRow(
		queries.InsertStepQuery,
		step.JobID,
		step.Type,
		step.Status,
		step.DependsOn,
		step.Attempts,
		step.MaxAttempts,
		step.LastError,
		step.Progress,
		step.Payload,
	).Scan(
		&step.ID,
		&step.JobID,
		&step.Type,
		&step.Status,
		&step.DependsOn,
		&step.Attempts,
		&step.MaxAttempts,
		&step.LastError,
		&step.Progress,
		&step.Payload,
		&step.CreatedAt,
		&startedAt,
		&endedAt,
	)
	if err != nil {
		return step, fmt.Errorf("CreateStep: %w", err)
	}

	if startedAt.Valid {
		value := startedAt.Time
		step.StartedAt = &value
	}
	if endedAt.Valid {
		value := endedAt.Time
		step.EndedAt = &value
	}

	return step, nil
}

func (r *Repository) GetJobByID(id int) (JobModel, error) {
	var job JobModel
	var startedAt sql.NullTime
	var endedAt sql.NullTime

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetJobByIDQuery, id).Scan(
			&job.ID,
			&job.Type,
			&job.Priority,
			&job.Scope,
			&job.Status,
			&job.CreatedAt,
			&startedAt,
			&endedAt,
			&job.CancelRequested,
			&job.LastError,
		)
	})
	if err != nil {
		return job, fmt.Errorf("GetJobByID: %w", err)
	}

	if startedAt.Valid {
		value := startedAt.Time
		job.StartedAt = &value
	}
	if endedAt.Valid {
		value := endedAt.Time
		job.EndedAt = &value
	}

	return job, nil
}

func (r *Repository) ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error) {
	response := utils.PaginationResponse[JobModel]{
		Items: []JobModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
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
		rows, queryErr := tx.Query(queries.ListJobsQuery, args...)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		for rows.Next() {
			var job JobModel
			var startedAt sql.NullTime
			var endedAt sql.NullTime

			if scanErr := rows.Scan(
				&job.ID,
				&job.Type,
				&job.Priority,
				&job.Scope,
				&job.Status,
				&job.CreatedAt,
				&startedAt,
				&endedAt,
				&job.CancelRequested,
				&job.LastError,
			); scanErr != nil {
				return scanErr
			}

			if startedAt.Valid {
				value := startedAt.Time
				job.StartedAt = &value
			}
			if endedAt.Valid {
				value := endedAt.Time
				job.EndedAt = &value
			}

			response.Items = append(response.Items, job)
		}

		return nil
	})
	if err != nil {
		return response, fmt.Errorf("ListJobs: %w", err)
	}

	response.UpdatePagination()
	return response, nil
}

func (r *Repository) GetStepsByJobID(jobID int) ([]StepModel, error) {
	response := []StepModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, queryErr := tx.Query(queries.GetStepsByJobIDQuery, jobID)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		for rows.Next() {
			var step StepModel
			var startedAt sql.NullTime
			var endedAt sql.NullTime

			if scanErr := rows.Scan(
				&step.ID,
				&step.JobID,
				&step.Type,
				&step.Status,
				&step.DependsOn,
				&step.Attempts,
				&step.MaxAttempts,
				&step.LastError,
				&step.Progress,
				&step.Payload,
				&step.CreatedAt,
				&startedAt,
				&endedAt,
			); scanErr != nil {
				return scanErr
			}

			if startedAt.Valid {
				value := startedAt.Time
				step.StartedAt = &value
			}
			if endedAt.Valid {
				value := endedAt.Time
				step.EndedAt = &value
			}

			response = append(response, step)
		}

		return nil
	})
	if err != nil {
		return response, fmt.Errorf("GetStepsByJobID: %w", err)
	}

	return response, nil
}

func (r *Repository) UpdateJobExecution(
	tx *sql.Tx,
	jobID int,
	status string,
	startedAt *time.Time,
	endedAt *time.Time,
	cancelRequested *bool,
	lastError *string,
) (bool, error) {
	result, err := tx.Exec(
		queries.UpdateJobExecutionQuery,
		status,
		startedAt,
		endedAt,
		cancelRequested,
		lastError,
		jobID,
	)
	if err != nil {
		return false, fmt.Errorf("UpdateJobExecution: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("UpdateJobExecution rows affected: %w", err)
	}

	if affectedRows > 1 {
		return false, fmt.Errorf("UpdateJobExecution affected more than one row")
	}

	return affectedRows == 1, nil
}

func (r *Repository) UpdateStepExecution(
	tx *sql.Tx,
	stepID int,
	status string,
	progress int,
	attempts int,
	startedAt *time.Time,
	endedAt *time.Time,
	lastError *string,
) (bool, error) {
	result, err := tx.Exec(
		queries.UpdateStepExecutionQuery,
		status,
		progress,
		attempts,
		startedAt,
		endedAt,
		lastError,
		stepID,
	)
	if err != nil {
		return false, fmt.Errorf("UpdateStepExecution: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("UpdateStepExecution rows affected: %w", err)
	}

	if affectedRows > 1 {
		return false, fmt.Errorf("UpdateStepExecution affected more than one row")
	}

	return affectedRows == 1, nil
}
