package takeout

import (
	"database/sql"
	"mime/multipart"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type ServiceInterface interface {
	InitUpload(dto InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error)
	UploadChunk(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error
	CompleteUpload(dto CompleteTakeoutUploadDto) (TakeoutImportResultDto, error)
	GetUploadStatus(uploadID string) (TakeoutUploadStatusDto, error)
	CancelUpload(uploadID string) error
	ListImports(page int, pageSize int) (utils.PaginationResponse[TakeoutImportDto], error)
}

type UploadJobDispatcherInterface interface {
	GetDbContext() *database.DbContext
	CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error)
	CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error)
}

type LibraryResolverInterface interface {
	GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error)
}

type ImportCreatorInterface interface {
	CreateQueuedImport(transaction *sql.Tx, jobID int, fileName string, size int64) (int, error)
}

type ImportRecorderInterface interface {
	MarkImportRunning(jobID int) error
	UpdateImportProgress(jobID int, processedEntries int, totalEntries int) error
	CloseImportCompleted(jobID int, summary ImportSummary) error
	CloseImportFailed(jobID int, failureMessage string) error
}

type SidecarRecorderInterface interface {
	SaveSidecars(sidecars []PhotoSidecar) error
}

type SidecarFinderInterface interface {
	FindSidecarByPath(path string) (*PhotoSidecar, error)
}

type ImportListerInterface interface {
	ListImports(page int, pageSize int) (utils.PaginationResponse[TakeoutImportDto], error)
}
