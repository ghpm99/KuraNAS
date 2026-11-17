package music

import (
	"database/sql"

	"nas-go/api/pkg/i18n"
)

// RenameGenre replaces the genre identified by genreKey with name on every
// active track that has it, as a user correction of the tracks' genre (the
// audio files are never written). The written text is the normalized genre
// label, the same one the genre groups show. Tracks that already had the new
// genre keep a single copy. It returns a *fieldValidationError for a refused
// name and sql.ErrNoRows when no active track has genreKey.
func (s *Service) RenameGenre(genreKey string, name string) (GenreRenameResultDto, error) {
	newLabel, validationError := validateGenreName(name)
	if validationError != nil {
		return GenreRenameResultDto{}, validationError
	}
	sources, err := s.TrackMetadataStore.ListGenreTrackSources(genreKey)
	if err != nil {
		return GenreRenameResultDto{}, err
	}
	if len(sources) == 0 {
		return GenreRenameResultDto{}, sql.ErrNoRows
	}
	newKey := genreLookupKey(newLabel)
	isMerge, err := s.isMergeIntoExistingGenre(genreKey, newKey)
	if err != nil {
		return GenreRenameResultDto{}, err
	}
	if err := s.applyGenreRename(buildGenreRenamePatchGroups(sources, genreKey, newLabel)); err != nil {
		return GenreRenameResultDto{}, err
	}
	renamedGenre, err := s.Repository.GetLibraryGenreSummary(newKey)
	if err != nil {
		return GenreRenameResultDto{}, err
	}
	return GenreRenameResultDto{
		Key:        renamedGenre.Key,
		Name:       renamedGenre.Name,
		TrackCount: renamedGenre.TrackCount,
		Merged:     isMerge,
	}, nil
}

func (s *Service) isMergeIntoExistingGenre(genreKey string, newKey string) (bool, error) {
	if newKey == genreKey {
		return false, nil
	}
	existingTrackCount, err := s.TrackMetadataStore.CountActiveTracksByGenreKey(newKey)
	return existingTrackCount > 0, err
}

func (s *Service) applyGenreRename(groups []TrackMetadataPatchGroup) error {
	compilationArtistLabel := i18n.GetMessage("MUSIC_VARIOUS_ARTISTS")
	return s.withTransaction(func(tx *sql.Tx) error {
		return s.TrackMetadataStore.ApplyTrackMetadataPatchGroups(tx, groups, compilationArtistLabel)
	})
}
