package image

import (
	"database/sql"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

// RepositoryInterface is the data-access contract for the image domain.
type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetImageMetadataByID(id int) (MetadataModel, error)
	UpsertImageMetadata(tx *sql.Tx, metadata MetadataModel) (MetadataModel, error)
	DeleteImageMetadata(id int) error
	GetImages(page int, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[files.FileModel], error)
	CountPendingAIClassification(confidenceThreshold float64) (int, error)
	ListPendingAIClassification(confidenceThreshold float64, afterFileID int, limit int) ([]PendingImageClassification, error)
}

// ServiceInterface is the business-logic contract for the image domain.
type ServiceInterface interface {
	GetImages(page int, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[files.FileDto], error)
	GetPendingAIClassificationCount() (int, error)
	EnqueueClassificationBackfill() (int, error)
}

// JobEnqueuer is the slice of the jobs repository the image service needs to
// enqueue a classification-backfill job and guard against duplicates. Declared
// locally so the domain depends on a tiny capability, not the whole repository.
type JobEnqueuer interface {
	GetDbContext() *database.DbContext
	CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error)
	CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error)
	ListJobs(filter jobs.JobFilter, page int, pageSize int) (utils.PaginationResponse[jobs.JobModel], error)
}
