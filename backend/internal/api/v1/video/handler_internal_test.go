package video

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type videoHandlerServiceMock struct{}

func (m *videoHandlerServiceMock) StartPlayback(clientID string, videoID int, playlistID *int) (PlaybackSessionDto, error) {
	if videoID == 500 {
		return PlaybackSessionDto{}, errors.New("start error")
	}
	return PlaybackSessionDto{}, nil
}
func (m *videoHandlerServiceMock) GetPlaybackState(clientID string) (PlaybackSessionDto, error) {
	if clientID == "not-found" {
		return PlaybackSessionDto{}, errors.New("not found")
	}
	return PlaybackSessionDto{}, nil
}
func (m *videoHandlerServiceMock) UpdatePlaybackState(clientID string, req UpdatePlaybackStateRequest) (VideoPlaybackStateDto, error) {
	return VideoPlaybackStateDto{}, nil
}
func (m *videoHandlerServiceMock) NextVideo(clientID string) (PlaybackSessionDto, error) {
	return PlaybackSessionDto{}, nil
}
func (m *videoHandlerServiceMock) PreviousVideo(clientID string) (PlaybackSessionDto, error) {
	return PlaybackSessionDto{}, nil
}
func (m *videoHandlerServiceMock) GetHomeCatalog(clientID string, limit int) (VideoHomeCatalogDto, error) {
	return VideoHomeCatalogDto{}, nil
}
func (m *videoHandlerServiceMock) RebuildSmartPlaylists() error { return nil }
func (m *videoHandlerServiceMock) GetPlaylists(includeHidden bool) ([]VideoPlaylistDto, error) {
	return []VideoPlaylistDto{{ID: 1, Name: "p"}}, nil
}
func (m *videoHandlerServiceMock) GetPlaylistByID(clientID string, id int) (VideoPlaylistDto, error) {
	if id == 404 {
		return VideoPlaylistDto{}, errors.New("missing")
	}
	return VideoPlaylistDto{ID: id, Name: "p"}, nil
}
func (m *videoHandlerServiceMock) SetPlaylistHidden(playlistID int, hidden bool) error  { return nil }
func (m *videoHandlerServiceMock) AddVideoToPlaylist(playlistID int, videoID int) error { return nil }
func (m *videoHandlerServiceMock) RemoveVideoFromPlaylist(playlistID int, videoID int) error {
	return nil
}
func (m *videoHandlerServiceMock) GetUnassignedVideos(limit int) ([]VideoFileDto, error) {
	return []VideoFileDto{{ID: 1, Name: "v"}}, nil
}
func (m *videoHandlerServiceMock) UpdatePlaylistName(playlistID int, name string) error { return nil }
func (m *videoHandlerServiceMock) ReorderPlaylistItems(playlistID int, items []ReorderPlaylistItemRequest) error {
	return nil
}
func (m *videoHandlerServiceMock) TrackBehaviorEvent(clientID string, req TrackBehaviorEventRequest) error {
	return nil
}

type videoHandlerErrServiceMock struct {
	videoHandlerServiceMock
}

