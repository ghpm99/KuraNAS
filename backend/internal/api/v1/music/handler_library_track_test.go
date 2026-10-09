package music

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/internal/api/v1/files"

	"github.com/gin-gonic/gin"
)

type libraryTrackServiceMock struct {
	musicHandlerServiceMock
	lookupError error
}

func (mock *libraryTrackServiceMock) GetLibraryTrackByID(fileID int) (files.FileDto, error) {
	if mock.lookupError != nil {
		return files.FileDto{}, mock.lookupError
	}
	return files.FileDto{ID: fileID, Name: "song.mp3"}, nil
}

func performLibraryTrackRequest(lookupError error, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&libraryTrackServiceMock{lookupError: lookupError}, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.GET("/music/library/tracks/:file_id", handler.GetLibraryTrackByIDHandler)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func TestGetLibraryTrackByIDHandlerReturnsSingleTrack(t *testing.T) {
	recorder := performLibraryTrackRequest(nil, "/music/library/tracks/42")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var track files.FileDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &track); err != nil {
		t.Fatalf("decode track: %v", err)
	}
	if track.ID != 42 || track.Name != "song.mp3" {
		t.Fatalf("unexpected track %+v", track)
	}
}

func TestGetLibraryTrackByIDHandlerReturnsNotFoundWhenTrackIsMissing(t *testing.T) {
	recorder := performLibraryTrackRequest(sql.ErrNoRows, "/music/library/tracks/42")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestGetLibraryTrackByIDHandlerReturnsServerErrorOnFailure(t *testing.T) {
	recorder := performLibraryTrackRequest(errors.New("boom"), "/music/library/tracks/42")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestGetLibraryTrackByIDHandlerRejectsNonNumericID(t *testing.T) {
	recorder := performLibraryTrackRequest(nil, "/music/library/tracks/abc")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}
