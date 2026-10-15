package files

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/files"

	"github.com/lib/pq"
)

func (r *Repository) CountStarredFilesByKind() (StarredCountsDto, error) {
	var counts StarredCountsDto

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.CountStarredFilesByKindQuery, pq.Array(mediaFormats()))
		return row.Scan(&counts.Total, &counts.Folders, &counts.Files, &counts.Media)
	})
	if err != nil {
		return StarredCountsDto{}, fmt.Errorf("CountStarredFilesByKind: %w", err)
	}
	return counts, nil
}
