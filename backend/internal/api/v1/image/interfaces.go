package image

import (
	"database/sql"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

// RepositoryInterface is the data-access contract for the image domain.
type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetImageMetadataByID(id int) (MetadataModel, error)
	UpsertImageMetadata(tx *sql.Tx, metadata MetadataModel) (MetadataModel, error)
	DeleteImageMetadata(id int) error
	GetImages(page int, pageSize int, groupBy files.ImageGroupBy) (utils.PaginationResponse[files.FileModel], error)
}

// ServiceInterface is the business-logic contract for the image domain.
type ServiceInterface interface {
	GetImages(page int, pageSize int, groupBy files.ImageGroupBy) (utils.PaginationResponse[files.FileDto], error)
}
