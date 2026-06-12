package trash

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"nas-go/api/internal/roots"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

// DirName is the trash directory living at the top of each storage root. The
// scan walker and the filesystem watcher must ignore it explicitly, so trashed
// content never re-enters the index.
const DirName = ".kuranas-trash"

// DefaultRetentionDays applies when no retention was ever configured.
const DefaultRetentionDays = 30

type Service struct {
	Repository RepositoryInterface
	FilesIndex FilesIndexInterface
}

func NewService(repository RepositoryInterface, filesIndex FilesIndexInterface) *Service {
	return &Service{Repository: repository, FilesIndex: filesIndex}
}

// Dir returns the absolute path of the primary storage root's trash directory.
func Dir() (string, error) {
	primary, ok := roots.Primary()
	if !ok {
		return "", fmt.Errorf("entry point not configured")
	}
	return filepath.Join(primary.Path, DirName), nil
}

// dirFor returns the trash directory of the storage root that owns path —
// trashing stays inside the same volume, so os.Rename remains free.
func dirFor(originalPath string) (string, error) {
	if owner, ok := roots.OwnerOf(originalPath); ok {
		return filepath.Join(owner.Path, DirName), nil
	}
	return Dir()
}

// IsInsideTrash reports whether path lives under the trash directory of root.
// It is the shared guard used by the scan walker and the watcher.
func IsInsideTrash(root string, path string) bool {
	trashDir := filepath.Join(filepath.Clean(root), DirName)
	cleanPath := filepath.Clean(path)
	return cleanPath == trashDir || strings.HasPrefix(cleanPath, trashDir+string(filepath.Separator))
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
}

// MoveToTrash relocates an absolute path into the trash directory of its
// owning storage root (same volume, so os.Rename is free) and registers it
// for later restore. The caller (files domain) has already validated the path
// against the storage roots and soft-deletes the home_file rows itself.
func (s *Service) MoveToTrash(originalPath string, size int64) error {
	trashDir, err := dirFor(originalPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return fmt.Errorf("create trash dir: %w", err)
	}

	trashPath, err := uniqueTrashPath(trashDir, filepath.Base(originalPath))
	if err != nil {
		return err
	}

	if err := os.Rename(originalPath, trashPath); err != nil {
		return fmt.Errorf("move to trash: %w", err)
	}

	err = s.withTransaction(func(tx *sql.Tx) error {
		_, createErr := s.Repository.CreateItem(tx, TrashItemModel{
			OriginalPath: originalPath,
			TrashPath:    trashPath,
			Size:         size,
			DeletedAt:    time.Now(),
		})
		return createErr
	})
	if err != nil {
		// The bytes are safe but unregistered — put them back so the user
		// never has data that exists nowhere in the UI.
		if renameErr := os.Rename(trashPath, originalPath); renameErr != nil {
			log.Printf("trash: failed to undo move of %q after registry error: %v", trashPath, renameErr)
		}
		return fmt.Errorf("register trash item: %w", err)
	}

	return nil
}

// uniqueTrashPath builds a collision-free destination name inside the trash
// dir: deleting two files with the same name must not overwrite each other.
func uniqueTrashPath(trashDir string, baseName string) (string, error) {
	candidate := filepath.Join(trashDir, fmt.Sprintf("%s.%d", baseName, time.Now().UnixNano()))
	for attempt := 0; attempt < 100; attempt++ {
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
		candidate = filepath.Join(trashDir, fmt.Sprintf("%s.%d.%d", baseName, time.Now().UnixNano(), attempt))
	}
	return "", fmt.Errorf("could not allocate a unique trash name for %q", baseName)
}

func (s *Service) GetItems(page int, pageSize int) (utils.PaginationResponse[TrashItemDto], error) {
	models, err := s.Repository.GetItems(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[TrashItemDto]{}, err
	}

	items := make([]TrashItemDto, 0, len(models.Items))
	for index := range models.Items {
		items = append(items, models.Items[index].ToDto())
	}

	return utils.PaginationResponse[TrashItemDto]{
		Items:      items,
		Pagination: models.Pagination,
	}, nil
}

