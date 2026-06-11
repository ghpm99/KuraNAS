package files

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"
)

// Disk → database synchronization for file operations. After a disk operation
// succeeds, these update the affected home_file rows in the same transaction so
// an immediate read reflects the new state; the async rescan stays as a safety
// net, not the primary sync mechanism.

// UpdateDescendantPaths rewrites the path prefix of every descendant of a
// moved/renamed directory (the directory's own row is updated separately).
// Returns the number of affected rows.
func (r *Repository) UpdateDescendantPaths(transaction *sql.Tx, oldPath string, newPath string) (int64, error) {
	result, err := transaction.Exec(
		queries.UpdateDescendantPathsQuery,
		oldPath,
		newPath,
		oldPath+string(filepath.Separator),
	)
	if err != nil {
		return 0, fmt.Errorf("UpdateDescendantPaths: %w", err)
	}
	return result.RowsAffected()
}

// MarkDeletedSubtree soft-deletes the row at path and every descendant row.
// Returns the number of affected rows.
func (r *Repository) MarkDeletedSubtree(transaction *sql.Tx, path string, deletedAt time.Time) (int64, error) {
	result, err := transaction.Exec(
		queries.MarkDeletedSubtreeQuery,
		path,
		deletedAt,
		path+string(filepath.Separator),
	)
	if err != nil {
		return 0, fmt.Errorf("MarkDeletedSubtree: %w", err)
	}
	return result.RowsAffected()
}

// logSyncFailure records a failed disk→database sync. The disk operation
// already succeeded and stays authoritative; the pending ScanDirTask
// reconciles the rows later, so the operation still reports success.
func (s *Service) logSyncFailure(operation string, path string, err error) {
	log.Printf("%s: database sync failed for %q (rescan will reconcile): %v", operation, path, err)
}

// syncPathRow inserts the row for a path just materialized on disk, reviving
// the soft-deleted row when the same path is recreated instead of duplicating it.
func (s *Service) syncPathRow(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	fileDto := FileDto{Path: path, ParentPath: filepath.Dir(path)}
	if err := fileDto.ParseFileInfoToFileDto(info); err != nil {
		return err
	}

	existing, err := s.GetFileByNameAndPath(fileDto.Name, fileDto.Path)
	if err == nil {
		existing.Size = fileDto.Size
		existing.UpdatedAt = fileDto.UpdatedAt
		existing.DeletedAt = utils.Optional[time.Time]{}
		_, err = s.UpdateFile(existing)
		return err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = s.CreateFile(fileDto)
	return err
}