func (m *videoHandlerErrServiceMock) StartPlayback(clientID string, videoID int, playlistID *int) (PlaybackSessionDto, error) {
	return PlaybackSessionDto{}, errors.New("start failed")
}
func (m *videoHandlerErrServiceMock) GetPlaybackState(clientID string) (PlaybackSessionDto, error) {
	return PlaybackSessionDto{}, errors.New("state missing")
}
func (m *videoHandlerErrServiceMock) UpdatePlaybackState(clientID string, req UpdatePlaybackStateRequest) (VideoPlaybackStateDto, error) {
	return VideoPlaybackStateDto{}, errors.New("update state failed")
}
func (m *videoHandlerErrServiceMock) NextVideo(clientID string) (PlaybackSessionDto, error) {
	return PlaybackSessionDto{}, errors.New("next failed")
}
func (m *videoHandlerErrServiceMock) PreviousVideo(clientID string) (PlaybackSessionDto, error) {
	return PlaybackSessionDto{}, errors.New("previous failed")
}
func (m *videoHandlerErrServiceMock) GetHomeCatalog(clientID string, limit int) (VideoHomeCatalogDto, error) {
	return VideoHomeCatalogDto{}, errors.New("catalog failed")
}
func (m *videoHandlerErrServiceMock) RebuildSmartPlaylists() error {
	return errors.New("rebuild failed")
}
func (m *videoHandlerErrServiceMock) GetPlaylists(includeHidden bool) ([]VideoPlaylistDto, error) {
	return nil, errors.New("playlists failed")
}
func (m *videoHandlerErrServiceMock) GetPlaylistByID(clientID string, id int) (VideoPlaylistDto, error) {
	return VideoPlaylistDto{}, errors.New("playlist missing")
}
func (m *videoHandlerErrServiceMock) SetPlaylistHidden(playlistID int, hidden bool) error {
	return errors.New("set hidden failed")
}
func (m *videoHandlerErrServiceMock) AddVideoToPlaylist(playlistID int, videoID int) error {
	return errors.New("add failed")
}
func (m *videoHandlerErrServiceMock) RemoveVideoFromPlaylist(playlistID int, videoID int) error {
	return errors.New("remove failed")
}
func (m *videoHandlerErrServiceMock) GetUnassignedVideos(limit int) ([]VideoFileDto, error) {
	return nil, errors.New("unassigned failed")
}
func (m *videoHandlerErrServiceMock) UpdatePlaylistName(playlistID int, name string) error {
	return errors.New("update playlist failed")
}
func (m *videoHandlerErrServiceMock) ReorderPlaylistItems(playlistID int, items []ReorderPlaylistItemRequest) error {
	return errors.New("reorder failed")
}
func (m *videoHandlerErrServiceMock) TrackBehaviorEvent(clientID string, req TrackBehaviorEventRequest) error {
	return errors.New("track failed")
}

type videoLoggerMock struct{ logger.LoggerServiceInterface }

func TestVideoHandlerEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&videoHandlerServiceMock{}, &videoLoggerMock{})
	router := gin.New()

	router.POST("/video/playback/start", handler.StartPlaybackHandler)
	router.GET("/video/playback/state", handler.GetPlaybackStateHandler)
	router.PUT("/video/playback/state", handler.UpdatePlaybackStateHandler)
	router.POST("/video/playback/next", handler.NextVideoHandler)
	router.POST("/video/playback/previous", handler.PreviousVideoHandler)
	router.GET("/video/catalog/home", handler.GetHomeCatalogHandler)
	router.POST("/video/playlists/rebuild", handler.RebuildPlaylistsHandler)
	router.GET("/video/playlists", handler.GetPlaylistsHandler)
	router.GET("/video/playlists/:id", handler.GetPlaylistByIDHandler)
	router.PUT("/video/playlists/:id/hidden", handler.SetPlaylistHiddenHandler)
	router.POST("/video/playlists/:id/videos", handler.AddPlaylistVideoHandler)
	router.DELETE("/video/playlists/:id/videos/:videoId", handler.RemovePlaylistVideoHandler)
	router.PUT("/video/playlists/:id", handler.UpdatePlaylistHandler)
	router.PUT("/video/playlists/:id/reorder", handler.ReorderPlaylistHandler)
	router.GET("/video/playlists/unassigned", handler.GetUnassignedVideosHandler)

	tests := []struct {
		method string
		path   string
		body   string
		code   int
	}{
		{http.MethodPost, "/video/playback/start", `{"video_id":1}`, http.StatusOK},
		{http.MethodGet, "/video/playback/state", "", http.StatusOK},
		{http.MethodPut, "/video/playback/state", `{}`, http.StatusOK},
		{http.MethodPost, "/video/playback/next", "", http.StatusOK},
		{http.MethodPost, "/video/playback/previous", "", http.StatusOK},
		{http.MethodGet, "/video/catalog/home?limit=10", "", http.StatusOK},
		{http.MethodPost, "/video/playlists/rebuild", "", http.StatusOK},
		{http.MethodGet, "/video/playlists?include_hidden=true", "", http.StatusOK},
		{http.MethodGet, "/video/playlists/1", "", http.StatusOK},
		{http.MethodPut, "/video/playlists/1/hidden", `{"hidden":true}`, http.StatusOK},
		{http.MethodPost, "/video/playlists/1/videos", `{"video_id":10}`, http.StatusCreated},
		{http.MethodDelete, "/video/playlists/1/videos/10", "", http.StatusOK},
		{http.MethodPut, "/video/playlists/1", `{"name":"new"}`, http.StatusOK},
		{http.MethodPut, "/video/playlists/1/reorder", `{"items":[{"video_id":1,"order_index":0}]}`, http.StatusOK},
		{http.MethodGet, "/video/playlists/unassigned?limit=100", "", http.StatusOK},
		{http.MethodPost, "/video/playback/start", `{}`, http.StatusBadRequest},
		{http.MethodGet, "/video/playlists/404", "", http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected status %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}

func TestVideoHandlerErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&videoHandlerErrServiceMock{}, &videoLoggerMock{})
	router := gin.New()

	router.POST("/video/playback/start", handler.StartPlaybackHandler)
	router.GET("/video/playback/state", handler.GetPlaybackStateHandler)
	router.PUT("/video/playback/state", handler.UpdatePlaybackStateHandler)
	router.POST("/video/playback/next", handler.NextVideoHandler)
	router.POST("/video/playback/previous", handler.PreviousVideoHandler)
	router.GET("/video/catalog/home", handler.GetHomeCatalogHandler)
	router.POST("/video/playlists/rebuild", handler.RebuildPlaylistsHandler)
	router.GET("/video/playlists", handler.GetPlaylistsHandler)
	router.GET("/video/playlists/:id", handler.GetPlaylistByIDHandler)
	router.PUT("/video/playlists/:id/hidden", handler.SetPlaylistHiddenHandler)
	router.POST("/video/playlists/:id/videos", handler.AddPlaylistVideoHandler)
	router.DELETE("/video/playlists/:id/videos/:videoId", handler.RemovePlaylistVideoHandler)
	router.PUT("/video/playlists/:id", handler.UpdatePlaylistHandler)
	router.PUT("/video/playlists/:id/reorder", handler.ReorderPlaylistHandler)
	router.GET("/video/playlists/unassigned", handler.GetUnassignedVideosHandler)

	tests := []struct {
		method string
		path   string
		body   string
		code   int
	}{
		{http.MethodPost, "/video/playback/start", `{"video_id":1}`, http.StatusInternalServerError},
		{http.MethodGet, "/video/playback/state", "", http.StatusNotFound},
		{http.MethodPut, "/video/playback/state", `{"volume":0.5}`, http.StatusInternalServerError},
		{http.MethodPost, "/video/playback/next", "", http.StatusBadRequest},
		{http.MethodPost, "/video/playback/previous", "", http.StatusBadRequest},
		{http.MethodGet, "/video/catalog/home?limit=10", "", http.StatusInternalServerError},
		{http.MethodPost, "/video/playlists/rebuild", "", http.StatusInternalServerError},
		{http.MethodGet, "/video/playlists?include_hidden=true", "", http.StatusInternalServerError},
		{http.MethodGet, "/video/playlists/1", "", http.StatusNotFound},
		{http.MethodPut, "/video/playlists/1/hidden", `{"hidden":true}`, http.StatusInternalServerError},
		{http.MethodPost, "/video/playlists/1/videos", `{"video_id":10}`, http.StatusInternalServerError},
		{http.MethodDelete, "/video/playlists/1/videos/10", "", http.StatusInternalServerError},
		{http.MethodPut, "/video/playlists/1", `{"name":"new"}`, http.StatusInternalServerError},
		{http.MethodPut, "/video/playlists/1/reorder", `{"items":[{"video_id":1,"order_index":0}]}`, http.StatusInternalServerError},
		{http.MethodGet, "/video/playlists/unassigned?limit=100", "", http.StatusInternalServerError},
		{http.MethodPost, "/video/playback/start", `{}`, http.StatusBadRequest},
		{http.MethodPut, "/video/playback/state", `{`, http.StatusBadRequest},
		{http.MethodPut, "/video/playlists/1/hidden", `{}`, http.StatusInternalServerError},
		{http.MethodPost, "/video/playlists/1/videos", `{}`, http.StatusBadRequest},
		{http.MethodPut, "/video/playlists/1", `{}`, http.StatusBadRequest},
		{http.MethodPut, "/video/playlists/1/reorder", `{}`, http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected status %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}
