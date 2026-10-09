package files

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/files"
)

func (r *Repository) GetFolderStats(descendantPathPrefix string) (FolderStatsDto, error) {
	var stats FolderStatsDto

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetFolderStatsQuery, descendantPathPrefix, File, Directory)
		return row.Scan(&stats.FileCount, &stats.FolderCount, &stats.TotalSizeBytes)
	})
	if err != nil {
		return FolderStatsDto{}, fmt.Errorf("GetFolderStats: %w", err)
	}
	return stats, nil
}
