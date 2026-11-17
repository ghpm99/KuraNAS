package music

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type genreRenameServiceMock struct {
	musicHandlerServiceMock
	result          GenreRenameResultDto
	renameError     error
	receivedKey     string
	receivedName    string
	renameCallCount int
}

func (mock *genreRenameServiceMock) RenameGenre(genreKey string, name string) (GenreRenameResultDto, error) {
	mock.renameCallCount++
	mock.receivedKey = genreKey
	mock.receivedName = name
	return mock.result, mock.renameError
}

func performGenreRenameRequest(mock *genreRenameServiceMock, path string, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(mock, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.PUT("/music/library/genres/:key", handler.RenameGenreHandler)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestRenameGenreHandlerPassesDecodedKeyAndNameAndReturnsTheResult(t *testing.T) {
	mock := &genreRenameServiceMock{result: GenreRenameResultDto{Key: "rock", Name: "Rock", TrackCount: 8, Merged: true}}

	recorder := performGenreRenameRequest(mock, "/music/library/genres/rock%20n%20roll", `{"name": "Rock"}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if mock.receivedKey != "rock n roll" || mock.receivedName != "Rock" {
		t.Fatalf("service received key=%q name=%q", mock.receivedKey, mock.receivedName)
	}
	var renamedGenre GenreRenameResultDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &renamedGenre); err != nil {
		t.Fatalf("decode %s: %v", recorder.Body.String(), err)
	}
	if renamedGenre != mock.result {
		t.Fatalf("response = %+v, want %+v", renamedGenre, mock.result)
	}
}

func TestRenameGenreHandlerRefusesUnreadableBodyWithoutField(t *testing.T) {
	for _, body := range []string{`not json`, `{"name": 5}`} {
		mock := &genreRenameServiceMock{}

		recorder := performGenreRenameRequest(mock, "/music/library/genres/rock", body)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", body, recorder.Code)
		}
		if _, hasField := decodeErrorBody(t, recorder)["field"]; hasField {
			t.Fatalf("%s: unreadable body must not name a field", body)
		}
		if mock.renameCallCount != 0 {
			t.Fatalf("%s: service must not be called", body)
		}
	}
}

func TestRenameGenreHandlerNamesTheRefusedNameField(t *testing.T) {
	mock := &genreRenameServiceMock{renameError: &fieldValidationError{Field: "name", MessageKey: "ERROR_MUSIC_GENRE_NAME_EMPTY"}}

	recorder := performGenreRenameRequest(mock, "/music/library/genres/rock", `{"name": "  "}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
	if field := decodeErrorBody(t, recorder)["field"]; field != "name" {
		t.Fatalf("field = %v, want name", field)
	}
}

func TestRenameGenreHandlerAnswersNotFoundWhenNoTrackHasTheGenre(t *testing.T) {
	mock := &genreRenameServiceMock{renameError: sql.ErrNoRows}

	recorder := performGenreRenameRequest(mock, "/music/library/genres/polka", `{"name": "Rock"}`)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}
