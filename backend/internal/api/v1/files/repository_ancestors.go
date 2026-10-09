package files

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/files"

	"github.com/lib/pq"
)

func (r *Repository) GetActiveFilesByPaths(paths []string) ([]FileModel, error) {
	var matchedFiles []FileModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetActiveFilesByPathsQuery, pq.Array(paths))
		if err != nil {
			return err
		}
		defer rows.Close()

		scannedFiles, scanErr := scanFileRows(rows)
		if scanErr != nil {
			return scanErr
		}
		matchedFiles = scannedFiles
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("GetActiveFilesByPaths: %w", err)
	}
	return matchedFiles, nil
}
