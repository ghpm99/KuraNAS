package jobs

import (
	"database/sql"
	"time"

	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	CreateJob(tx *sql.Tx, job JobModel) (JobModel, error)
	CreateStep(tx *sql.Tx, step StepModel) (StepModel, error)
	GetJobByID(id string) (JobModel, error)
	ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error)
	GetStepsByJobID(jobID string) ([]StepModel, error)
	UpdateJobStatus(tx *sql.Tx, id string, fromStatus string, toStatus string, startedAt *time.Time, endedAt *time.Time, lastError string) (bool, error)
	UpdateStepStatus(tx *sql.Tx, id string, fromStatus string, toStatus string, startedAt *time.Time, endedAt *time.Time, lastError string) (bool, error)
	UpdateStepExecution(tx *sql.Tx, id string, attempts int, lastError string, progress int, startedAt *time.Time, endedAt *time.Time) (bool, error)
	RequestJobCancel(tx *sql.Tx, id string) (bool, error)
	RequestJobCancelCascade(tx *sql.Tx, id string) (bool, error)
}

type ServiceInterface interface {
	GetJobByID(id string) (JobSummaryDto, error)
	ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobSummaryDto], error)
	GetStepsByJobID(jobID string) ([]StepDto, error)
	CancelJob(id string) (JobSummaryDto, error)
}
