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
			&playlist.Name,
			&playlist.IsHidden,
			&playlist.IsAuto,
			&playlist.GroupMode,
			&playlist.Classification,
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
	playlist := VideoPlaylistModel{Type: contextType, SourcePath: sourcePath, Name: sourcePath}
	err := tx.QueryRow(queries.CreatePlaylistQuery, contextType, sourcePath, sourcePath).Scan(
		&playlist.ID,
		&playlist.Type,
		&playlist.SourcePath,
		&playlist.Name,
		&playlist.IsHidden,
		&playlist.IsAuto,
		&playlist.GroupMode,
		&playlist.Classification,
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
				&item.SourceKind,
				&item.Video.Name,
				&item.Video.Path,
				&item.Video.ParentPath,
				&item.Video.Format,
				&item.Video.Size,
				&item.Video.CreatedAt,
				&item.Video.UpdatedAt,
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

func (r *Repository) GetAllVideosForGrouping() ([]VideoFileModel, error) {
	results := []VideoFileModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetAllVideosForGroupingQuery, pq.Array(utils.VideoFormats))
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
		return results, fmt.Errorf("falha ao buscar videos para agrupamento inteligente: %w", err)
	}

	return results, nil
}

func (r *Repository) UpsertAutoPlaylist(tx *sql.Tx, contextType, sourcePath, name, groupMode, classification string) (VideoPlaylistModel, error) {
	playlist := VideoPlaylistModel{}
	err := tx.QueryRow(
		queries.UpsertAutoPlaylistQuery,
		contextType,
		sourcePath,
		name,
		groupMode,
		classification,
	).Scan(
		&playlist.ID,
		&playlist.Type,
		&playlist.SourcePath,
		&playlist.Name,
		&playlist.IsHidden,
		&playlist.IsAuto,
		&playlist.GroupMode,
		&playlist.Classification,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
		&playlist.LastPlayedAt,
	)
	if err != nil {
		return playlist, fmt.Errorf("falha ao criar/atualizar playlist inteligente: %w", err)
	}

	return playlist, nil
}

func (r *Repository) DeleteAutoPlaylistItems(tx *sql.Tx, playlistID int) error {
	_, err := tx.Exec(queries.DeleteAutoPlaylistItemsQuery, playlistID)
	if err != nil {
		return fmt.Errorf("falha ao limpar itens auto da playlist: %w", err)
	}
	return nil
}

func (r *Repository) InsertPlaylistItemsWithSource(tx *sql.Tx, playlistID int, videoIDs []int, sourceKind string) error {
	if len(videoIDs) == 0 {
		return nil
	}
	_, err := tx.Exec(queries.InsertPlaylistItemsWithSourceQuery, playlistID, pq.Array(videoIDs), sourceKind)
	if err != nil {
		return fmt.Errorf("falha ao inserir itens na playlist com source_kind=%s: %w", sourceKind, err)
	}
	return nil
}

func (r *Repository) GetPlaylistExclusions(playlistID int) (map[int]bool, error) {
	exclusions := map[int]bool{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetPlaylistExclusionsQuery, playlistID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var videoID int
			if err := rows.Scan(&videoID); err != nil {
				return err
			}
			exclusions[videoID] = true
		}
		return nil
	})
	if err != nil {
		return exclusions, fmt.Errorf("falha ao buscar exclusoes da playlist: %w", err)
	}

	return exclusions, nil
}

func (r *Repository) GetVideoPlaylists(includeHidden bool) ([]VideoPlaylistModel, error) {
	playlists := []VideoPlaylistModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetVideoPlaylistsQuery, includeHidden)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item VideoPlaylistModel
			if err := rows.Scan(
				&item.ID,
				&item.Type,
				&item.SourcePath,
				&item.Name,
				&item.IsHidden,
				&item.IsAuto,
				&item.GroupMode,
				&item.Classification,
				&item.CreatedAt,
				&item.UpdatedAt,
				&item.LastPlayedAt,
				&item.ItemCount,
				&item.CoverVideoID,
			); err != nil {
				return err
			}
			playlists = append(playlists, item)
		}
		return nil
	})
	if err != nil {
		return playlists, fmt.Errorf("falha ao listar playlists de video: %w", err)
	}

	return playlists, nil
}

