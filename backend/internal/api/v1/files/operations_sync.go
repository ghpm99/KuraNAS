package files

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	queries "nas-go/api/pkg/database/queries/files"
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
