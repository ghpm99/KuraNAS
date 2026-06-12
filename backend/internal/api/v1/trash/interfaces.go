package trash

import (
	"database/sql"
	"time"

	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	CreateItem(tx *sql.Tx, item TrashItemModel) (TrashItemModel, error)
	GetItems(page int, pageSize int) (utils.PaginationResponse[TrashItemModel], error)
	GetItemByID(id int) (TrashItemModel, bool, error)
	GetExpiredItems(cutoff time.Time) ([]TrashItemModel, error)
	GetAllItems() ([]TrashItemModel, error)
	DeleteItem(tx *sql.Tx, id int) error
	GetRetentionDays() (int, bool, error)
	SetRetentionDays(days int) error
}

type ServiceInterface interface {
	MoveToTrash(originalPath string, size int64) error
	GetItems(page int, pageSize int) (utils.PaginationResponse[TrashItemDto], error)
	RestoreItem(id int) (string, error)
	DeleteItemPermanently(id int) error
	EmptyTrash() (int, error)
	PurgeExpired() (int, error)
	GetRetentionDays() (int, error)
	SetRetentionDays(days int) error
}

// FilesIndexInterface is the slice of the files domain the trash needs: after
// a restore, the soft-deleted home_file rows of the subtree must come back and
// the destination directory gets rescanned. Declared here (not imported from
// files) so the dependency stays one-directional at the package level.
type FilesIndexInterface interface {
	RestoreSubtree(path string) error
	ScanDirTask(path string)
}
