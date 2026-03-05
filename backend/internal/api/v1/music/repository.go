package music

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/music"
	"nas-go/api/pkg/utils"
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

func (r *Repository) GetPlaylists(page int, pageSize int) (utils.PaginationResponse[PlaylistModel], error) {
	paginationResponse := utils.PaginationResponse[PlaylistModel]{
		Items: []PlaylistModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetPlaylistsQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var playlist PlaylistModel
			if err := rows.Scan(
				&playlist.ID, &playlist.Name, &playlist.Description,
				&playlist.IsSystem, &playlist.CreatedAt, &playlist.UpdatedAt,
				&playlist.TrackCount,
			); err != nil {
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, playlist)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("falha na consulta de playlists: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetPlaylistByID(id int) (PlaylistModel, error) {
	var playlist PlaylistModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetPlaylistByIDQuery, id)
		return row.Scan(
			&playlist.ID, &playlist.Name, &playlist.Description,
			&playlist.IsSystem, &playlist.CreatedAt, &playlist.UpdatedAt,
			&playlist.TrackCount,
		)
	})

	if err != nil {
		return playlist, fmt.Errorf("falha ao obter playlist: %w", err)
	}

	return playlist, nil
}

func (r *Repository) CreatePlaylist(tx *sql.Tx, name string, description string, isSystem bool) (PlaylistModel, error) {
	var playlist PlaylistModel
	playlist.Name = name
	playlist.Description = description
	playlist.IsSystem = isSystem

	err := tx.QueryRow(queries.CreatePlaylistQuery, name, description, isSystem).Scan(
		&playlist.ID, &playlist.CreatedAt, &playlist.UpdatedAt,
	)

	if err != nil {
		return playlist, fmt.Errorf("falha ao criar playlist: %w", err)
	}

	return playlist, nil
}

func (r *Repository) UpdatePlaylist(tx *sql.Tx, id int, name string, description string) (PlaylistModel, error) {
	var playlist PlaylistModel
	playlist.ID = id
	playlist.Name = name
	playlist.Description = description

	err := tx.QueryRow(queries.UpdatePlaylistQuery, name, description, id).Scan(
		&playlist.UpdatedAt,
	)

	if err != nil {
		return playlist, fmt.Errorf("falha ao atualizar playlist: %w", err)
	}

	return playlist, nil
}

func (r *Repository) DeletePlaylist(tx *sql.Tx, id int) error {
	result, err := tx.Exec(queries.DeletePlaylistQuery, id)
	if err != nil {
		return fmt.Errorf("falha ao deletar playlist: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao verificar exclusão: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("playlist não encontrada ou é do sistema")
	}

	return nil
}

func (r *Repository) GetPlaylistTracks(playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackModel], error) {
	paginationResponse := utils.PaginationResponse[PlaylistTrackModel]{
		Items: []PlaylistTrackModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		playlistID,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetPlaylistTracksQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var track PlaylistTrackModel
			if err := rows.Scan(
				&track.ID, &track.PlaylistID, &track.FileID,
				&track.Position, &track.AddedAt,
				&track.FileID, &track.FileName, &track.FilePath,
				&track.FileParentPath, &track.FileFormat, &track.FileSize,
				&track.FileUpdatedAt, &track.FileCreatedAt,
				&track.LastInteraction, &track.LastBackup,
				&track.FileType, &track.FileCheckSum,
				&track.FileDeletedAt, &track.FileStarred,
				&track.MetadataID, &track.MetadataFileId, &track.MetadataPath,
				&track.MetadataMime, &track.MetadataLength, &track.MetadataBitrate,
				&track.MetadataSampleRate, &track.MetadataChannels,
				&track.MetadataBitrateMode, &track.MetadataEncoderInfo, &track.MetadataBitDepth,
				&track.MetadataTitle, &track.MetadataArtist, &track.MetadataAlbum,
				&track.MetadataAlbumArtist, &track.MetadataTrackNumber, &track.MetadataGenre,
				&track.MetadataComposer, &track.MetadataYear, &track.MetadataRecordingDate,
				&track.MetadataEncoder, &track.MetadataPublisher, &track.MetadataOriginalReleaseDate,
				&track.MetadataOriginalArtist, &track.MetadataLyricist, &track.MetadataLyrics,
				&track.MetadataCreatedAt,
			); err != nil {
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, track)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("falha na consulta de tracks: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) AddPlaylistTrack(tx *sql.Tx, playlistID int, fileID int) (PlaylistTrackModel, error) {
	var track PlaylistTrackModel
	track.PlaylistID = playlistID
	track.FileID = fileID

	err := tx.QueryRow(queries.AddPlaylistTrackQuery, playlistID, fileID).Scan(
		&track.ID, &track.Position, &track.AddedAt,
	)

	if err != nil {
		return track, fmt.Errorf("falha ao adicionar track: %w", err)
	}

	return track, nil
}

func (r *Repository) RemovePlaylistTrack(tx *sql.Tx, playlistID int, fileID int) error {
	result, err := tx.Exec(queries.RemovePlaylistTrackQuery, playlistID, fileID)
	if err != nil {
		return fmt.Errorf("falha ao remover track: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("falha ao verificar remoção: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("track não encontrada na playlist")
	}

	return nil
}

func (r *Repository) ReorderPlaylistTrack(tx *sql.Tx, playlistID int, fileID int, position int) error {
	_, err := tx.Exec(queries.ReorderPlaylistTrackQuery, position, playlistID, fileID)
	if err != nil {
		return fmt.Errorf("falha ao reordenar track: %w", err)
	}
	return nil
}

func (r *Repository) GetNowPlaying() (PlaylistModel, error) {
	var playlist PlaylistModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetNowPlayingQuery)
		return row.Scan(
			&playlist.ID, &playlist.Name, &playlist.Description,
			&playlist.IsSystem, &playlist.CreatedAt, &playlist.UpdatedAt,
			&playlist.TrackCount,
		)
	})

	if err != nil {
		return playlist, err
	}

	return playlist, nil
}

func (r *Repository) GetPlayerState(clientID string) (PlayerStateModel, error) {
	var state PlayerStateModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetPlayerStateQuery, clientID)
		return row.Scan(
			&state.ID, &state.ClientID, &state.PlaylistID,
			&state.CurrentFileID, &state.CurrentPosition,
			&state.Volume, &state.Shuffle, &state.RepeatMode,
			&state.UpdatedAt,
		)
	})

	if err != nil {
		return state, fmt.Errorf("falha ao obter player state: %w", err)
	}

	return state, nil
}

func (r *Repository) UpsertPlayerState(tx *sql.Tx, state PlayerStateModel) (PlayerStateModel, error) {
	err := tx.QueryRow(
		queries.UpsertPlayerStateQuery,
		state.ClientID, state.PlaylistID, state.CurrentFileID,
		state.CurrentPosition, state.Volume, state.Shuffle, state.RepeatMode,
	).Scan(&state.ID, &state.UpdatedAt)

	if err != nil {
		return state, fmt.Errorf("falha ao salvar player state: %w", err)
	}

	return state, nil
}