func (r *Repository) GetVideoPlaylistByID(id int) (VideoPlaylistModel, error) {
	var playlist VideoPlaylistModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetVideoPlaylistByIDQuery, id).Scan(
			&playlist.ID,
			&playlist.Type,
			&playlist.SourcePath,
			&playlist.Name,
			&playlist.IsHidden,
			&playlist.IsAuto,
			&playlist.GroupMode,
			&playlist.Classification,
			&playlist.CreatedAt,
			&playlist.UpdatedAt,
			&playlist.LastPlayedAt,
		)
	})
	if err != nil {
		return playlist, fmt.Errorf("falha ao obter playlist de video por id: %w", err)
	}

	return playlist, nil
}

func (r *Repository) GetVideoPlaylistItemsDetailed(playlistID int) ([]VideoPlaylistItemModel, error) {
	items := []VideoPlaylistItemModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetVideoPlaylistItemsDetailedQuery, playlistID)
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
				&item.SourceKind,
				&item.Video.Name,
				&item.Video.Path,
				&item.Video.ParentPath,
				&item.Video.Format,
				&item.Video.Size,
				&item.Video.CreatedAt,
				&item.Video.UpdatedAt,
			); err != nil {
				return err
			}
			item.Video.ID = item.VideoID
			items = append(items, item)
		}
		return nil
	})
	if err != nil {
		return items, fmt.Errorf("falha ao buscar itens detalhados da playlist de video: %w", err)
	}

	return items, nil
}

func (r *Repository) SetPlaylistHidden(tx *sql.Tx, playlistID int, hidden bool) error {
	_, err := tx.Exec(queries.SetPlaylistHiddenQuery, playlistID, hidden)
	if err != nil {
		return fmt.Errorf("falha ao atualizar ocultacao da playlist: %w", err)
	}
	return nil
}

func (r *Repository) AddPlaylistVideoManual(tx *sql.Tx, playlistID int, videoID int) error {
	_, err := tx.Exec(queries.AddPlaylistVideoManualQuery, playlistID, videoID)
	if err != nil {
		return fmt.Errorf("falha ao adicionar video manualmente na playlist: %w", err)
	}
	return nil
}

func (r *Repository) RemovePlaylistVideo(tx *sql.Tx, playlistID int, videoID int) error {
	_, err := tx.Exec(queries.RemovePlaylistVideoQuery, playlistID, videoID)
	if err != nil {
		return fmt.Errorf("falha ao remover video da playlist: %w", err)
	}
	return nil
}

func (r *Repository) UpsertPlaylistExclusion(tx *sql.Tx, playlistID int, videoID int) error {
	_, err := tx.Exec(queries.UpsertPlaylistExclusionQuery, playlistID, videoID)
	if err != nil {
		return fmt.Errorf("falha ao marcar exclusao manual de video da playlist: %w", err)
	}
	return nil
}

func (r *Repository) DeletePlaylistExclusion(tx *sql.Tx, playlistID int, videoID int) error {
	_, err := tx.Exec(queries.DeletePlaylistExclusionQuery, playlistID, videoID)
	if err != nil {
		return fmt.Errorf("falha ao remover exclusao manual da playlist: %w", err)
	}
	return nil
}

func (r *Repository) GetUnassignedVideos(limit int) ([]VideoFileModel, error) {
	results := []VideoFileModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetUnassignedVideosQuery, pq.Array(utils.VideoFormats), limit)
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
		return results, fmt.Errorf("falha ao listar videos sem playlist: %w", err)
	}

	return results, nil
}

func (r *Repository) CheckVideoInPlaylist(playlistID int, videoID int) (bool, error) {
	var count int
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.CheckVideoInPlaylistQuery, playlistID, videoID).Scan(&count)
	})
	if err != nil {
		return false, fmt.Errorf("falha ao validar video na playlist: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) UpdatePlaylistName(tx *sql.Tx, playlistID int, name string) error {
	_, err := tx.Exec(queries.UpdatePlaylistNameQuery, playlistID, name)
	if err != nil {
		return fmt.Errorf("falha ao atualizar nome da playlist: %w", err)
	}
	return nil
}

func (r *Repository) ReorderPlaylistItem(tx *sql.Tx, playlistID int, videoID int, orderIndex int) error {
	_, err := tx.Exec(queries.ReorderPlaylistItemQuery, playlistID, videoID, orderIndex)
	if err != nil {
		return fmt.Errorf("falha ao reordenar item da playlist: %w", err)
	}
	return nil
}
