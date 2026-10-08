package files

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/files"
)

func (r *Repository) GetActiveFileByPathOrPhysicalPath(path string) (FileModel, bool, error) {
	var file FileModel
	found := false

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetActiveFileByPathOrPhysicalPathQuery, path)
		if err != nil {
			return err
		}
		defer rows.Close()

		items, scanErr := scanFileRows(rows)
		if scanErr != nil {
			return scanErr
		}
		if len(items) > 0 {
			file = items[0]
			found = true
		}
		return nil
	})
	if err != nil {
		return FileModel{}, false, fmt.Errorf("GetActiveFileByPathOrPhysicalPath: %w", err)
	}
	return file, found, nil
}
