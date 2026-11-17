package music

import (
	"database/sql"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
)

// UpdateTrackMetadata validates and stores the user correction of one track and
// returns the track as GetLibraryTrackByID does. It returns a
// *fieldValidationError for a refused field and sql.ErrNoRows when the track is
// not an active indexed audio file. The audio file itself is never written.
func (s *Service) UpdateTrackMetadata(fileID int, patch TrackMetadataPatch) (files.FileDto, error) {
	if validationError := validateTrackMetadataPatch(patch); validationError != nil {
		return files.FileDto{}, validationError
	}
	if err := s.requireIndexedTrack(fileID); err != nil {
		return files.FileDto{}, err
	}
	compilationArtistLabel := i18n.GetMessage("MUSIC_VARIOUS_ARTISTS")
	err := s.withTransaction(func(tx *sql.Tx) error {
		return s.TrackMetadataStore.ApplyTrackMetadataPatch(tx, []int{fileID}, patch, compilationArtistLabel)
	})
	if err != nil {
		return files.FileDto{}, err
	}
	return s.GetLibraryTrackByID(fileID)
}

func (s *Service) requireIndexedTrack(fileID int) error {
	track, err := s.GetLibraryTrackByID(fileID)
	if err != nil {
		return err
	}
	metadata, isAudioMetadata := track.Metadata.(AudioMetadataModel)
	if !isAudioMetadata || metadata.ID == 0 {
		return sql.ErrNoRows
	}
	return nil
}
