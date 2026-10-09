package music

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type playlistEditingRecordingServiceMock struct {
	musicHandlerServiceMock
	receivedNameSearch    string
	receivedPlaylistID    int
	receivedReorderItems  []ReorderTrackItem
	receivedUpdateRequest UpdatePlaylistRequest
	receivedAddedFileID   int
	addFailure            error
}

func (m *playlistEditingRecordingServiceMock) GetPlaylists(page int, pageSize int, nameSearch string) (utils.PaginationResponse[PlaylistDto], error) {
	m.receivedNameSearch = nameSearch
	return m.musicHandlerServiceMock.GetPlaylists(page, pageSize, nameSearch)
}

func (m *playlistEditingRecordingServiceMock) ReorderPlaylistTracks(playlistID int, tracks []ReorderTrackItem) error {
	m.receivedPlaylistID = playlistID
	m.receivedReorderItems = tracks
	return nil
}

func (m *playlistEditingRecordingServiceMock) UpdatePlaylist(id int, req UpdatePlaylistRequest) (PlaylistDto, error) {
	m.receivedPlaylistID = id
	m.receivedUpdateRequest = req
	return PlaylistDto{ID: id, Name: req.Name, Description: req.Description}, nil
}

func (m *playlistEditingRecordingServiceMock) AddPlaylistTrack(playlistID int, fileID int) (PlaylistTrackDto, error) {
	m.receivedPlaylistID = playlistID
	m.receivedAddedFileID = fileID
	return PlaylistTrackDto{}, m.addFailure
}

func newPlaylistEditingRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.GET("/music/playlists/", handler.GetPlaylistsHandler)
	router.PUT("/music/playlists/:id", handler.UpdatePlaylistHandler)
	router.POST("/music/playlists/:id/tracks", handler.AddPlaylistTrackHandler)
	router.PUT("/music/playlists/:id/tracks/reorder", handler.ReorderPlaylistTracksHandler)
	return router
}

func performPlaylistRequest(router *gin.Engine, method string, path string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestReorderHandlerDecodesTrackPositionPayload(t *testing.T) {
	service := &playlistEditingRecordingServiceMock{}
	router := newPlaylistEditingRouter(service)

	recorder := performPlaylistRequest(router, http.MethodPut, "/music/playlists/7/tracks/reorder", `{"tracks":[{"file_id":42,"position":3}]}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.receivedPlaylistID != 7 {
		t.Fatalf("expected playlist 7, got %d", service.receivedPlaylistID)
	}
	if len(service.receivedReorderItems) != 1 || service.receivedReorderItems[0] != (ReorderTrackItem{FileID: 42, Position: 3}) {
		t.Fatalf("unexpected reorder items: %+v", service.receivedReorderItems)
	}
}

func TestUpdateHandlerDecodesRenamePayload(t *testing.T) {
	service := &playlistEditingRecordingServiceMock{}
	router := newPlaylistEditingRouter(service)

	recorder := performPlaylistRequest(router, http.MethodPut, "/music/playlists/9", `{"name":"Road trip","description":"long drives"}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.receivedPlaylistID != 9 || service.receivedUpdateRequest != (UpdatePlaylistRequest{Name: "Road trip", Description: "long drives"}) {
		t.Fatalf("unexpected update: id=%d request=%+v", service.receivedPlaylistID, service.receivedUpdateRequest)
	}
}

func TestAddTrackHandlerAnswersConflictWhenTrackAlreadyInPlaylist(t *testing.T) {
	service := &playlistEditingRecordingServiceMock{addFailure: ErrTrackAlreadyInPlaylist}
	router := newPlaylistEditingRouter(service)

	recorder := performPlaylistRequest(router, http.MethodPost, "/music/playlists/3/tracks", `{"file_id":55}`)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] == "" {
		t.Fatalf("expected translated error body, got %q (%v)", recorder.Body.String(), err)
	}
	if service.receivedPlaylistID != 3 || service.receivedAddedFileID != 55 {
		t.Fatalf("unexpected add: playlist=%d file=%d", service.receivedPlaylistID, service.receivedAddedFileID)
	}
}

func TestListHandlerForwardsNameSearch(t *testing.T) {
	service := &playlistEditingRecordingServiceMock{}
	router := newPlaylistEditingRouter(service)

	recorder := performPlaylistRequest(router, http.MethodGet, "/music/playlists/?q=road&page=1&page_size=20", "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if service.receivedNameSearch != "road" {
		t.Fatalf("expected search road, got %q", service.receivedNameSearch)
	}
}
