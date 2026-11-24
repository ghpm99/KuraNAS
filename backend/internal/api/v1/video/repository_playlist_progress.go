package video

import (
	"database/sql"
	"errors"
	"fmt"

	queries "nas-go/api/pkg/database/queries/video"
)

func (r *Repository) GetPlaylistSeasonSummaries(clientID string, playlistID int) ([]PlaylistSeasonSummaryModel, error) {
	summaries := []PlaylistSeasonSummaryModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetPlaylistSeasonSummariesQuery, playlistID, clientID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var summary PlaylistSeasonSummaryModel
			if err := rows.Scan(&summary.Season, &summary.EpisodeCount, &summary.WatchedCount); err != nil {
				return err
			}
			summaries = append(summaries, summary)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar temporadas da playlist de video: %w", err)
	}
	return summaries, nil
}

func (r *Repository) GetPlaylistResume(clientID string, playlistID int) (*PlaylistResumeModel, error) {
	var resume PlaylistResumeModel
	var season, episode sql.NullInt64

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetPlaylistResumeEpisodeQuery, playlistID, clientID).Scan(
			&resume.VideoID, &season, &episode, &resume.PositionSeconds, &resume.Mode,
		)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar ponto de retomada da playlist de video: %w", err)
	}

	resume.Season = nullableInt(season)
	resume.Episode = nullableInt(episode)
	return &resume, nil
}

func (r *Repository) SetPlaylistWatched(tx *sql.Tx, clientID string, playlistID int, season *int, isWatched bool) (int, error) {
	execResult, err := tx.Exec(queries.SetPlaylistWatchedQuery, clientID, playlistID, nullableSeasonArgument(season), isWatched)
	if err != nil {
		return 0, fmt.Errorf("falha ao marcar playlist de video como assistida: %w", err)
	}
	updatedCount, err := execResult.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("falha ao contar episodios atualizados: %w", err)
	}
	return int(updatedCount), nil
}
