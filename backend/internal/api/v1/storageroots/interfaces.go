package storageroots

import (
	"database/sql"

	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetAll() ([]StorageRootModel, error)
	GetByID(id int) (StorageRootModel, bool, error)
	Create(tx *sql.Tx, model StorageRootModel) (StorageRootModel, error)
	Update(tx *sql.Tx, model StorageRootModel) (StorageRootModel, error)
	Delete(tx *sql.Tx, id int) error
}

type ServiceInterface interface {
	GetRoots() ([]StorageRootDto, error)
	CreateRoot(request CreateStorageRootDto) (StorageRootDto, error)
	UpdateRoot(id int, request UpdateStorageRootDto) (StorageRootDto, error)
	DeleteRoot(id int) error
	// ReloadRegistry re-reads the table into the in-memory registry (boot,
	// and after every CRUD change). Seeds ENTRY_POINT when the table is empty.
	ReloadRegistry() error
}

// IndexTrigger is the slice of the files domain the roots need: a newly
// registered root must be indexed without waiting for the next boot.
type IndexTrigger interface {
	ScanDirTask(path string)
}
