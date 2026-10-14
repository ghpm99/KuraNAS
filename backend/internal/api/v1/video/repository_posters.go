package video

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/video"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

func (r *Repository) ListPosterCandidatesInFolder(folderPath string) ([]PosterCandidateModel, error) {
	candidates := []PosterCandidateModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ListPosterCandidateImagesInFolderQuery, folderPath, pq.Array(posterImageExtensions))
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var candidate PosterCandidateModel
			if err := rows.Scan(&candidate.ID, &candidate.Name); err != nil {
				return err
			}
			candidates = append(candidates, candidate)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListPosterCandidatesInFolder: %w", err)
	}
	return candidates, nil
}

func (r *Repository) CountVideosInFolder(folderPath string) (int, error) {
	var videoCount int
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.CountVideosInFolderQuery, folderPath, pq.Array(utils.VideoFormats)).Scan(&videoCount)
	})
	if err != nil {
		return 0, fmt.Errorf("CountVideosInFolder: %w", err)
	}
	return videoCount, nil
}

func (r *Repository) GetPlaylistDominantFolder(playlistID int) (string, error) {
	var folderPath string
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetPlaylistDominantFolderQuery, playlistID).Scan(&folderPath)
	})
	if err != nil {
		return "", fmt.Errorf("GetPlaylistDominantFolder: %w", err)
	}
	return folderPath, nil
}
