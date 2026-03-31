package takeout

import (
	"database/sql"
	"mime/multipart"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/pkg/database"
)

type ServiceInterface interface {
	InitUpload(dto InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error)
	UploadChunk(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error
	CompleteUpload(dto CompleteTakeoutUploadDto) (TakeoutImportResultDto, error)
}

type UploadJobDispatcherInterface interface {
	GetDbContext() *database.DbContext
	CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error)
	CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error)
}

type LibraryResolverInterface interface {
	GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error)
}
