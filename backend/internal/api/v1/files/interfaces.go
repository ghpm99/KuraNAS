package files

import (
	"database/sql"
	"mime/multipart"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"time"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	CreateFile(transaction *sql.Tx, file FileModel) (FileModel, error)
	GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	GetFileByID(id int) (FileModel, bool, error)
	GetFilesByNameAndPath(name string, path string, limit int) ([]FileModel, error)
	GetActiveChildrenByParentPath(parentPath string, category FileCategory, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	GetActiveFilesByPath(path string, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	GetActiveFiles(page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	GetFilesByPathPrefix(prefix string, page int, pageSize int) (utils.PaginationResponse[FileModel], error)
	GetFileStatByPath(path string) (FileStat, bool, error)
	UpdateFile(transaction *sql.Tx, file FileModel) (bool, error)
	UpdateDescendantPaths(transaction *sql.Tx, oldPath string, newPath string) (int64, error)
	MarkDeletedSubtree(transaction *sql.Tx, path string, deletedAt time.Time) (int64, error)
	GetDirectoryContentCount(fileId int, parentPath string) (int, error)
	GetCountByType(fileType FileType) (int, error)
	GetTotalSpaceUsed() (int, error)
	GetReportSizeByFormat() ([]SizeReportModel, error)
	GetTopFilesBySize(limit int) ([]FileModel, error)
	GetDuplicateFiles(page int, pageSize int) (utils.PaginationResponse[DuplicateFilesModel], error)
}

type ServiceInterface interface {
	CreateFile(fileDto FileDto) (fileDtoResult FileDto, err error)
	GetFileByNameAndPath(name string, path string) (FileDto, error)
	GetFileById(id int) (FileDto, error)
	GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileDto], error)
	GetFileStatByPath(path string) (FileStat, bool, error)
	UpdateFile(file FileDto) (result bool, err error)
	ScanFilesTask(data string)
	ScanDirTask(data string)
	UpdateCheckSum(fileId int) error
	CreateUploadProcessJob(paths []string) (int, error)
	GetFileThumbnail(fileDto FileDto, width, height int) ([]byte, error)
	GetFileBlobById(fileId int) (FileBlob, error)
	GetTotalSpaceUsed() (int, error)
	GetTotalFiles() (int, error)
	GetTotalDirectory() (int, error)
	GetReportSizeByFormat() ([]SizeReportDto, error)
	GetTopFilesBySize(limit int) ([]FileDto, error)
	GetDuplicateFiles(page int, pageSize int) (DuplicateFileReportDto, error)
	CheckFileExists(fileId int) bool
	CheckFileExistsByPath(path string) bool
	DeleteFile(file FileDto, bySystem bool) error
	UploadFiles(targetFolderID int, files []*multipart.FileHeader) (UploadFilesResult, error)
	CreateFolder(parentID *int, name string) (string, error)
	MoveFile(sourceID int, destinationFolderID *int, destinationPath string) (string, error)
	DeleteFileFromDisk(id int) error
	RenameFile(id int, newName string) (string, error)
	CopyFile(sourceID int, destinationFolderID *int, destinationPath string, newName string) (string, error)
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
