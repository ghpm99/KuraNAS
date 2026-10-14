package video

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

const maxCustomPlaylistNameLength = 120

func (s *Service) CreateCustomPlaylist(request CreateCustomPlaylistRequest) (VideoPlaylistDto, error) {
	playlistName := strings.TrimSpace(request.Name)
	if playlistName == "" {
		return VideoPlaylistDto{}, ErrPlaylistNameRequired
	}
	if utf8.RuneCountInString(playlistName) > maxCustomPlaylistNameLength {
		return VideoPlaylistDto{}, ErrPlaylistNameTooLong
	}
	if request.VideoID != nil {
		if _, err := s.Repository.GetVideoFileByID(*request.VideoID); err != nil {
			return VideoPlaylistDto{}, err
		}
	}
	sourcePath, err := newCustomPlaylistSourcePath()
	if err != nil {
		return VideoPlaylistDto{}, err
	}

	var createdPlaylist VideoPlaylistModel
	err = s.withTransaction(func(tx *sql.Tx) error {
		created, createErr := s.Repository.CreateCustomPlaylist(tx, sourcePath, playlistName)
		if createErr != nil {
			return createErr
		}
		createdPlaylist = created
		if request.VideoID == nil {
			return nil
		}
		createdPlaylist.ItemCount = 1
		return s.Repository.AddPlaylistVideoManual(tx, created.ID, *request.VideoID)
	})
	if err != nil {
		return VideoPlaylistDto{}, err
	}
	return createdPlaylist.ToDto(nil), nil
}

func (s *Service) DeleteCustomPlaylist(playlistID int) error {
	playlistToDelete, err := s.Repository.GetVideoPlaylistByID(playlistID)
	if err != nil {
		return err
	}
	if playlistToDelete.IsAuto {
		return ErrPlaylistAutomaticNotDeletable
	}
	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.DeleteCustomPlaylist(tx, playlistID)
	})
}

func newCustomPlaylistSourcePath() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return "custom/" + hex.EncodeToString(randomBytes), nil
}
