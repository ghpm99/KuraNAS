package captures

import (
	"database/sql"
	"mime/multipart"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	CreateCapture(transaction *sql.Tx, capture CaptureModel) (CaptureModel, error)
	GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureModel], error)
	GetCaptureByID(id int) (CaptureModel, error)
	GetCapturePlatforms() ([]string, error)
	GetCaptureByEpisodeKey(episodeKey string) (CaptureModel, bool, error)
	UpdateCapturePromotion(transaction *sql.Tx, capture CaptureModel) error
	MarkCapturePromoting(transaction *sql.Tx, id int) error
	MarkCapturePromotionFailed(transaction *sql.Tx, id int, receivedFileName string, receivedFilePath string, failureReason string) error
	DeleteCapture(transaction *sql.Tx, id int) error
}

type UploadJobDispatcherInterface interface {
	CreateUploadProcessJob(paths []string) (int, error)
	CreateCaptureProcessJob(captureID int) (int, error)
}

// LibrariesProviderInterface is the slice of the libraries domain the promotion
// needs: the destination root for a category (videos).
type LibrariesProviderInterface interface {
	GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error)
}

// FilesProviderInterface is the slice of the files domain the promotion needs:
// pre-register the home_file stub at the final path, and hard-delete it on a
// move-failure rollback.
type FilesProviderInterface interface {
	CreateFile(fileDto files.FileDto) (files.FileDto, error)
	DeleteFileRecord(id int) error
}

type ServiceInterface interface {
	UploadCapture(file *multipart.FileHeader, dto CreateCaptureDto) (CaptureDto, error)
	InitCaptureUpload(dto InitCaptureUploadDto) (InitCaptureUploadResultDto, error)
	UploadCaptureChunk(file *multipart.FileHeader, dto UploadCaptureChunkDto) error
	CompleteCaptureUpload(dto CompleteCaptureUploadDto) (CaptureDto, error)
	GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureDto], error)
	GetCaptureByID(id int) (CaptureDto, error)
	GetCapturePlatforms() ([]string, error)
	PromoteCapture(captureID int) error
	RetryCapturePromotion(captureID int) (int, error)
	DeleteCapture(id int) (bool, error)
}
