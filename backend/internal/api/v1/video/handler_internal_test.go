package video

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

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
func (m *videoHandlerServiceMock) GetContinueWatching(clientID string, limit int) ([]ContinueWatchingItemDto, error) {
	return []ContinueWatchingItemDto{}, nil
}
func (m *videoHandlerServiceMock) SetVideoWatched(clientID string, videoID int, isWatched bool) error {
	return nil
}
func (m *videoHandlerServiceMock) RebuildSmartPlaylists() error { return nil }
func (m *videoHandlerServiceMock) GetPlaylists(includeHidden bool) ([]VideoPlaylistDto, error) {
	return []VideoPlaylistDto{{ID: 1, Name: "p"}}, nil
}
func (m *videoHandlerServiceMock) GetPlaylistsBySection(request PlaylistSectionRequest) (utils.PaginationResponse[VideoPlaylistDto], error) {
	return utils.PaginationResponse[VideoPlaylistDto]{Items: []VideoPlaylistDto{{ID: request.Page, Name: string(request.Section)}}}, nil
}
func (m *videoHandlerServiceMock) GetPlaylistMemberships(includeHidden bool) ([]VideoPlaylistMembershipDto, error) {
	return []VideoPlaylistMembershipDto{{PlaylistID: 1, VideoID: 10}}, nil
}
func (m *videoHandlerServiceMock) GetPlaylistByID(clientID string, id int) (VideoPlaylistDto, error) {
	if id == 404 {
		return VideoPlaylistDto{}, errors.New("missing")
	}
	return VideoPlaylistDto{ID: id, Name: "p"}, nil
}
func (m *videoHandlerServiceMock) GetPlaylistItemsPage(clientID string, playlistID int, page int, pageSize int) (utils.PaginationResponse[VideoPlaylistItemDto], error) {
	if playlistID == 404 {
		return utils.PaginationResponse[VideoPlaylistItemDto]{}, errors.New("missing")
	}
	return utils.PaginationResponse[VideoPlaylistItemDto]{
		Items:      []VideoPlaylistItemDto{{ID: page, OrderIndex: pageSize}},
		Pagination: utils.Pagination{Page: page, PageSize: pageSize},
	}, nil
}
func (m *videoHandlerServiceMock) ListLibraryVideos(page int, pageSize int, searchQuery string) (utils.PaginationResponse[VideoFileDto], error) {
	return utils.PaginationResponse[VideoFileDto]{Items: []VideoFileDto{{ID: 1, Name: "v"}}}, nil
}
func (m *videoHandlerServiceMock) ListLibraryFolders(request LibraryFolderRequest) (utils.PaginationResponse[LibraryFolderDto], error) {
	return utils.PaginationResponse[LibraryFolderDto]{Items: []LibraryFolderDto{{Path: request.ParentPath, Name: "f", VideoCount: 2, CoverFileID: 9}}}, nil
}
func (m *videoHandlerServiceMock) ListLibraryFolderVideos(request LibraryFolderVideosRequest) (utils.PaginationResponse[VideoFileDto], error) {
	return utils.PaginationResponse[VideoFileDto]{Items: []VideoFileDto{{ID: 1, Name: request.FolderPath}}}, nil
}
func (m *videoHandlerServiceMock) ListLibraryMovies(request LibraryMoviesRequest) (utils.PaginationResponse[VideoFileDto], error) {
	return utils.PaginationResponse[VideoFileDto]{Items: []VideoFileDto{{ID: 1, Name: string(request.Sort)}}}, nil
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
func (m *videoHandlerServiceMock) GetVideos(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{Items: []files.FileDto{{ID: 1}}}, nil
}
func (m *videoHandlerServiceMock) GetVideoThumbnail(fileDto files.FileDto, width, height int) ([]byte, error) {
	return []byte("png"), nil
}
func (m *videoHandlerServiceMock) GetVideoPreviewGif(fileDto files.FileDto, width, height int) ([]byte, error) {
	return []byte("gif"), nil
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
func (m *videoHandlerErrServiceMock) GetContinueWatching(clientID string, limit int) ([]ContinueWatchingItemDto, error) {
	return nil, errors.New("continue failed")
}
func (m *videoHandlerErrServiceMock) SetVideoWatched(clientID string, videoID int, isWatched bool) error {
	return errors.New("watched failed")
}
func (m *videoHandlerErrServiceMock) RebuildSmartPlaylists() error {
	return errors.New("rebuild failed")
}
func (m *videoHandlerErrServiceMock) GetPlaylists(includeHidden bool) ([]VideoPlaylistDto, error) {
	return nil, errors.New("playlists failed")
}
func (m *videoHandlerErrServiceMock) GetPlaylistsBySection(request PlaylistSectionRequest) (utils.PaginationResponse[VideoPlaylistDto], error) {
	return utils.PaginationResponse[VideoPlaylistDto]{}, errors.New("section playlists failed")
}
func (m *videoHandlerErrServiceMock) GetPlaylistMemberships(includeHidden bool) ([]VideoPlaylistMembershipDto, error) {
	return nil, errors.New("memberships failed")
}
func (m *videoHandlerErrServiceMock) GetPlaylistByID(clientID string, id int) (VideoPlaylistDto, error) {
	return VideoPlaylistDto{}, errors.New("playlist missing")
}
func (m *videoHandlerErrServiceMock) GetPlaylistItemsPage(clientID string, playlistID int, page int, pageSize int) (utils.PaginationResponse[VideoPlaylistItemDto], error) {
	return utils.PaginationResponse[VideoPlaylistItemDto]{}, errors.New("items failed")
}
func (m *videoHandlerErrServiceMock) ListLibraryVideos(page int, pageSize int, searchQuery string) (utils.PaginationResponse[VideoFileDto], error) {
	return utils.PaginationResponse[VideoFileDto]{}, errors.New("library failed")
}
func (m *videoHandlerErrServiceMock) ListLibraryFolders(request LibraryFolderRequest) (utils.PaginationResponse[LibraryFolderDto], error) {
	return utils.PaginationResponse[LibraryFolderDto]{}, errors.New("folders failed")
}
func (m *videoHandlerErrServiceMock) ListLibraryFolderVideos(request LibraryFolderVideosRequest) (utils.PaginationResponse[VideoFileDto], error) {
	return utils.PaginationResponse[VideoFileDto]{}, errors.New("folder videos failed")
}
func (m *videoHandlerErrServiceMock) ListLibraryMovies(request LibraryMoviesRequest) (utils.PaginationResponse[VideoFileDto], error) {
	return utils.PaginationResponse[VideoFileDto]{}, errors.New("movies failed")
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
func (m *videoHandlerErrServiceMock) GetVideos(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{}, errors.New("videos failed")
}
func (m *videoHandlerErrServiceMock) GetVideoThumbnail(fileDto files.FileDto, width, height int) ([]byte, error) {
	return nil, files.ErrFileMissingDisk
}
func (m *videoHandlerErrServiceMock) GetVideoPreviewGif(fileDto files.FileDto, width, height int) ([]byte, error) {
	return nil, errors.New("preview failed")
}

type videoLoggerMock struct{ logger.LoggerServiceInterface }

func (m *videoLoggerMock) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *videoLoggerMock) CompleteWithSuccessLog(log logger.LoggerModel) error { return nil }
func (m *videoLoggerMock) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

func TestVideoHandlerEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&videoHandlerServiceMock{}, &videoFilesServiceMock{}, &videoRecentServiceMock{}, &videoLoggerMock{})
	router := gin.New()

	router.POST("/video/playback/start", handler.StartPlaybackHandler)
	router.GET("/video/playback/state", handler.GetPlaybackStateHandler)
	router.PUT("/video/playback/state", handler.UpdatePlaybackStateHandler)
	router.POST("/video/playback/next", handler.NextVideoHandler)
	router.POST("/video/playback/previous", handler.PreviousVideoHandler)
	router.GET("/video/catalog/home", handler.GetHomeCatalogHandler)
	router.POST("/video/playlists/rebuild", handler.RebuildPlaylistsHandler)
	router.GET("/video/playlists", handler.GetPlaylistsHandler)
	router.GET("/video/playlists/section/:section", handler.GetPlaylistsBySectionHandler)
	router.GET("/video/playlists/memberships", handler.GetPlaylistMembershipsHandler)
	router.GET("/video/playlists/:id", handler.GetPlaylistByIDHandler)
	router.GET("/video/playlists/:id/items", handler.GetPlaylistItemsPageHandler)
	router.PUT("/video/playlists/:id/hidden", handler.SetPlaylistHiddenHandler)
	router.POST("/video/playlists/:id/videos", handler.AddPlaylistVideoHandler)
	router.DELETE("/video/playlists/:id/videos/:videoId", handler.RemovePlaylistVideoHandler)
	router.PUT("/video/playlists/:id", handler.UpdatePlaylistHandler)
	router.PUT("/video/playlists/:id/reorder", handler.ReorderPlaylistHandler)
	router.GET("/video/playlists/unassigned", handler.GetUnassignedVideosHandler)
	router.GET("/video/library/files", handler.ListLibraryVideosHandler)
	router.GET("/video/library/folders", handler.ListLibraryFoldersHandler)
	router.GET("/video/library/folders/videos", handler.ListLibraryFolderVideosHandler)
	router.GET("/video/library/movies", handler.ListLibraryMoviesHandler)

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
		{http.MethodGet, "/video/playlists/memberships?include_hidden=true", "", http.StatusOK},
		{http.MethodGet, "/video/playlists/section/series?page=2&page_size=4", "", http.StatusOK},
		{http.MethodGet, "/video/playlists/section/bogus", "", http.StatusBadRequest},
		{http.MethodGet, "/video/playlists/section/series?page=abc", "", http.StatusBadRequest},
		{http.MethodGet, "/video/playlists/1", "", http.StatusOK},
		{http.MethodGet, "/video/playlists/1/items", "", http.StatusOK},
		{http.MethodGet, "/video/playlists/1/items?page=2&page_size=10", "", http.StatusOK},
		{http.MethodGet, "/video/playlists/1/items?page=abc", "", http.StatusBadRequest},
		{http.MethodGet, "/video/playlists/404/items", "", http.StatusNotFound},
		{http.MethodPut, "/video/playlists/1/hidden", `{"hidden":true}`, http.StatusOK},
		{http.MethodPost, "/video/playlists/1/videos", `{"video_id":10}`, http.StatusCreated},
		{http.MethodDelete, "/video/playlists/1/videos/10", "", http.StatusOK},
		{http.MethodPut, "/video/playlists/1", `{"name":"new"}`, http.StatusOK},
		{http.MethodPut, "/video/playlists/1/reorder", `{"items":[{"video_id":1,"order_index":0}]}`, http.StatusOK},
		{http.MethodGet, "/video/playlists/unassigned?limit=100", "", http.StatusOK},
		{http.MethodGet, "/video/library/files?page=2&page_size=25&query=clip", "", http.StatusOK},
		{http.MethodGet, "/video/library/folders", "", http.StatusOK},
		{http.MethodGet, "/video/library/folders?parent=/Series&page=2&page_size=10", "", http.StatusOK},
		{http.MethodGet, "/video/library/folders/videos?path=/Series", "", http.StatusOK},
		{http.MethodGet, "/video/library/folders/videos", "", http.StatusBadRequest},
		{http.MethodGet, "/video/library/movies?sort=recent&page=2&page_size=10", "", http.StatusOK},
		{http.MethodGet, "/video/library/movies?sort=bogus", "", http.StatusBadRequest},
		{http.MethodGet, "/video/library/movies?page=abc", "", http.StatusBadRequest},
		{http.MethodGet, "/video/library/folders?page=abc", "", http.StatusBadRequest},
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
	handler := NewHandler(&videoHandlerErrServiceMock{}, &videoFilesServiceMock{}, &videoRecentServiceMock{}, &videoLoggerMock{})
	router := gin.New()

	router.POST("/video/playback/start", handler.StartPlaybackHandler)
	router.GET("/video/playback/state", handler.GetPlaybackStateHandler)
	router.PUT("/video/playback/state", handler.UpdatePlaybackStateHandler)
	router.POST("/video/playback/next", handler.NextVideoHandler)
	router.POST("/video/playback/previous", handler.PreviousVideoHandler)
	router.GET("/video/catalog/home", handler.GetHomeCatalogHandler)
	router.POST("/video/playlists/rebuild", handler.RebuildPlaylistsHandler)
	router.GET("/video/playlists", handler.GetPlaylistsHandler)
	router.GET("/video/playlists/section/:section", handler.GetPlaylistsBySectionHandler)
	router.GET("/video/playlists/memberships", handler.GetPlaylistMembershipsHandler)
	router.GET("/video/playlists/:id", handler.GetPlaylistByIDHandler)
	router.GET("/video/playlists/:id/items", handler.GetPlaylistItemsPageHandler)
	router.PUT("/video/playlists/:id/hidden", handler.SetPlaylistHiddenHandler)
	router.POST("/video/playlists/:id/videos", handler.AddPlaylistVideoHandler)
	router.DELETE("/video/playlists/:id/videos/:videoId", handler.RemovePlaylistVideoHandler)
	router.PUT("/video/playlists/:id", handler.UpdatePlaylistHandler)
	router.PUT("/video/playlists/:id/reorder", handler.ReorderPlaylistHandler)
	router.GET("/video/playlists/unassigned", handler.GetUnassignedVideosHandler)
	router.GET("/video/library/files", handler.ListLibraryVideosHandler)
	router.GET("/video/library/folders", handler.ListLibraryFoldersHandler)
	router.GET("/video/library/folders/videos", handler.ListLibraryFolderVideosHandler)
	router.GET("/video/library/movies", handler.ListLibraryMoviesHandler)

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
		{http.MethodGet, "/video/playlists/memberships?include_hidden=true", "", http.StatusInternalServerError},
		{http.MethodGet, "/video/playlists/section/series", "", http.StatusInternalServerError},
		{http.MethodGet, "/video/playlists/1", "", http.StatusNotFound},
		{http.MethodGet, "/video/playlists/1/items", "", http.StatusNotFound},
		{http.MethodPut, "/video/playlists/1/hidden", `{"hidden":true}`, http.StatusInternalServerError},
		{http.MethodPost, "/video/playlists/1/videos", `{"video_id":10}`, http.StatusInternalServerError},
		{http.MethodDelete, "/video/playlists/1/videos/10", "", http.StatusInternalServerError},
		{http.MethodPut, "/video/playlists/1", `{"name":"new"}`, http.StatusInternalServerError},
		{http.MethodPut, "/video/playlists/1/reorder", `{"items":[{"video_id":1,"order_index":0}]}`, http.StatusInternalServerError},
		{http.MethodGet, "/video/playlists/unassigned?limit=100", "", http.StatusInternalServerError},
		{http.MethodGet, "/video/library/files?query=test", "", http.StatusInternalServerError},
		{http.MethodGet, "/video/library/folders", "", http.StatusInternalServerError},
		{http.MethodGet, "/video/library/folders/videos?path=/Series", "", http.StatusInternalServerError},
		{http.MethodGet, "/video/library/movies", "", http.StatusInternalServerError},
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

func TestRespondVideoErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		err  error
		code int
	}{
		{"sql.ErrNoRows", sql.ErrNoRows, http.StatusNotFound},
		{"ErrPlaybackStateNotFound", ErrPlaybackStateNotFound, http.StatusNotFound},
		{"ErrVideoNotInPlaylist", ErrVideoNotInPlaylist, http.StatusBadRequest},
		{"ErrInvalidBehaviorEvent", ErrInvalidBehaviorEvent, http.StatusBadRequest},
		{"ErrPlaylistNameRequired", ErrPlaylistNameRequired, http.StatusBadRequest},
		{"ErrPlaylistReorderRequired", ErrPlaylistReorderRequired, http.StatusBadRequest},
		{"ErrPlaybackNavigation", ErrPlaybackNavigation, http.StatusBadRequest},
		{"ErrPlaylistWithoutItems", ErrPlaylistWithoutItems, http.StatusBadRequest},
		{"ErrNoVideosForContext", ErrNoVideosForContext, http.StatusNotFound},
		{"generic error", errors.New("unknown"), http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				respondVideoError(c, tc.err)
			})
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, w.Code)
			}
		})
	}
}
