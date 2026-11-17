package music

import (
	"database/sql"
	"nas-go/api/pkg/i18n"
)

type trackSelection struct {
	matchedFileIDs []int
	failed         []BatchMetadataFailureDto
}

// UpdateTracksMetadata applies request.Fields to every track chosen by the
// request selector in a single transaction and reports how many matched, how
// many were updated and which requested ids could not be used. With isDryRun
// nothing is written. It returns a *fieldValidationError for a refused request.
func (s *Service) UpdateTracksMetadata(request BatchTracksMetadataRequest, isDryRun bool) (BatchMetadataResultDto, error) {
	if validationError := validateBatchTracksMetadataRequest(request); validationError != nil {
		return BatchMetadataResultDto{}, validationError
	}
	selection, err := s.selectBatchTracks(request)
	if err != nil {
		return BatchMetadataResultDto{}, err
	}
	result := BatchMetadataResultDto{Matched: len(selection.matchedFileIDs), Failed: selection.failed}
	if isDryRun || result.Matched == 0 {
		return result, nil
	}
	if err := s.applyPatchToTracks(selection.matchedFileIDs, request.Fields); err != nil {
		return BatchMetadataResultDto{}, err
	}
	result.Updated = result.Matched
	return result, nil
}

func (s *Service) selectBatchTracks(request BatchTracksMetadataRequest) (trackSelection, error) {
	switch {
	case request.FileIDs != nil:
		return s.selectIndexedTracks(*request.FileIDs)
	case request.AlbumKey != "":
		matchedFileIDs, err := s.TrackMetadataStore.ListTrackFileIDsByAlbumKey(request.AlbumKey)
		return trackSelection{matchedFileIDs: matchedFileIDs, failed: []BatchMetadataFailureDto{}}, err
	default:
		matchedFileIDs, err := s.TrackMetadataStore.ListTrackFileIDsByFolder(request.FolderKey)
		return trackSelection{matchedFileIDs: matchedFileIDs, failed: []BatchMetadataFailureDto{}}, err
	}
}

func (s *Service) selectIndexedTracks(requestedFileIDs []int) (trackSelection, error) {
	matchedFileIDs, err := s.TrackMetadataStore.FilterIndexedTrackFileIDs(requestedFileIDs)
	if err != nil {
		return trackSelection{}, err
	}
	return trackSelection{
		matchedFileIDs: matchedFileIDs,
		failed:         trackNotFoundFailures(requestedFileIDs, matchedFileIDs),
	}, nil
}

func trackNotFoundFailures(requestedFileIDs []int, matchedFileIDs []int) []BatchMetadataFailureDto {
	matchedFileIDSet := make(map[int]bool, len(matchedFileIDs))
	for _, matchedFileID := range matchedFileIDs {
		matchedFileIDSet[matchedFileID] = true
	}
	notFoundMessage := i18n.GetMessage("ERROR_MUSIC_NOT_FOUND")
	failures := []BatchMetadataFailureDto{}
	for _, requestedFileID := range requestedFileIDs {
		if !matchedFileIDSet[requestedFileID] {
			failures = append(failures, BatchMetadataFailureDto{FileID: requestedFileID, Error: notFoundMessage})
		}
	}
	return failures
}

func (s *Service) applyPatchToTracks(fileIDs []int, patch TrackMetadataPatch) error {
	compilationArtistLabel := i18n.GetMessage("MUSIC_VARIOUS_ARTISTS")
	return s.withTransaction(func(tx *sql.Tx) error {
		return s.TrackMetadataStore.ApplyTrackMetadataPatch(tx, fileIDs, patch, compilationArtistLabel)
	})
}
