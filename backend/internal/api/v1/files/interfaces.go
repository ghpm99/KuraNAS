package files

import (
	"database/sql"
	"image"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *sql.DB
	CreateFile(transaction *sql.Tx, file FileModel) (FileModel, error)
	GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	UpdateFile(transaction *sql.Tx, file FileModel) (bool, error)
	GetDirectoryContentCount(fileId int, parentPath string) (int, error)
	GetCountByType(fileType FileType) (int, error)
}

type ServiceInterface interface {
	CreateFile(fileDto FileDto) (fileDtoResult FileDto, err error)
	GetFileByNameAndPath(name string, path string) (FileDto, error)
	GetFileById(id int) (FileDto, error)
	GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	UpdateFile(file FileDto) (result bool, err error)
	ScanFilesTask(data string)
	ScanDirTask(data string)
	UpdateCheckSumTask(fileId int)
	GetFileThumbnail(fileDto FileDto, width int) (image.Image, error)
	GetFileBlobById(fileId int) (FileBlob, error)
	GetTotalSpaceUsed() (int64, error)
	GetTotalFiles() (int, error)
	GetTotalDirectory() (int, error)
}

type RecentFileRepositoryInterface interface {
	Upsert(ip string, fileID int) error
	DeleteOld(ip string, keep int) error
	GetRecentFiles(page int, pageSize int) ([]RecentFileModel, error)
	Delete(ip string, fileID int) error
	GetByFileID(fileID int) ([]RecentFileModel, error)
}

type RecentFileServiceInterface interface {
	RegisterAccess(ip string, fileID int, keep int) error
	GetRecentFiles(page int, pageSize int) ([]RecentFileDto, error)
	DeleteRecentFile(ip string, fileID int) error
	GetRecentAccessByFileID(fileID int) ([]RecentFileDto, error)
}
