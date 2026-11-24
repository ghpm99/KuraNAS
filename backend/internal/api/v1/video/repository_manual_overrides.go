package video

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/video"
)

func (r *Repository) UpsertVideoClassificationOverride(tx *sql.Tx, fileID int, classification sql.NullString) error {
	if _, err := tx.Exec(queries.UpsertVideoClassificationOverrideQuery, fileID, classification); err != nil {
		return fmt.Errorf("falha ao gravar correcao manual de classificacao: %w", err)
	}
	return nil
}

func (r *Repository) UpsertVideoEpisodeOverride(tx *sql.Tx, fileID int, season sql.NullInt64, episode sql.NullInt64) error {
	if _, err := tx.Exec(queries.UpsertVideoEpisodeOverrideQuery, fileID, season, episode); err != nil {
		return fmt.Errorf("falha ao gravar correcao manual de episodio: %w", err)
	}
	return nil
}

func (r *Repository) DeleteEmptyVideoManualOverride(tx *sql.Tx, fileID int) error {
	if _, err := tx.Exec(queries.DeleteEmptyVideoManualOverrideQuery, fileID); err != nil {
		return fmt.Errorf("falha ao remover correcao manual vazia: %w", err)
	}
	return nil
}

func (r *Repository) GetVideoClassificationState(fileID int) (VideoClassificationStateModel, error) {
	var state VideoClassificationStateModel
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetVideoClassificationStateQuery, fileID).
			Scan(&state.OverrideClassification, &state.AutomaticClassification)
	})
	if err != nil {
		return state, fmt.Errorf("falha ao buscar classificacao do video: %w", err)
	}
	return state, nil
}

func (r *Repository) GetVideoEpisodeState(fileID int) (VideoEpisodeStateModel, error) {
	var state VideoEpisodeStateModel
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetVideoEpisodeStateQuery, fileID).Scan(
			&state.Name,
			&state.OverrideSeason,
			&state.OverrideEpisode,
			&state.AutomaticSeason,
			&state.AutomaticEpisode,
		)
	})
	if err != nil {
		return state, fmt.Errorf("falha ao buscar episodio do video: %w", err)
	}
	return state, nil
}

func (r *Repository) GetAutoSeriesPlaylistIDsByVideo(videoID int) ([]int, error) {
	playlistIDs := []int{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetAutoSeriesPlaylistIDsByVideoQuery, videoID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var playlistID int
			if err := rows.Scan(&playlistID); err != nil {
				return err
			}
			playlistIDs = append(playlistIDs, playlistID)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar playlists de serie do video: %w", err)
	}
	return playlistIDs, nil
}
