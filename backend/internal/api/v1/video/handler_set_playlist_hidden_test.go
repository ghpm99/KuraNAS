package video

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type setPlaylistHiddenServiceStub struct {
	ServiceInterface
	receivedPlaylistID int
	receivedHidden     bool
	failure            error
}

func (s *setPlaylistHiddenServiceStub) SetPlaylistHidden(playlistID int, hidden bool) error {
	s.receivedPlaylistID = playlistID
	s.receivedHidden = hidden
	return s.failure
}

func performSetPlaylistHiddenRequest(stub *setPlaylistHiddenServiceStub, path string, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/video/playlists/:id/hidden", NewHandler(stub, nil, nil, nil).SetPlaylistHiddenHandler)
	request := httptest.NewRequest(http.MethodPut, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeErrorMessage(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var responseBody map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("unexpected body %s (%v)", recorder.Body.String(), err)
	}
	return responseBody["error"]
}

func TestSetPlaylistHiddenHandlerDecodesBodyAndReturnsSuccess(t *testing.T) {
	for _, hidden := range []bool{true, false} {
		stub := &setPlaylistHiddenServiceStub{}
		recorder := performSetPlaylistHiddenRequest(stub, "/video/playlists/17/hidden", fmt.Sprintf(`{"hidden":%t}`, hidden))

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", recorder.Code)
		}
		if stub.receivedPlaylistID != 17 || stub.receivedHidden != hidden {
			t.Fatalf("decoded playlistID=%d hidden=%t, want 17/%t", stub.receivedPlaylistID, stub.receivedHidden, hidden)
		}
		var responseBody map[string]bool
		if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil || !responseBody["success"] {
			t.Fatalf("unexpected body %s (%v)", recorder.Body.String(), err)
		}
	}
}

func TestSetPlaylistHiddenHandlerReturnsTranslatedNotFound(t *testing.T) {
	stub := &setPlaylistHiddenServiceStub{failure: fmt.Errorf("playlist 9: %w", sql.ErrNoRows)}
	recorder := performSetPlaylistHiddenRequest(stub, "/video/playlists/9/hidden", `{"hidden":true}`)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	if message := decodeErrorMessage(t, recorder); message != i18n.GetMessage("ERROR_VIDEO_NOT_FOUND") {
		t.Fatalf("error = %q", message)
	}
}

func TestSetPlaylistHiddenHandlerReturnsTranslatedBadRequest(t *testing.T) {
	cases := map[string]string{
		"/video/playlists/9/hidden":   `not-json`,
		"/video/playlists/abc/hidden": `{"hidden":true}`,
	}
	for path, body := range cases {
		stub := &setPlaylistHiddenServiceStub{}
		recorder := performSetPlaylistHiddenRequest(stub, path, body)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want 400", path, recorder.Code)
		}
		if message := decodeErrorMessage(t, recorder); message != i18n.GetMessage("ERROR_INVALID_REQUEST") {
			t.Fatalf("%s error = %q", path, message)
		}
	}
}
