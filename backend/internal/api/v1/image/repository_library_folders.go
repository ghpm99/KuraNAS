package image

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/image"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

func (r *LibraryRepository) ListLibraryFolders(query LibraryFolderQuery) ([]LibraryFolderModel, error) {
	prefixes := make([]string, 0, len(query.Scopes))
	wholeRootFlags := make([]bool, 0, len(query.Scopes))
	labels := make([]string, 0, len(query.Scopes))
	for _, scope := range query.Scopes {
		prefixes = append(prefixes, scope.Prefix)
		wholeRootFlags = append(wholeRootFlags, scope.IsWholeRoot)
		labels = append(labels, scope.Label)
	}

	folders := []LibraryFolderModel{}
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.LibraryFoldersQuery,
			pq.Array(utils.ImageFormats),
			pq.Array(prefixes),
			pq.Array(wholeRootFlags),
			pq.Array(labels),
			query.Separator,
			query.Limit,
			query.Offset,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var folder LibraryFolderModel
			if err := rows.Scan(&folder.Path, &folder.Name, &folder.ImageCount, &folder.CoverFileID); err != nil {
				return err
			}
			folders = append(folders, folder)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListLibraryFolders: %w", err)
	}
	return folders, nil
}
