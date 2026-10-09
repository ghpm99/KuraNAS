package dav

import (
	"database/sql"
	"fmt"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/dav"
)

type ColdFile struct {
	Name         string
	LogicalPath  string
	PhysicalPath string
	Size         int64
	ModTime      time.Time
}

// ColdFileCatalog lists the files of a directory whose bytes were demoted to
// the cold tier and are therefore missing from the hot disk.
type ColdFileCatalog interface {
	ListColdFilesByParentPath(parentPath string) ([]ColdFile, error)
}

type ColdFileRepository struct {
	dbContext *database.DbContext
}

func NewColdFileRepository(dbContext *database.DbContext) *ColdFileRepository {
	return &ColdFileRepository{dbContext: dbContext}
}

func (repository *ColdFileRepository) ListColdFilesByParentPath(parentPath string) ([]ColdFile, error) {
	var coldFiles []ColdFile

	err := repository.dbContext.QueryTx(func(tx *sql.Tx) error {
		rows, queryErr := tx.Query(queries.ListColdFilesByParentPathQuery, parentPath)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		for rows.Next() {
			var coldFile ColdFile
			if scanErr := rows.Scan(&coldFile.Name, &coldFile.LogicalPath, &coldFile.PhysicalPath, &coldFile.Size, &coldFile.ModTime); scanErr != nil {
				return scanErr
			}
			coldFiles = append(coldFiles, coldFile)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListColdFilesByParentPath: %w", err)
	}
	return coldFiles, nil
}
