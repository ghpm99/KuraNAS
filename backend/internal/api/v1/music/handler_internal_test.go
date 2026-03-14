package music

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"nas-go/api/internal/api/v1/files"
	"strings"
	"testing"

	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type musicHandlerServiceMock struct{}

func (m *musicHandlerServiceMock) GetPlaylists(page int, pageSize int) (utils.PaginationResponse[PlaylistDto], error) {
	return utils.PaginationResponse[PlaylistDto]{Items: []PlaylistDto{{ID: 1, Name: "p"}}}, nil
}
func (m *musicHandlerServiceMock) GetAutomaticPlaylists(clientID string) ([]PlaylistDto, error) {
	return []PlaylistDto{{ID: AutoPlaylistFavoritesID, Name: "favorites", IsAuto: true}}, nil
}
func (m *musicHandlerServiceMock) GetPlaylistByID(id int) (PlaylistDto, error) {
	if id == 404 {
		return PlaylistDto{}, errors.New("missing")
	}
	return PlaylistDto{ID: id, Name: "p"}, nil
}
func (m *musicHandlerServiceMock) CreatePlaylist(req CreatePlaylistRequest) (PlaylistDto, error) {
	return PlaylistDto{ID: 1, Name: req.Name}, nil
}
func (m *musicHandlerServiceMock) UpdatePlaylist(id int, req UpdatePlaylistRequest) (PlaylistDto, error) {
	return PlaylistDto{ID: id, Name: req.Name}, nil
}
func (m *musicHandlerServiceMock) DeletePlaylist(id int) error { return nil }
func (m *musicHandlerServiceMock) GetPlaylistTracks(clientID string, playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackDto], error) {
	return utils.PaginationResponse[PlaylistTrackDto]{Items: []PlaylistTrackDto{{ID: 1}}}, nil
}
func (m *musicHandlerServiceMock) AddPlaylistTrack(playlistID int, fileID int) (PlaylistTrackDto, error) {
	return PlaylistTrackDto{ID: 1}, nil
}
func (m *musicHandlerServiceMock) RemovePlaylistTrack(playlistID int, fileID int) error { return nil }
func (m *musicHandlerServiceMock) ReorderPlaylistTracks(playlistID int, tracks []ReorderTrackItem) error {
	return nil
}
func (m *musicHandlerServiceMock) GetOrCreateNowPlaying() (PlaylistDto, error) {
	return PlaylistDto{ID: 1}, nil
}
func (m *musicHandlerServiceMock) GetHomeCatalog(clientID string, limit int) (MusicHomeCatalogDto, error) {
	return MusicHomeCatalogDto{}, nil
}
func (m *musicHandlerServiceMock) GetLibraryTracks(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{Items: []files.FileDto{{ID: 1}}}, nil
}
func (m *musicHandlerServiceMock) GetLibraryArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistGroupDto], error) {
	return utils.PaginationResponse[MusicArtistGroupDto]{Items: []MusicArtistGroupDto{{Key: "artist", Artist: "Artist"}}}, nil
}
func (m *musicHandlerServiceMock) GetLibraryTracksByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{Items: []files.FileDto{{ID: 1}}}, nil
}
func (m *musicHandlerServiceMock) GetLibraryAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
	return utils.PaginationResponse[MusicAlbumGroupDto]{Items: []MusicAlbumGroupDto{{Key: "album", Album: "Album"}}}, nil
}
func (m *musicHandlerServiceMock) GetLibraryTracksByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{Items: []files.FileDto{{ID: 1}}}, nil
}
func (m *musicHandlerServiceMock) GetLibraryGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreGroupDto], error) {
	return utils.PaginationResponse[MusicGenreGroupDto]{Items: []MusicGenreGroupDto{{Key: "genre", Genre: "Genre"}}}, nil
}
func (m *musicHandlerServiceMock) GetLibraryTracksByGenre(genreKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{Items: []files.FileDto{{ID: 1}}}, nil
}
func (m *musicHandlerServiceMock) GetLibraryFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderGroupDto], error) {
	return utils.PaginationResponse[MusicFolderGroupDto]{Items: []MusicFolderGroupDto{{Folder: "/", TrackCount: 1}}}, nil
}
func (m *musicHandlerServiceMock) GetPlayerState(clientID string) (PlayerStateDto, error) {
	return PlayerStateDto{ID: 1, ClientID: clientID}, nil
}
func (m *musicHandlerServiceMock) UpdatePlayerState(clientID string, req UpdatePlayerStateRequest) (PlayerStateDto, error) {
	return PlayerStateDto{ID: 1, ClientID: clientID}, nil
}

type musicHandlerErrServiceMock struct {
	musicHandlerServiceMock
}

