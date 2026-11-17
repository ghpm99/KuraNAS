package music

import (
	"bytes"
	"encoding/json"
	"errors"
	"nas-go/api/internal/api/v1/files"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
)

type batchMetadataServiceMock struct {
	musicHandlerServiceMock
	result          BatchMetadataResultDto
	updateError     error
	receivedRequest BatchTracksMetadataRequest
	receivedDryRun  bool
	callCount       int
	singleCallCount int
}

func (mock *batchMetadataServiceMock) UpdateTracksMetadata(request BatchTracksMetadataRequest, isDryRun bool) (BatchMetadataResultDto, error) {
	mock.callCount++
	mock.receivedRequest = request
	mock.receivedDryRun = isDryRun
	return mock.result, mock.updateError
}

func (mock *batchMetadataServiceMock) UpdateTrackMetadata(fileID int, patch TrackMetadataPatch) (files.FileDto, error) {
	mock.singleCallCount++
	return files.FileDto{ID: fileID}, nil
}

func performBatchMetadataRequest(mock *batchMetadataServiceMock, path string, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(mock, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.PUT("/music/tracks/metadata", handler.UpdateTracksMetadataHandler)
	router.PUT("/music/tracks/:file_id/metadata", handler.UpdateTrackMetadataHandler)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestUpdateTracksMetadataHandlerDecodesTheRequestFieldByField(t *testing.T) {
	mock := &batchMetadataServiceMock{result: BatchMetadataResultDto{Failed: []BatchMetadataFailureDto{}}}
	body := `{"fields": {"title": "T", "genres": ["Rock", "Pop"], "year": 1999, "disc_number": null}, "file_ids": [3, 5, 8]}`

	recorder := performBatchMetadataRequest(mock, "/music/tracks/metadata", body)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	request := mock.receivedRequest
	if request.FileIDs == nil || !slices.Equal(*request.FileIDs, []int{3, 5, 8}) {
		t.Fatalf("file ids = %v", request.FileIDs)
	}
	if request.AlbumKey != "" || request.FolderKey != "" {
		t.Fatalf("album=%q folder=%q", request.AlbumKey, request.FolderKey)
	}
	fields := request.Fields
	if !fields.Title.IsSet || fields.Title.Value == nil || *fields.Title.Value != "T" {
		t.Fatalf("title = %+v", fields.Title)
	}
	if !slices.Equal(fields.Genres.Values, []string{"Rock", "Pop"}) {
		t.Fatalf("genres = %+v", fields.Genres)
	}
	if !fields.Year.IsSet || fields.Year.Value == nil || *fields.Year.Value != 1999 {
		t.Fatalf("year = %+v", fields.Year)
	}
	if !fields.DiscNumber.IsSet || fields.DiscNumber.Value != nil {
		t.Fatalf("disc_number null must be set and empty: %+v", fields.DiscNumber)
	}
	if fields.Artist.IsSet || fields.Album.IsSet || fields.AlbumArtist.IsSet || fields.TrackNumber.IsSet {
		t.Fatalf("absent fields must stay unset: %+v", fields)
	}
	if mock.receivedDryRun {
		t.Fatal("dry run must default to false")
	}
}

func TestUpdateTracksMetadataHandlerDecodesAlbumAndFolderSelectors(t *testing.T) {
	mock := &batchMetadataServiceMock{}

	performBatchMetadataRequest(mock, "/music/tracks/metadata", `{"fields": {"year": 2000}, "album_key": "album-x"}`)
	if mock.receivedRequest.AlbumKey != "album-x" || mock.receivedRequest.FileIDs != nil || mock.receivedRequest.FolderKey != "" {
		t.Fatalf("album request = %+v", mock.receivedRequest)
	}

	performBatchMetadataRequest(mock, "/music/tracks/metadata", `{"fields": {"year": 2000}, "folder_key": "/m/rock"}`)
	if mock.receivedRequest.FolderKey != "/m/rock" || mock.receivedRequest.FileIDs != nil || mock.receivedRequest.AlbumKey != "" {
		t.Fatalf("folder request = %+v", mock.receivedRequest)
	}
}

func TestUpdateTracksMetadataHandlerPassesDryRunFromTheQuery(t *testing.T) {
	for query, wantDryRun := range map[string]bool{"?dry_run=true": true, "?dry_run=false": false, "": false} {
		mock := &batchMetadataServiceMock{}

		recorder := performBatchMetadataRequest(mock, "/music/tracks/metadata"+query, `{"fields": {"year": 2000}, "album_key": "a"}`)

		if recorder.Code != http.StatusOK || mock.callCount != 1 || mock.receivedDryRun != wantDryRun {
			t.Fatalf("query %q: code=%d calls=%d dryRun=%v", query, recorder.Code, mock.callCount, mock.receivedDryRun)
		}
	}
}

func TestUpdateTracksMetadataHandlerRefusesUnreadableDryRunInsteadOfWriting(t *testing.T) {
	mock := &batchMetadataServiceMock{}

	recorder := performBatchMetadataRequest(mock, "/music/tracks/metadata?dry_run=yes", `{"fields": {"year": 2000}, "album_key": "a"}`)

	if recorder.Code != http.StatusBadRequest || mock.callCount != 0 {
		t.Fatalf("code=%d calls=%d", recorder.Code, mock.callCount)
	}
	if errorBody := decodeErrorBody(t, recorder); errorBody["field"] != "dry_run" {
		t.Fatalf("error body = %v", errorBody)
	}
}

func TestUpdateTracksMetadataHandlerNamesTheFieldOfEveryRefusal(t *testing.T) {
	refusals := map[string]*fieldValidationError{
		"selector": {Field: "selector", MessageKey: "ERROR_MUSIC_METADATA_SELECTOR_INVALID"},
		"file_ids": {Field: "file_ids", MessageKey: "ERROR_MUSIC_METADATA_FILE_IDS_INVALID"},
		"fields":   {Field: "fields", MessageKey: "ERROR_MUSIC_METADATA_FIELDS_EMPTY"},
		"year":     {Field: "year", MessageKey: "ERROR_MUSIC_METADATA_YEAR_INVALID"},
	}
	for wantField, refusal := range refusals {
		mock := &batchMetadataServiceMock{updateError: refusal}

		recorder := performBatchMetadataRequest(mock, "/music/tracks/metadata", `{"fields": {"year": 5}, "album_key": "a"}`)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", wantField, recorder.Code)
		}
		errorBody := decodeErrorBody(t, recorder)
		if errorBody["field"] != wantField || errorBody["error"] == nil || errorBody["error"] == "" {
			t.Fatalf("%s: error body = %v", wantField, errorBody)
		}
	}
}

func TestUpdateTracksMetadataHandlerNamesTheNestedFieldWithAWrongType(t *testing.T) {
	for body, wantField := range map[string]string{
		`{"fields": {"year": "abc"}, "album_key": "a"}`: "year",
		`{"fields": {"year": 1999}, "file_ids": "x"}`:   "file_ids",
	} {
		mock := &batchMetadataServiceMock{}

		recorder := performBatchMetadataRequest(mock, "/music/tracks/metadata", body)

		if recorder.Code != http.StatusBadRequest || mock.callCount != 0 {
			t.Fatalf("%s: code=%d calls=%d", body, recorder.Code, mock.callCount)
		}
		if errorBody := decodeErrorBody(t, recorder); errorBody["field"] != wantField {
			t.Fatalf("%s: error body = %v", body, errorBody)
		}
	}
}

func TestUpdateTracksMetadataHandlerRejectsMalformedJSONWithoutField(t *testing.T) {
	mock := &batchMetadataServiceMock{}

	recorder := performBatchMetadataRequest(mock, "/music/tracks/metadata", `{"fields": `)

	if recorder.Code != http.StatusBadRequest || mock.callCount != 0 {
		t.Fatalf("code=%d calls=%d", recorder.Code, mock.callCount)
	}
	if errorBody := decodeErrorBody(t, recorder); errorBody["field"] != nil {
		t.Fatalf("malformed body must not name a field: %v", errorBody)
	}
}

func TestUpdateTracksMetadataHandlerReturnsTheResultWithFailedAsArray(t *testing.T) {
	mock := &batchMetadataServiceMock{result: BatchMetadataResultDto{Matched: 2, Updated: 2, Failed: []BatchMetadataFailureDto{}}}

	recorder := performBatchMetadataRequest(mock, "/music/tracks/metadata", `{"fields": {"year": 2000}, "album_key": "a"}`)

	var rawBody map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &rawBody); err != nil {
		t.Fatalf("decode %s: %v", recorder.Body.String(), err)
	}
	if string(rawBody["failed"]) != "[]" {
		t.Fatalf("failed must be an empty array, got %s", rawBody["failed"])
	}
	var result BatchMetadataResultDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result.Matched != 2 || result.Updated != 2 {
		t.Fatalf("result = %+v", result)
	}
}

