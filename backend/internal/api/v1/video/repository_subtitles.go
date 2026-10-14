package video

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/video"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type SubtitleRepositoryInterface interface {
	ListSubtitleFilesByNamePrefix(parentPath string, videoName string) ([]SubtitleFileModel, error)
}

type SubtitleRepository struct {
	Db *database.DbContext
}

func NewSubtitleRepository(db *database.DbContext) *SubtitleRepository {
	return &SubtitleRepository{Db: db}
}

func (r *SubtitleRepository) ListSubtitleFilesByNamePrefix(parentPath string, videoName string) ([]SubtitleFileModel, error) {
	subtitleFiles := []SubtitleFileModel{}
	namePattern := utils.BuildPrefixLikePattern(SubtitleNamePrefixFor(videoName))
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ListSubtitleFilesByNamePrefixQuery, parentPath, pq.Array(subtitleExtensions), namePattern)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var subtitleFile SubtitleFileModel
			if err := rows.Scan(&subtitleFile.ID, &subtitleFile.Name); err != nil {
				return err
			}
			subtitleFiles = append(subtitleFiles, subtitleFile)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListSubtitleFilesByNamePrefix: %w", err)
	}
	return subtitleFiles, nil
}
