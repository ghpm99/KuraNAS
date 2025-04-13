package files

import (
	"database/sql"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *sql.DB
	GetFiles(filter FileFilter, pagination utils.Pagination) (utils.PaginationResponse[FileModel], error)
	GetFilesByPath(path string) ([]FileModel, error)
	GetFileByNameAndPath(name string, path string) (FileModel, error)
	CreateFile(transaction *sql.Tx, file FileModel) (FileModel, error)
	UpdateFile(transaction *sql.Tx, file FileModel) (bool, error)
	GetPathByFileId(fileId int) (string, error)
}

type ServiceInterface interface {
	GetFiles(filter FileFilter, fileDtoList *utils.PaginationResponse[FileDto]) error
	GetFilesByPath(path string) ([]FileDto, error)
	GetFileByNameAndPath(name string, path string) (FileDto, error)
	CreateFile(fileDto FileDto) (FileDto, error)
	UpdateFile(file FileDto) (bool, error)
	ScanFilesTask(data string)
	ScanDirTask(data string)
}
