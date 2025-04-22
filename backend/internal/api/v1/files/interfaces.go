package files

import (
	"database/sql"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *sql.DB
	CreateFile(transaction *sql.Tx, file FileModel) (FileModel, error)
	GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	UpdateFile(transaction *sql.Tx, file FileModel) (bool, error)
}

type ServiceInterface interface {
	CreateFile(fileDto FileDto) (FileDto, error)
	GetFileByNameAndPath(name string, path string) (FileDto, error)
	GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	UpdateFile(file FileDto) (bool, error)
	ScanFilesTask(data string)
	ScanDirTask(data string)
}
