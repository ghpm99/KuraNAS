package watchfolders

import (
	"database/sql"
	"nas-go/api/pkg/database"
	"time"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetAll() ([]WatchFolderModel, error)
	GetByID(id int) (WatchFolderModel, error)
	Create(tx *sql.Tx, model WatchFolderModel) (WatchFolderModel, error)
	Update(tx *sql.Tx, model WatchFolderModel) (WatchFolderModel, error)
	Delete(tx *sql.Tx, id int) error
	UpdateLastScan(tx *sql.Tx, id int, lastScanAt time.Time) error
}

type ServiceInterface interface {
	GetWatchFolders() ([]WatchFolderDto, error)
	CreateWatchFolder(dto CreateWatchFolderDto) (WatchFolderDto, error)
	UpdateWatchFolder(id int, dto UpdateWatchFolderDto) (WatchFolderDto, error)
	DeleteWatchFolder(id int) error
	GetEnabledWatchFolders() ([]WatchFolderModel, error)
	UpdateWatchFolderLastScan(id int, lastScanAt time.Time) error
}
