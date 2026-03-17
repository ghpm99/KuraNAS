package music

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
)

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{
		Repository: repository,
	}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
}

func (s *Service) GetPlaylists(page int, pageSize int) (utils.PaginationResponse[PlaylistDto], error) {
	playlistsModel, err := s.Repository.GetPlaylists(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[PlaylistDto]{}, err
	}
	return ParsePlaylistPaginationToDto(&playlistsModel), nil
}

func (s *Service) GetPlaylistByID(id int) (PlaylistDto, error) {
	if id < 0 {
		playlists, err := s.GetAutomaticPlaylists("")
		if err != nil {
			return PlaylistDto{}, err
		}
		for _, playlist := range playlists {
			if playlist.ID == id {
				return playlist, nil
			}
		}
		return PlaylistDto{}, sql.ErrNoRows
	}

	playlist, err := s.Repository.GetPlaylistByID(id)
	if err != nil {
		return PlaylistDto{}, err
	}
	return playlist.ToDto(), nil
}

func (s *Service) CreatePlaylist(req CreatePlaylistRequest) (PlaylistDto, error) {
	var result PlaylistModel

	err := s.withTransaction(func(tx *sql.Tx) error {
		playlist, err := s.Repository.CreatePlaylist(tx, req.Name, req.Description, false)
		if err != nil {
			return err
		}
		result = playlist
		return nil
	})

	if err != nil {
		return PlaylistDto{}, fmt.Errorf("erro ao criar playlist: %w", err)
	}

	return result.ToDto(), nil
}

func (s *Service) UpdatePlaylist(id int, req UpdatePlaylistRequest) (PlaylistDto, error) {
	if id < 0 {
		return PlaylistDto{}, ErrAutoPlaylistReadOnly
	}

	var result PlaylistModel

	err := s.withTransaction(func(tx *sql.Tx) error {
		playlist, err := s.Repository.UpdatePlaylist(tx, id, req.Name, req.Description)
		if err != nil {
			return err
		}
		result = playlist
		return nil
	})

	if err != nil {
		return PlaylistDto{}, fmt.Errorf("erro ao atualizar playlist: %w", err)
	}

	return result.ToDto(), nil
}

func (s *Service) DeletePlaylist(id int) error {
	if id < 0 {
		return ErrAutoPlaylistReadOnly
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.DeletePlaylist(tx, id)
	})
}

func (s *Service) GetPlaylistTracks(clientID string, playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackDto], error) {
	if playlistID < 0 {
		indexEntries, err := s.Repository.GetLibraryIndexEntries()
		if err != nil {
			return utils.PaginationResponse[PlaylistTrackDto]{}, err
		}

		fileIDs, err := s.automaticPlaylistTrackIDs(clientID, playlistID, indexEntries)
		if err != nil {
			return utils.PaginationResponse[PlaylistTrackDto]{}, err
		}

		return s.loadPlaylistTracksByIDs(fileIDs, page, pageSize)
	}

	tracksModel, err := s.Repository.GetPlaylistTracks(playlistID, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[PlaylistTrackDto]{}, err
	}
	return ParseTrackPaginationToDto(&tracksModel)
}

func (s *Service) AddPlaylistTrack(playlistID int, fileID int) (PlaylistTrackDto, error) {
	if playlistID < 0 {
		return PlaylistTrackDto{}, ErrAutoPlaylistReadOnly
	}

	var result PlaylistTrackModel

	err := s.withTransaction(func(tx *sql.Tx) error {
		track, err := s.Repository.AddPlaylistTrack(tx, playlistID, fileID)
		if err != nil {
			return err
		}
		result = track
		return nil
	})

	if err != nil {
		return PlaylistTrackDto{}, fmt.Errorf("erro ao adicionar track: %w", err)
	}

	dto, err := result.ToDto()
	if err != nil {
		return PlaylistTrackDto{}, err
	}

	return dto, nil
}

func (s *Service) RemovePlaylistTrack(playlistID int, fileID int) error {
	if playlistID < 0 {
		return ErrAutoPlaylistReadOnly
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.RemovePlaylistTrack(tx, playlistID, fileID)
	})
}

func (s *Service) ReorderPlaylistTracks(playlistID int, tracks []ReorderTrackItem) error {
	if playlistID < 0 {
		return ErrAutoPlaylistReadOnly
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		for _, track := range tracks {
			err := s.Repository.ReorderPlaylistTrack(tx, playlistID, track.FileID, track.Position)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) GetOrCreateNowPlaying() (PlaylistDto, error) {
	playlist, err := s.Repository.GetNowPlaying()
	if err == nil {
		return playlist.ToDto(), nil
	}

	// Now Playing queue doesn't exist, create it
	var result PlaylistModel
	err = s.withTransaction(func(tx *sql.Tx) error {
		p, err := s.Repository.CreatePlaylist(tx, i18n.GetMessage("MUSIC_NOW_PLAYING_NAME"), "", true)
		if err != nil {
			return err
		}
		result = p
		return nil
	})

	if err != nil {
		return PlaylistDto{}, fmt.Errorf("erro ao criar Now Playing queue: %w", err)
	}

	return result.ToDto(), nil
}

func (s *Service) GetPlayerState(clientID string) (PlayerStateDto, error) {
	state, err := s.Repository.GetPlayerState(clientID)
	if err != nil {
		return PlayerStateDto{}, err
	}
	return state.ToDto(), nil
}

func (s *Service) UpdatePlayerState(clientID string, req UpdatePlayerStateRequest) (PlayerStateDto, error) {
	state := PlayerStateModel{
		ClientID:        clientID,
		CurrentPosition: req.CurrentPosition,
		Volume:          req.Volume,
		Shuffle:         req.Shuffle,
		RepeatMode:      req.RepeatMode,
	}

	if req.PlaylistID != nil {
		state.PlaylistID.Valid = true
		state.PlaylistID.Int64 = int64(*req.PlaylistID)
	}
	if req.CurrentFileID != nil {
		state.CurrentFileID.Valid = true
		state.CurrentFileID.Int64 = int64(*req.CurrentFileID)
	}

	var result PlayerStateModel
	err := s.withTransaction(func(tx *sql.Tx) error {
		r, err := s.Repository.UpsertPlayerState(tx, state)
		if err != nil {
			return err
		}
		result = r
		return nil
	})

	if err != nil {
		return PlayerStateDto{}, fmt.Errorf("erro ao atualizar player state: %w", err)
	}

	result.ClientID = clientID
	result.PlaylistID = state.PlaylistID
	result.CurrentFileID = state.CurrentFileID
	result.CurrentPosition = req.CurrentPosition
	result.Volume = req.Volume
	result.Shuffle = req.Shuffle
	result.RepeatMode = req.RepeatMode

	return result.ToDto(), nil
}
