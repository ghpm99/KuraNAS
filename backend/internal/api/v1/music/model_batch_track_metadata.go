package music

const maxBatchTrackFileIDs = 500

// BatchTracksMetadataRequest applies one metadata patch to the tracks chosen by
// exactly one selector: file_ids, album_key or folder_key.
type BatchTracksMetadataRequest struct {
	Fields    TrackMetadataPatch `json:"fields"`
	FileIDs   *[]int             `json:"file_ids"`
	AlbumKey  string             `json:"album_key"`
	FolderKey string             `json:"folder_key"`
}

func (request BatchTracksMetadataRequest) selectorCount() int {
	selectorCount := 0
	if request.FileIDs != nil {
		selectorCount++
	}
	if request.AlbumKey != "" {
		selectorCount++
	}
	if request.FolderKey != "" {
		selectorCount++
	}
	return selectorCount
}

func validateBatchTracksMetadataRequest(request BatchTracksMetadataRequest) *fieldValidationError {
	if request.selectorCount() != 1 {
		return &fieldValidationError{Field: "selector", MessageKey: "ERROR_MUSIC_METADATA_SELECTOR_INVALID"}
	}
	if request.FileIDs != nil && (len(*request.FileIDs) == 0 || len(*request.FileIDs) > maxBatchTrackFileIDs) {
		return &fieldValidationError{Field: "file_ids", MessageKey: "ERROR_MUSIC_METADATA_FILE_IDS_INVALID"}
	}
	if !request.Fields.hasAnyField() {
		return &fieldValidationError{Field: "fields", MessageKey: "ERROR_MUSIC_METADATA_FIELDS_EMPTY"}
	}
	return validateTrackMetadataPatch(request.Fields)
}