func (m *musicHandlerErrServiceMock) GetPlaylists(page int, pageSize int) (utils.PaginationResponse[PlaylistDto], error) {
	return utils.PaginationResponse[PlaylistDto]{}, errors.New("list error")
}
func (m *musicHandlerErrServiceMock) GetAutomaticPlaylists(clientID string) ([]PlaylistDto, error) {
	return nil, errors.New("automatic playlists error")
}
func (m *musicHandlerErrServiceMock) CreatePlaylist(req CreatePlaylistRequest) (PlaylistDto, error) {
	return PlaylistDto{}, errors.New("create error")
}
func (m *musicHandlerErrServiceMock) UpdatePlaylist(id int, req UpdatePlaylistRequest) (PlaylistDto, error) {
	return PlaylistDto{}, errors.New("update error")
}
func (m *musicHandlerErrServiceMock) DeletePlaylist(id int) error { return errors.New("delete error") }
func (m *musicHandlerErrServiceMock) GetPlaylistTracks(clientID string, playlistID int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackDto], error) {
	return utils.PaginationResponse[PlaylistTrackDto]{}, errors.New("tracks error")
}
func (m *musicHandlerErrServiceMock) AddPlaylistTrack(playlistID int, fileID int) (PlaylistTrackDto, error) {
	return PlaylistTrackDto{}, errors.New("add error")
}
func (m *musicHandlerErrServiceMock) RemovePlaylistTrack(playlistID int, fileID int) error {
	return errors.New("remove error")
}
func (m *musicHandlerErrServiceMock) ReorderPlaylistTracks(playlistID int, tracks []ReorderTrackItem) error {
	return errors.New("reorder error")
}
func (m *musicHandlerErrServiceMock) GetOrCreateNowPlaying() (PlaylistDto, error) {
	return PlaylistDto{}, errors.New("now playing error")
}
func (m *musicHandlerErrServiceMock) GetHomeCatalog(clientID string, limit int) (MusicHomeCatalogDto, error) {
	return MusicHomeCatalogDto{}, errors.New("home catalog error")
}
func (m *musicHandlerErrServiceMock) GetLibraryTracks(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{}, errors.New("library tracks error")
}
func (m *musicHandlerErrServiceMock) GetLibraryArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistGroupDto], error) {
	return utils.PaginationResponse[MusicArtistGroupDto]{}, errors.New("library artists error")
}
func (m *musicHandlerErrServiceMock) GetLibraryTracksByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{}, errors.New("artist tracks error")
}
func (m *musicHandlerErrServiceMock) GetLibraryAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
	return utils.PaginationResponse[MusicAlbumGroupDto]{}, errors.New("library albums error")
}
func (m *musicHandlerErrServiceMock) GetLibraryTracksByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{}, errors.New("album tracks error")
}
func (m *musicHandlerErrServiceMock) GetLibraryGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreGroupDto], error) {
	return utils.PaginationResponse[MusicGenreGroupDto]{}, errors.New("library genres error")
}
func (m *musicHandlerErrServiceMock) GetLibraryTracksByGenre(genreKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return utils.PaginationResponse[files.FileDto]{}, errors.New("genre tracks error")
}
func (m *musicHandlerErrServiceMock) GetLibraryFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderGroupDto], error) {
	return utils.PaginationResponse[MusicFolderGroupDto]{}, errors.New("library folders error")
}
func (m *musicHandlerErrServiceMock) GetPlayerState(clientID string) (PlayerStateDto, error) {
	return PlayerStateDto{}, errors.New("state error")
}
func (m *musicHandlerErrServiceMock) UpdatePlayerState(clientID string, req UpdatePlayerStateRequest) (PlayerStateDto, error) {
	return PlayerStateDto{}, errors.New("update state error")
}

type musicLoggerMock struct{ logger.LoggerServiceInterface }

func (m *musicLoggerMock) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *musicLoggerMock) CompleteWithSuccessLog(log logger.LoggerModel) error { return nil }
func (m *musicLoggerMock) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

func TestMusicHandlerEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&musicHandlerServiceMock{}, &musicLoggerMock{})
	router := gin.New()

	router.GET("/music/playlists", handler.GetPlaylistsHandler)
	router.GET("/music/playlists/:id", handler.GetPlaylistByIDHandler)
	router.POST("/music/playlists", handler.CreatePlaylistHandler)
	router.PUT("/music/playlists/:id", handler.UpdatePlaylistHandler)
	router.DELETE("/music/playlists/:id", handler.DeletePlaylistHandler)
	router.GET("/music/playlists/:id/tracks", handler.GetPlaylistTracksHandler)
	router.POST("/music/playlists/:id/tracks", handler.AddPlaylistTrackHandler)
	router.DELETE("/music/playlists/:id/tracks/:fileId", handler.RemovePlaylistTrackHandler)
	router.PUT("/music/playlists/:id/tracks/reorder", handler.ReorderPlaylistTracksHandler)
	router.GET("/music/playlists/now-playing", handler.GetNowPlayingHandler)
	router.GET("/music/player-state", handler.GetPlayerStateHandler)
	router.PUT("/music/player-state", handler.UpdatePlayerStateHandler)

	tests := []struct {
		method string
		path   string
		body   string
		code   int
	}{
		{http.MethodGet, "/music/playlists", "", http.StatusOK},
		{http.MethodGet, "/music/playlists/1", "", http.StatusOK},
		{http.MethodPost, "/music/playlists", `{"name":"n"}`, http.StatusCreated},
		{http.MethodPut, "/music/playlists/1", `{"name":"n2"}`, http.StatusOK},
		{http.MethodDelete, "/music/playlists/1", "", http.StatusOK},
		{http.MethodGet, "/music/playlists/1/tracks", "", http.StatusOK},
		{http.MethodPost, "/music/playlists/1/tracks", `{"file_id":2}`, http.StatusCreated},
		{http.MethodDelete, "/music/playlists/1/tracks/2", "", http.StatusOK},
		{http.MethodPut, "/music/playlists/1/tracks/reorder", `{"tracks":[{"file_id":2,"position":0}]}`, http.StatusOK},
		{http.MethodGet, "/music/playlists/now-playing", "", http.StatusOK},
		{http.MethodGet, "/music/player-state", "", http.StatusOK},
		{http.MethodPut, "/music/player-state", `{"volume":0.5}`, http.StatusOK},
		{http.MethodGet, "/music/playlists/404", "", http.StatusNotFound},
		{http.MethodPost, "/music/playlists", `{}`, http.StatusBadRequest},
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
				t.Fatalf("expected %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}

func TestMusicHandlerErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&musicHandlerErrServiceMock{}, &musicLoggerMock{})
	router := gin.New()

	router.GET("/music/playlists", handler.GetPlaylistsHandler)
	router.POST("/music/playlists", handler.CreatePlaylistHandler)
	router.PUT("/music/playlists/:id", handler.UpdatePlaylistHandler)
	router.DELETE("/music/playlists/:id", handler.DeletePlaylistHandler)
	router.GET("/music/playlists/:id/tracks", handler.GetPlaylistTracksHandler)
	router.POST("/music/playlists/:id/tracks", handler.AddPlaylistTrackHandler)
	router.DELETE("/music/playlists/:id/tracks/:fileId", handler.RemovePlaylistTrackHandler)
	router.PUT("/music/playlists/:id/tracks/reorder", handler.ReorderPlaylistTracksHandler)
	router.GET("/music/playlists/now-playing", handler.GetNowPlayingHandler)
	router.GET("/music/player-state", handler.GetPlayerStateHandler)
	router.PUT("/music/player-state", handler.UpdatePlayerStateHandler)

	tests := []struct {
		method string
		path   string
		body   string
		code   int
	}{
		{http.MethodGet, "/music/playlists", "", http.StatusInternalServerError},
		{http.MethodPost, "/music/playlists", `{"name":"x"}`, http.StatusInternalServerError},
		{http.MethodPut, "/music/playlists/1", `{"name":"x"}`, http.StatusInternalServerError},
		{http.MethodDelete, "/music/playlists/1", "", http.StatusInternalServerError},
		{http.MethodGet, "/music/playlists/1/tracks", "", http.StatusInternalServerError},
		{http.MethodPost, "/music/playlists/1/tracks", `{"file_id":2}`, http.StatusInternalServerError},
		{http.MethodDelete, "/music/playlists/1/tracks/2", "", http.StatusInternalServerError},
		{http.MethodPut, "/music/playlists/1/tracks/reorder", `{"tracks":[{"file_id":2,"position":0}]}`, http.StatusInternalServerError},
		{http.MethodGet, "/music/playlists/now-playing", "", http.StatusInternalServerError},
		{http.MethodGet, "/music/player-state", "", http.StatusNotFound},
		{http.MethodPut, "/music/player-state", `{"volume":0.5}`, http.StatusInternalServerError},
		{http.MethodPut, "/music/player-state", `{`, http.StatusBadRequest},
		{http.MethodPost, "/music/playlists", `{}`, http.StatusBadRequest},
		{http.MethodPut, "/music/playlists/1/tracks/reorder", `{}`, http.StatusBadRequest},
		{http.MethodPost, "/music/playlists/1/tracks", `{}`, http.StatusBadRequest},
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
				t.Fatalf("expected %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}
