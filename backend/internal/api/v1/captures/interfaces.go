package captures

import (
	"database/sql"
	"mime/multipart"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	CreateCapture(transaction *sql.Tx, capture CaptureModel) (CaptureModel, error)
	GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureModel], error)
	GetCaptureByID(id int) (CaptureModel, error)
	DeleteCapture(transaction *sql.Tx, id int) error
}

type UploadJobDispatcherInterface interface {
	CreateUploadProcessJob(paths []string) (int, error)
}

type ServiceInterface interface {
	UploadCapture(file *multipart.FileHeader, dto CreateCaptureDto) (CaptureDto, error)
	GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureDto], error)
	GetCaptureByID(id int) (CaptureDto, error)
	DeleteCapture(id int) error
}
