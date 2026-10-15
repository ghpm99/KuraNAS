package files

import (
	"database/sql"
	"fmt"
	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"
)

func (r *Repository) GetStarredFiles(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.GetStarredFilesQuery, page, pageSize)
}

func (r *Repository) GetRecentlyAccessedFiles(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	return r.queryFilesPage(queries.GetRecentlyAccessedFilesQuery, page, pageSize)
}

// ToggleFileStarred flips the starred flag of an active file in a single
// statement and returns the new value. It returns sql.ErrNoRows when no
// active file has the given id.
func (r *Repository) ToggleFileStarred(id int) (bool, error) {
	var isStarred bool
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.ToggleFileStarredQuery, id).Scan(&isStarred)
	})
	if err != nil {
		return false, fmt.Errorf("ToggleFileStarred: %w", err)
	}
	return isStarred, nil
}

// SetFileStarred sets the starred flag of an active file to isStarred and
// returns the stored value. Repeating the call yields the same state. It
// returns sql.ErrNoRows when no active file has the given id.
func (r *Repository) SetFileStarred(id int, isStarred bool) (bool, error) {
	var storedValue bool
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.SetFileStarredQuery, id, isStarred).Scan(&storedValue)
	})
	if err != nil {
		return false, fmt.Errorf("SetFileStarred: %w", err)
	}
	return storedValue, nil
}
