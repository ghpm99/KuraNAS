package music

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/internal/api/v1/files"

	"github.com/gin-gonic/gin"
)

type trackMetadataServiceMock struct {
	musicHandlerServiceMock
	updateError    error
	receivedFileID int
	receivedPatch  TrackMetadataPatch
	callCount      int
}

func (mock *trackMetadataServiceMock) UpdateTrackMetadata(fileID int, patch TrackMetadataPatch) (files.FileDto, error) {
	mock.callCount++
	mock.receivedFileID = fileID
	mock.receivedPatch = patch
	if mock.updateError != nil {
		return files.FileDto{}, mock.updateError
	}
	return files.FileDto{ID: fileID, Name: "song.mp3", Metadata: AudioMetadataModel{Title: "Fixed"}}, nil
}

func performTrackMetadataRequest(mock *trackMetadataServiceMock, path string, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(mock, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.PUT("/music/tracks/:file_id/metadata", handler.UpdateTrackMetadataHandler)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeErrorBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var errorBody map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &errorBody); err != nil {
		t.Fatalf("decode error body %s: %v", recorder.Body.String(), err)
	}
	return errorBody
}

func TestUpdateTrackMetadataHandlerDecodesEveryFieldIntoThePatch(t *testing.T) {
	mock := &trackMetadataServiceMock{}
	body := `{"title": " Song ", "artist": "Band", "album": "Debut", "album_artist": "The Band", "genres": ["Rock", " Pop "], "year": 1999, "disc_number": 2, "track_number": 7}`

	recorder := performTrackMetadataRequest(mock, "/music/tracks/42/metadata", body)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	patch := mock.receivedPatch
	if mock.receivedFileID != 42 {
		t.Fatalf("file id = %d", mock.receivedFileID)
	}
	textFields := map[string]struct {
		got  optionalText
		want string
	}{
		"title": {patch.Title, "Song"}, "artist": {patch.Artist, "Band"}, "album": {patch.Album, "Debut"}, "album_artist": {patch.AlbumArtist, "The Band"},
	}
	for name, field := range textFields {
		if !field.got.IsSet || field.got.Value == nil || *field.got.Value != field.want {
			t.Fatalf("%s = %+v, want %q", name, field.got, field.want)
		}
	}
	if !patch.Genres.IsSet || len(patch.Genres.Values) != 2 || patch.Genres.Values[0] != "Rock" || patch.Genres.Values[1] != "Pop" {
		t.Fatalf("genres = %+v", patch.Genres)
	}
	integerFields := map[string]struct {
		got  optionalInteger
		want int
	}{
		"year": {patch.Year, 1999}, "disc_number": {patch.DiscNumber, 2}, "track_number": {patch.TrackNumber, 7},
	}
	for name, field := range integerFields {
		if !field.got.IsSet || field.got.Value == nil || *field.got.Value != field.want {
			t.Fatalf("%s = %+v, want %d", name, field.got, field.want)
		}
	}
}

func TestUpdateTrackMetadataHandlerDistinguishesAbsentFromNull(t *testing.T) {
	mock := &trackMetadataServiceMock{}

	recorder := performTrackMetadataRequest(mock, "/music/tracks/42/metadata", `{"genres": null, "title": "Only"}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	patch := mock.receivedPatch
	if !patch.Genres.IsSet || len(patch.Genres.Values) != 0 {
		t.Fatalf("genres null must be set and empty: %+v", patch.Genres)
	}
	if !patch.Title.IsSet {
		t.Fatalf("title must be set: %+v", patch.Title)
	}
	if patch.Artist.IsSet || patch.Album.IsSet || patch.AlbumArtist.IsSet || patch.Year.IsSet || patch.DiscNumber.IsSet || patch.TrackNumber.IsSet {
		t.Fatalf("absent fields must stay unset: %+v", patch)
	}
}

func TestUpdateTrackMetadataHandlerReturnsTheTrackInTheLibraryFormat(t *testing.T) {
	recorder := performTrackMetadataRequest(&trackMetadataServiceMock{}, "/music/tracks/42/metadata", `{"title": "Fixed"}`)

	var track files.FileDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &track); err != nil {
		t.Fatalf("decode track: %v", err)
	}
	metadata, isObject := track.Metadata.(map[string]any)
	if track.ID != 42 || track.Name != "song.mp3" || !isObject || metadata["title"] != "Fixed" {
		t.Fatalf("unexpected track %+v", track)
	}
}

func TestUpdateTrackMetadataHandlerNamesTheRefusedFieldFromTheService(t *testing.T) {
	mock := &trackMetadataServiceMock{updateError: &fieldValidationError{Field: "year", MessageKey: "ERROR_MUSIC_METADATA_YEAR_INVALID"}}

	recorder := performTrackMetadataRequest(mock, "/music/tracks/42/metadata", `{"year": 5}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	errorBody := decodeErrorBody(t, recorder)
	if errorBody["field"] != "year" || errorBody["error"] == "" || errorBody["error"] == nil {
		t.Fatalf("unexpected error body %v", errorBody)
	}
}

func TestUpdateTrackMetadataHandlerNamesTheFieldWithAWrongType(t *testing.T) {
	mock := &trackMetadataServiceMock{}

	recorder := performTrackMetadataRequest(mock, "/music/tracks/42/metadata", `{"year": "abc"}`)

	if recorder.Code != http.StatusBadRequest || mock.callCount != 0 {
		t.Fatalf("expected 400 without calling the service, got %d calls=%d", recorder.Code, mock.callCount)
	}
	if errorBody := decodeErrorBody(t, recorder); errorBody["field"] != "year" {
		t.Fatalf("unexpected error body %v", errorBody)
	}
}

func TestUpdateTrackMetadataHandlerRejectsMalformedJSONWithoutField(t *testing.T) {
	mock := &trackMetadataServiceMock{}

	recorder := performTrackMetadataRequest(mock, "/music/tracks/42/metadata", `{"title": `)

	if recorder.Code != http.StatusBadRequest || mock.callCount != 0 {
		t.Fatalf("expected 400 without calling the service, got %d calls=%d", recorder.Code, mock.callCount)
	}
	errorBody := decodeErrorBody(t, recorder)
	if _, hasField := errorBody["field"]; hasField || errorBody["error"] == nil {
		t.Fatalf("unexpected error body %v", errorBody)
	}
}

func TestUpdateTrackMetadataHandlerRejectsNonNumericID(t *testing.T) {
	mock := &trackMetadataServiceMock{}

	recorder := performTrackMetadataRequest(mock, "/music/tracks/abc/metadata", `{}`)

	if recorder.Code != http.StatusBadRequest || mock.callCount != 0 {
		t.Fatalf("expected 400 without calling the service, got %d calls=%d", recorder.Code, mock.callCount)
	}
}

func TestUpdateTrackMetadataHandlerReturnsNotFoundWhenTrackIsMissing(t *testing.T) {
	recorder := performTrackMetadataRequest(&trackMetadataServiceMock{updateError: sql.ErrNoRows}, "/music/tracks/42/metadata", `{}`)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestUpdateTrackMetadataHandlerReturnsServerErrorOnFailure(t *testing.T) {
	recorder := performTrackMetadataRequest(&trackMetadataServiceMock{updateError: errors.New("boom")}, "/music/tracks/42/metadata", `{}`)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}
