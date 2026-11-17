package music

import "unicode/utf8"

const (
	maxMetadataTextLength = 255
	minMetadataYear       = 1000
	maxMetadataYear       = 9999
	minMetadataNumber     = 1
)

type fieldValidationError struct {
	Field      string
	MessageKey string
}

func (validationError *fieldValidationError) Error() string {
	return validationError.MessageKey + ": " + validationError.Field
}

func validateTrackMetadataPatch(patch TrackMetadataPatch) *fieldValidationError {
	textFields := []struct {
		name  string
		field optionalText
	}{
		{"title", patch.Title},
		{"artist", patch.Artist},
		{"album", patch.Album},
		{"album_artist", patch.AlbumArtist},
	}
	for _, textField := range textFields {
		if textField.field.Value != nil && isTextTooLong(*textField.field.Value) {
			return newTextTooLongError(textField.name)
		}
	}
	for _, genre := range patch.Genres.Values {
		if isTextTooLong(genre) {
			return newTextTooLongError("genres")
		}
	}
	if isYearInvalid(patch.Year) {
		return &fieldValidationError{Field: "year", MessageKey: "ERROR_MUSIC_METADATA_YEAR_INVALID"}
	}
	if isNumberInvalid(patch.DiscNumber) {
		return newNumberInvalidError("disc_number")
	}
	if isNumberInvalid(patch.TrackNumber) {
		return newNumberInvalidError("track_number")
	}
	return nil
}

func isTextTooLong(text string) bool {
	return utf8.RuneCountInString(text) > maxMetadataTextLength
}

func isYearInvalid(year optionalInteger) bool {
	return year.Value != nil && (*year.Value < minMetadataYear || *year.Value > maxMetadataYear)
}

func isNumberInvalid(number optionalInteger) bool {
	return number.Value != nil && *number.Value < minMetadataNumber
}

func newTextTooLongError(fieldName string) *fieldValidationError {
	return &fieldValidationError{Field: fieldName, MessageKey: "ERROR_MUSIC_METADATA_TEXT_TOO_LONG"}
}

func newNumberInvalidError(fieldName string) *fieldValidationError {
	return &fieldValidationError{Field: fieldName, MessageKey: "ERROR_MUSIC_METADATA_NUMBER_INVALID"}
}