// RestoreItem moves the item back to its original path and revives the
// soft-deleted home_file rows of the subtree. Returns the restored path.
func (s *Service) RestoreItem(id int) (string, error) {
	item, found, err := s.Repository.GetItemByID(id)
	if err != nil {
		return "", err
	}
	if !found {
		return "", ErrItemNotFound
	}

	if _, statErr := os.Stat(item.OriginalPath); statErr == nil {
		return "", ErrRestoreConflict
	}

	if err := os.MkdirAll(filepath.Dir(item.OriginalPath), 0755); err != nil {
		return "", fmt.Errorf("recreate parent dir: %w", err)
	}
	if err := os.Rename(item.TrashPath, item.OriginalPath); err != nil {
		return "", fmt.Errorf("restore from trash: %w", err)
	}

	if err := s.deleteRegistryRow(item.ID); err != nil {
		// The restore itself succeeded; a stale registry row is repairable
		// (its trash_path no longer exists), so don't fail the request.
		log.Printf("trash: restored %q but failed to drop registry row %d: %v", item.OriginalPath, item.ID, err)
	}

	if s.FilesIndex != nil {
		if err := s.FilesIndex.RestoreSubtree(item.OriginalPath); err != nil {
			log.Printf("trash: index restore for %q failed (rescan will reconcile): %v", item.OriginalPath, err)
		}
		s.FilesIndex.ScanDirTask(filepath.Dir(item.OriginalPath))
	}

	return item.OriginalPath, nil
}

func (s *Service) deleteRegistryRow(id int) error {
	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.DeleteItem(tx, id)
	})
}

// removeItemFromDisk permanently deletes the bytes of one registered item.
// It refuses to touch anything outside a trash directory (any root's).
func (s *Service) removeItemFromDisk(item TrashItemModel) error {
	cleanPath := filepath.Clean(item.TrashPath)
	separator := string(filepath.Separator)
	if !strings.Contains(cleanPath, separator+DirName+separator) {
		return fmt.Errorf("trash item %d points outside the trash dir: %q", item.ID, item.TrashPath)
	}
	if err := os.RemoveAll(cleanPath); err != nil {
		return fmt.Errorf("purge trash item %d: %w", item.ID, err)
	}
	return nil
}

func (s *Service) DeleteItemPermanently(id int) error {
	item, found, err := s.Repository.GetItemByID(id)
	if err != nil {
		return err
	}
	if !found {
		return ErrItemNotFound
	}

	if err := s.removeItemFromDisk(item); err != nil {
		return err
	}

	return s.deleteRegistryRow(item.ID)
}

// purgeItems removes a batch from disk and registry, returning how many items
// were actually purged. One broken item does not stop the rest.
func (s *Service) purgeItems(items []TrashItemModel) int {
	purged := 0
	for _, item := range items {
		if err := s.removeItemFromDisk(item); err != nil {
			log.Printf("trash: %v", err)
			continue
		}
		if err := s.deleteRegistryRow(item.ID); err != nil {
			log.Printf("trash: purged %q but failed to drop registry row %d: %v", item.TrashPath, item.ID, err)
			continue
		}
		purged++
	}
	return purged
}

func (s *Service) EmptyTrash() (int, error) {
	items, err := s.Repository.GetAllItems()
	if err != nil {
		return 0, err
	}
	return s.purgeItems(items), nil
}

// PurgeExpired removes every item older than the configured retention.
func (s *Service) PurgeExpired() (int, error) {
	days, err := s.GetRetentionDays()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	items, err := s.Repository.GetExpiredItems(cutoff)
	if err != nil {
		return 0, err
	}
	return s.purgeItems(items), nil
}

func (s *Service) GetRetentionDays() (int, error) {
	days, found, err := s.Repository.GetRetentionDays()
	if err != nil {
		return 0, err
	}
	if !found || days <= 0 {
		return DefaultRetentionDays, nil
	}
	return days, nil
}

func (s *Service) SetRetentionDays(days int) error {
	if days <= 0 {
		return ErrInvalidRetention
	}
	if err := s.Repository.SetRetentionDays(days); err != nil {
		return err
	}
	log.Printf("trash: retention set to %s days", strconv.Itoa(days))
	return nil
}
