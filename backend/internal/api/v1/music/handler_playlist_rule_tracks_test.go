package music

import (
	"encoding/json"
	"net/http"
	"testing"

	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type rulePlaylistWritesServiceMock struct {
	musicHandlerServiceMock
	removedPlaylistID int
	removedFileID     int
}

func (m *rulePlaylistWritesServiceMock) AddPlaylistTrack(playlistID int, fileID int) (PlaylistTrackDto, error) {
	return PlaylistTrackDto{}, ErrPlaylistRuleReadOnly
}

func (m *rulePlaylistWritesServiceMock) ReorderPlaylistTracks(playlistID int, tracks []ReorderTrackItem) error {
	return ErrPlaylistRuleReadOnly
}

func (m *rulePlaylistWritesServiceMock) RemovePlaylistTrack(playlistID int, fileID int) error {
	m.removedPlaylistID = playlistID
	m.removedFileID = fileID
	return nil
}

func newRulePlaylistWritesRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.POST("/music/playlists/:id/tracks", handler.AddPlaylistTrackHandler)
	router.PUT("/music/playlists/:id/tracks/reorder", handler.ReorderPlaylistTracksHandler)
	router.DELETE("/music/playlists/:id/tracks/:fileId", handler.RemovePlaylistTrackHandler)
	return router
}

func TestAddAndReorderOnRulePlaylistAnswer409WithTranslatedMessage(t *testing.T) {
	router := newRulePlaylistWritesRouter(&rulePlaylistWritesServiceMock{})
	requests := []struct{ method, path, body string }{
		{http.MethodPost, "/music/playlists/3/tracks", `{"file_id":55}`},
		{http.MethodPut, "/music/playlists/3/tracks/reorder", `{"tracks":[{"file_id":55,"position":1}]}`},
	}

	for _, request := range requests {
		recorder := performPlaylistRequest(router, request.method, request.path, request.body)

		var body map[string]string
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: decode: %v", request.path, err)
		}
		expectedMessage := i18n.GetMessage("ERROR_MUSIC_PLAYLIST_RULE_READ_ONLY")
		if recorder.Code != http.StatusConflict || body["error"] == "" || body["error"] != expectedMessage {
			t.Fatalf("%s: expected 409 with %q, got %d %s", request.path, expectedMessage, recorder.Code, recorder.Body.String())
		}
	}
}

func TestRemoveTrackOnRulePlaylistForwardsPlaylistAndFileAndAnswersSuccess(t *testing.T) {
	service := &rulePlaylistWritesServiceMock{}
	router := newRulePlaylistWritesRouter(service)

	recorder := performPlaylistRequest(router, http.MethodDelete, "/music/playlists/3/tracks/55", "")

	if recorder.Code != http.StatusOK || recorder.Body.String() != `{"success":true}` {
		t.Fatalf("unexpected response %d %s", recorder.Code, recorder.Body.String())
	}
	if service.removedPlaylistID != 3 || service.removedFileID != 55 {
		t.Fatalf("unexpected removal: playlist=%d file=%d", service.removedPlaylistID, service.removedFileID)
	}
}
