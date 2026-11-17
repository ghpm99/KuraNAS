package music

import (
	"bytes"
	"encoding/json"
	"strings"
)

const genreSeparator = ";"

type optionalText struct {
	IsSet bool
	Value *string
}

func (field *optionalText) UnmarshalJSON(rawValue []byte) error {
	field.IsSet = true
	field.Value = nil
	if isJSONNull(rawValue) {
		return nil
	}
	var text string
	if err := json.Unmarshal(rawValue, &text); err != nil {
		return err
	}
	field.Value = nonBlankTextOrNil(text)
	return nil
}

type optionalInteger struct {
	IsSet bool
	Value *int
}

func (field *optionalInteger) UnmarshalJSON(rawValue []byte) error {
	field.IsSet = true
	field.Value = nil
	if isJSONNull(rawValue) {
		return nil
	}
	var number int
	if err := json.Unmarshal(rawValue, &number); err != nil {
		return err
	}
	field.Value = &number
	return nil
}

type optionalTextList struct {
	IsSet  bool
	Values []string
}

func (field *optionalTextList) UnmarshalJSON(rawValue []byte) error {
	field.IsSet = true
	field.Values = nil
	if isJSONNull(rawValue) {
		return nil
	}
	var texts []string
	if err := json.Unmarshal(rawValue, &texts); err != nil {
		return err
	}
	for _, text := range texts {
		if trimmedText := strings.TrimSpace(text); trimmedText != "" {
			field.Values = append(field.Values, trimmedText)
		}
	}
	return nil
}

func (field optionalTextList) joinedGenres() *string {
	if len(field.Values) == 0 {
		return nil
	}
	joinedGenres := strings.Join(field.Values, genreSeparator)
	return &joinedGenres
}

// TrackMetadataPatch is a partial correction of a track's metadata. A field
// that was not sent keeps its current correction; a field sent as null, empty
// text or an empty list goes back to the value of the file tag.
type TrackMetadataPatch struct {
	Title       optionalText     `json:"title"`
	Artist      optionalText     `json:"artist"`
	Album       optionalText     `json:"album"`
	AlbumArtist optionalText     `json:"album_artist"`
	Genres      optionalTextList `json:"genres"`
	Year        optionalInteger  `json:"year"`
	DiscNumber  optionalInteger  `json:"disc_number"`
	TrackNumber optionalInteger  `json:"track_number"`
}

func isJSONNull(rawValue []byte) bool {
	return bytes.Equal(bytes.TrimSpace(rawValue), []byte("null"))
}

func nonBlankTextOrNil(text string) *string {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return nil
	}
	return &trimmedText
}

func (patch TrackMetadataPatch) hasAnyField() bool {
	return patch.Title.IsSet || patch.Artist.IsSet || patch.Album.IsSet || patch.AlbumArtist.IsSet ||
		patch.Genres.IsSet || patch.Year.IsSet || patch.DiscNumber.IsSet || patch.TrackNumber.IsSet
}
