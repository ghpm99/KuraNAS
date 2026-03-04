package video

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/video"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(database *database.DbContext) *Repository {
	return &Repository{database}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
}

func (r *Repository) GetVideoFileByID(id int) (VideoFileModel, error) {
	var result VideoFileModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetVideoFileByIDQuery, id, pq.Array(utils.VideoFormats)).Scan(
			&result.ID,
			&result.Name,
			&result.Path,
			&result.ParentPath,
			&result.Format,
			&result.Size,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
	})
	if err != nil {
		return result, fmt.Errorf("falha ao buscar video por id: %w", err)
	}

	return result, nil
}

func (r *Repository) GetVideosByParentPath(parentPath string) ([]VideoFileModel, error) {
	results := []VideoFileModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetVideosByParentPathQuery, parentPath, pq.Array(utils.VideoFormats))
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item VideoFileModel
			if err := rows.Scan(
				&item.ID,
				&item.Name,
				&item.Path,
				&item.ParentPath,
				&item.Format,
				&item.Size,
				&item.CreatedAt,
				&item.UpdatedAt,
			); err != nil {
				return err
			}
			results = append(results, item)
		}
		return nil
	})
	if err != nil {
		return results, fmt.Errorf("falha ao buscar videos por pasta: %w", err)
	}

	return results, nil
}

func (r *Repository) GetPlaylistByContext(contextType string, sourcePath string) (VideoPlaylistModel, error) {
	var playlist VideoPlaylistModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetPlaylistByContextQuery, contextType, sourcePath).Scan(
			&playlist.ID,
			&playlist.Type,
			&playlist.SourcePath,
			&playlist.CreatedAt,
			&playlist.UpdatedAt,
			&playlist.LastPlayedAt,
		)
	})
	if err != nil {
		return playlist, fmt.Errorf("falha ao buscar playlist de contexto: %w", err)
	}

	return playlist, nil
}

func (r *Repository) CreatePlaylist(tx *sql.Tx, contextType string, sourcePath string) (VideoPlaylistModel, error) {
	playlist := VideoPlaylistModel{Type: contextType, SourcePath: sourcePath}
	err := tx.QueryRow(queries.CreatePlaylistQuery, contextType, sourcePath).Scan(
		&playlist.ID,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
		&playlist.LastPlayedAt,
	)
	if err != nil {
		return playlist, fmt.Errorf("falha ao criar playlist de video: %w", err)
	}
	return playlist, nil
}

func (r *Repository) ReplacePlaylistItems(tx *sql.Tx, playlistID int, videoIDs []int) error {
	_, err := tx.Exec(queries.DeletePlaylistItemsQuery, playlistID)
	if err != nil {
		return fmt.Errorf("falha ao limpar itens da playlist: %w", err)
	}

	if len(videoIDs) == 0 {
		return nil
	}

	_, err = tx.Exec(queries.InsertPlaylistItemsQuery, playlistID, pq.Array(videoIDs))
	if err != nil {
		return fmt.Errorf("falha ao substituir itens da playlist: %w", err)
	}
	return nil
}

func (r *Repository) GetPlaylistItems(playlistID int) ([]VideoPlaylistItemModel, error) {
	items := []VideoPlaylistItemModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetPlaylistItemsQuery, playlistID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item VideoPlaylistItemModel
			if err := rows.Scan(
				&item.ID,
				&item.PlaylistID,
				&item.VideoID,
				&item.OrderIndex,
				&item.Video.Name,
				&item.Video.Path,
				&item.Video.ParentPath,
				&item.Video.Format,
				&item.Video.Size,
			); err != nil {
				return err
			}
			item.Video.ID = item.VideoID
			items = append(items, item)
		}
		return nil
	})
	if err != nil {
		return items, fmt.Errorf("falha ao buscar itens da playlist: %w", err)
	}

	return items, nil
}

func (r *Repository) GetPlaybackState(clientID string) (VideoPlaybackStateModel, error) {
	var state VideoPlaybackStateModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetPlaybackStateQuery, clientID).Scan(
			&state.ID,
			&state.ClientID,
			&state.PlaylistID,
			&state.VideoID,
			&state.CurrentTime,
			&state.Duration,
			&state.IsPaused,
			&state.Completed,
			&state.LastUpdate,
		)
	})
	if err != nil {
		return state, fmt.Errorf("falha ao buscar estado de reproducao: %w", err)
	}

	return state, nil
}

func (r *Repository) UpsertPlaybackState(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
	err := tx.QueryRow(
		queries.UpsertPlaybackStateQuery,
		state.ClientID,
		state.PlaylistID,
		state.VideoID,
		state.CurrentTime,
		state.Duration,
		state.IsPaused,
		state.Completed,
	).Scan(&state.ID, &state.LastUpdate)
	if err != nil {
		return state, fmt.Errorf("falha ao salvar estado de reproducao: %w", err)
	}
	return state, nil
}

func (r *Repository) TouchPlaylist(tx *sql.Tx, playlistID int) error {
	_, err := tx.Exec(queries.TouchPlaylistQuery, playlistID)
	if err != nil {
		return fmt.Errorf("falha ao atualizar last_played da playlist: %w", err)
	}
	return nil
}

func (r *Repository) GetCatalogVideos(limit int) ([]VideoFileModel, error) {
	results := []VideoFileModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetCatalogVideosQuery, pq.Array(utils.VideoFormats), limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item VideoFileModel
			if err := rows.Scan(
				&item.ID,
				&item.Name,
				&item.Path,
				&item.ParentPath,
				&item.Format,
				&item.Size,
				&item.CreatedAt,
				&item.UpdatedAt,
			); err != nil {
				return err
			}
			results = append(results, item)
		}

		return nil
	})
	if err != nil {
		return results, fmt.Errorf("falha ao buscar videos para catalogo: %w", err)
	}

	return results, nil
}

func (r *Repository) GetRecentVideos(limit int) ([]VideoFileModel, error) {
	results := []VideoFileModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetRecentVideosQuery, pq.Array(utils.VideoFormats), limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item VideoFileModel
			if err := rows.Scan(
				&item.ID,
				&item.Name,
				&item.Path,
				&item.ParentPath,
				&item.Format,
				&item.Size,
				&item.CreatedAt,
				&item.UpdatedAt,
			); err != nil {
				return err
			}
			results = append(results, item)
		}

		return nil
	})
	if err != nil {
		return results, fmt.Errorf("falha ao buscar videos recentes: %w", err)
	}

	return results, nil
}