func TestUpdateTracksMetadataHandlerReturnsFailedItems(t *testing.T) {
	mock := &batchMetadataServiceMock{result: BatchMetadataResultDto{
		Matched: 1, Updated: 1, Failed: []BatchMetadataFailureDto{{FileID: 999, Error: "not found"}},
	}}

	recorder := performBatchMetadataRequest(mock, "/music/tracks/metadata", `{"fields": {"year": 2000}, "file_ids": [1, 999]}`)

	var result BatchMetadataResultDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if len(result.Failed) != 1 || result.Failed[0].FileID != 999 || result.Failed[0].Error != "not found" {
		t.Fatalf("failed = %+v", result.Failed)
	}
}

func TestUpdateTracksMetadataHandlerReturnsServerErrorWhenTheTransactionFails(t *testing.T) {
	mock := &batchMetadataServiceMock{updateError: errors.New("boom")}

	recorder := performBatchMetadataRequest(mock, "/music/tracks/metadata", `{"fields": {"year": 2000}, "album_key": "a"}`)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if errorBody := decodeErrorBody(t, recorder); errorBody["error"] == nil || errorBody["field"] != nil {
		t.Fatalf("error body = %v", errorBody)
	}
}

func TestBatchRouteAndSingleTrackRouteDoNotShadowEachOther(t *testing.T) {
	mock := &batchMetadataServiceMock{}

	performBatchMetadataRequest(mock, "/music/tracks/metadata", `{"fields": {"year": 2000}, "album_key": "a"}`)
	if mock.callCount != 1 || mock.singleCallCount != 0 {
		t.Fatalf("batch path: batch=%d single=%d", mock.callCount, mock.singleCallCount)
	}

	performBatchMetadataRequest(mock, "/music/tracks/42/metadata", `{"year": 2000}`)
	if mock.callCount != 1 || mock.singleCallCount != 1 {
		t.Fatalf("single path: batch=%d single=%d", mock.callCount, mock.singleCallCount)
	}
}
