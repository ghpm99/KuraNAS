package music

import (
	"database/sql"
	"fmt"
	"time"

	queries "nas-go/api/pkg/database/queries/music"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

func (r *Repository) InsertPlayEvent(tx *sql.Tx, clientID string, fileID int, playedSeconds int) error {
	insertResult, err := tx.Exec(queries.InsertMusicPlayEventQuery, fileID, pq.Array(utils.AudioFormats), clientID, playedSeconds)
	if err != nil {
		return fmt.Errorf("falha ao registrar a reproducao: %w", err)
	}
	insertedRows, err := insertResult.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao registrar a reproducao: %w", err)
	}
	if insertedRows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) GetMostPlayedTracks(earliestPlayedAt *time.Time, page int, pageSize int) (utils.PaginationResponse[PlayedTrackModel], error) {
	mostPlayed, err := queryCatalogPage(r.DbContext, queries.GetMostPlayedTracksQuery,
		[]any{pq.Array(utils.AudioFormats), earliestPlayedAt}, page, pageSize, scanPlayedTrack)
	if err != nil {
		return mostPlayed, fmt.Errorf("falha ao listar as faixas mais tocadas: %w", err)
	}
	return mostPlayed, nil
}

func (r *Repository) GetRecentlyPlayedTracks(page int, pageSize int) (utils.PaginationResponse[PlayedTrackModel], error) {
	recentlyPlayed, err := queryCatalogPage(r.DbContext, queries.GetRecentlyPlayedTracksQuery,
		[]any{pq.Array(utils.AudioFormats)}, page, pageSize, scanPlayedTrack)
	if err != nil {
		return recentlyPlayed, fmt.Errorf("falha ao listar as faixas tocadas recentemente: %w", err)
	}
	return recentlyPlayed, nil
}

func scanPlayedTrack(rows *sql.Rows) (PlayedTrackModel, error) {
	var playedTrack PlayedTrackModel
	err := rows.Scan(&playedTrack.FileID, &playedTrack.PlayCount, &playedTrack.LastPlayedAt)
	return playedTrack, err
}
