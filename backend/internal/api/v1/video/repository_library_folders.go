package video

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/video"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

func (r *Repository) ListLibraryFolders(query LibraryFolderQuery) ([]LibraryFolderModel, error) {
	prefixes := make([]string, 0, len(query.Scopes))
	wholeRootFlags := make([]bool, 0, len(query.Scopes))
	labels := make([]string, 0, len(query.Scopes))
	for _, scope := range query.Scopes {
		prefixes = append(prefixes, scope.Prefix)
		wholeRootFlags = append(wholeRootFlags, scope.IsWholeRoot)
		labels = append(labels, scope.Label)
	}

	folders := []LibraryFolderModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.GetLibraryFoldersQuery,
			pq.Array(utils.VideoFormats),
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
			if err := rows.Scan(&folder.Path, &folder.Name, &folder.VideoCount, &folder.CoverFileID); err != nil {
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

func (r *Repository) ListLibraryFolderVideos(folderPath string, limit int, offset int) ([]VideoFileModel, error) {
	videos := []VideoFileModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetLibraryFolderVideosQuery, pq.Array(utils.VideoFormats), folderPath, limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var video VideoFileModel
			if err := rows.Scan(
				&video.ID,
				&video.Name,
				&video.Path,
				&video.ParentPath,
				&video.Format,
				&video.Size,
				&video.CreatedAt,
				&video.UpdatedAt,
			); err != nil {
				return err
			}
			videos = append(videos, video)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListLibraryFolderVideos: %w", err)
	}
	return videos, nil
}
