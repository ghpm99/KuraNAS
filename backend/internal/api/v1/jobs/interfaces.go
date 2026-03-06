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
	GetJobByID(id int) (JobModel, error)
	ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobModel], error)
	GetStepsByJobID(jobID int) ([]StepModel, error)
	UpdateJobExecution(tx *sql.Tx, jobID int, status string, startedAt *time.Time, endedAt *time.Time, cancelRequested *bool, lastError *string) (bool, error)
	UpdateStepExecution(tx *sql.Tx, stepID int, status string, progress int, attempts int, startedAt *time.Time, endedAt *time.Time, lastError *string) (bool, error)
}

type ServiceInterface interface {
	GetJobByID(id int) (JobDto, error)
	ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobDto], error)
	GetStepsByJobID(jobID int) ([]StepDto, error)
	CancelJob(jobID int) error
}

type JobFilter struct {
	Status   utils.Optional[string]
	Type     utils.Optional[string]
	Priority utils.Optional[string]
}
