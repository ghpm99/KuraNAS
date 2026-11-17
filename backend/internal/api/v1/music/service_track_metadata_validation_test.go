package music

import (
	"strings"
	"testing"
)

func TestValidateTrackMetadataPatchReportsTheRefusedField(t *testing.T) {
	tooLongText := strings.Repeat("a", 256)
	testCases := []struct {
		name            string
		body            string
		expectedField   string
		expectedMessage string
	}{
		{"year below range", `{"year": 999}`, "year", "ERROR_MUSIC_METADATA_YEAR_INVALID"},
		{"year above range", `{"year": 10000}`, "year", "ERROR_MUSIC_METADATA_YEAR_INVALID"},
		{"disc zero", `{"disc_number": 0}`, "disc_number", "ERROR_MUSIC_METADATA_NUMBER_INVALID"},
		{"track negative", `{"track_number": -3}`, "track_number", "ERROR_MUSIC_METADATA_NUMBER_INVALID"},
		{"title too long", `{"title": "` + tooLongText + `"}`, "title", "ERROR_MUSIC_METADATA_TEXT_TOO_LONG"},
		{"artist too long", `{"artist": "` + tooLongText + `"}`, "artist", "ERROR_MUSIC_METADATA_TEXT_TOO_LONG"},
		{"album too long", `{"album": "` + tooLongText + `"}`, "album", "ERROR_MUSIC_METADATA_TEXT_TOO_LONG"},
		{"album artist too long", `{"album_artist": "` + tooLongText + `"}`, "album_artist", "ERROR_MUSIC_METADATA_TEXT_TOO_LONG"},
		{"one genre too long", `{"genres": ["Rock", "` + tooLongText + `"]}`, "genres", "ERROR_MUSIC_METADATA_TEXT_TOO_LONG"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			validationError := validateTrackMetadataPatch(decodePatch(t, testCase.body))

			if validationError == nil {
				t.Fatal("expected a validation error")
			}
			if validationError.Field != testCase.expectedField || validationError.MessageKey != testCase.expectedMessage {
				t.Fatalf("got field=%q key=%q, want field=%q key=%q", validationError.Field, validationError.MessageKey, testCase.expectedField, testCase.expectedMessage)
			}
		})
	}
}

func TestValidateTrackMetadataPatchAcceptsLimitValues(t *testing.T) {
	maxRunesText := strings.Repeat("é", 255)
	acceptedBodies := []string{
		`{}`,
		`{"year": 1000}`,
		`{"year": 9999}`,
		`{"disc_number": 1, "track_number": 1}`,
		`{"title": "` + maxRunesText + `", "genres": ["` + maxRunesText + `"]}`,
		`{"year": null, "title": null, "genres": null, "disc_number": null}`,
	}

	for _, body := range acceptedBodies {
		if validationError := validateTrackMetadataPatch(decodePatch(t, body)); validationError != nil {
			t.Fatalf("%.60s refused: %+v", body, validationError)
		}
	}
}

func TestValidateTrackMetadataPatchCountsRunesNotBytes(t *testing.T) {
	tooLongByRunes := strings.Repeat("é", 256)

	validationError := validateTrackMetadataPatch(decodePatch(t, `{"title": "`+tooLongByRunes+`"}`))

	if validationError == nil || validationError.Field != "title" {
		t.Fatalf("expected title refusal, got %+v", validationError)
	}
}
