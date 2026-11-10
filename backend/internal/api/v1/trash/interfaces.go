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
	GetItems(listing Listing, page int, pageSize int) (utils.PaginationResponse[TrashItemModel], error)
	GetItemByID(id int) (TrashItemModel, bool, error)
	GetExpiredItems(cutoff time.Time) ([]TrashItemModel, error)
	GetAllItems() ([]TrashItemModel, error)
	DeleteItem(tx *sql.Tx, id int) error
	CountOverlappingItems(originalPath string) (int, error)
	GetSummary() (TrashSummaryDto, error)
	GetRetentionDays() (int, bool, error)
	SetRetentionDays(days int) error
}

type ServiceInterface interface {
	MoveToTrash(originalPath string, size int64) error
	MoveToTrashFrom(logicalPath string, contentPath string, size int64) error
	GetItems(listing Listing, page int, pageSize int) (utils.PaginationResponse[TrashItemDto], error)
	GetSummary() (TrashSummaryDto, error)
	RestoreItem(id int) (string, error)
	DeleteItemPermanently(id int) error
	RestoreItems(ids []int) ([]int, []ItemFailure)
	DeleteItemsPermanently(ids []int) ([]int, []ItemFailure)
	EmptyTrash() (int, error)
	PurgeExpired() (int, error)
	GetRetentionDays() (int, error)
	SetRetentionDays(days int) error
}

// FilesIndexInterface is the slice of the files domain the trash needs: after
// a restore, the soft-deleted home_file rows of the subtree must come back and
// the destination directory gets rescanned; after a purge, those rows are
// dropped for good. Declared here (not imported from files) so the dependency
// stays one-directional at the package level.
type FilesIndexInterface interface {
	RestoreSubtree(path string) error
	PurgeDeletedSubtree(path string) error
	ScanDirTask(path string)
}
