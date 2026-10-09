package music

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newQueueRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.GET("/music/library/artists/:key/queue", handler.GetLibraryQueueByArtistHandler)
	router.GET("/music/library/albums/:key/queue", handler.GetLibraryQueueByAlbumHandler)
	router.GET("/music/library/genres/:key/queue", handler.GetLibraryQueueByGenreHandler)
	router.GET("/music/library/folders/:key/queue", handler.GetLibraryQueueByFolderHandler)
	router.GET("/music/playlists/:id/queue", handler.GetPlaylistQueueHandler)
	return router
}

func requestQueue(router *gin.Engine, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

var queueEndpointPaths = []string{
	"/music/library/artists/k/queue",
	"/music/library/albums/k/queue",
	"/music/library/genres/k/queue",
	"/music/library/folders/k/queue",
	"/music/playlists/3/queue",
	"/music/playlists/-2/queue",
}

func TestQueueEndpointsReturnMinimalPayload(t *testing.T) {
	router := newQueueRouter(&musicHandlerServiceMock{})

	for _, path := range queueEndpointPaths {
		recorder := requestQueue(router, path)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, recorder.Code)
		}

		var body map[string]any
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s decode: %v", path, err)
		}
		if _, hasTruncated := body["truncated"]; !hasTruncated {
			t.Fatalf("%s missing truncated flag: %v", path, body)
		}
		items := body["items"].([]any)
		entry := items[0].(map[string]any)
		for _, field := range []string{"file_id", "name", "path", "format", "title", "artist", "album", "length"} {
			if _, hasField := entry[field]; !hasField {
				t.Fatalf("%s entry missing %q: %v", path, field, entry)
			}
		}
	}
}

func TestQueueEndpointsReportServiceFailure(t *testing.T) {
	router := newQueueRouter(&musicHandlerErrServiceMock{})

	for _, path := range queueEndpointPaths {
		if recorder := requestQueue(router, path); recorder.Code == http.StatusOK {
			t.Fatalf("%s answered 200 despite service failure", path)
		}
	}
}

func TestPlaylistQueueRejectsNonNumericID(t *testing.T) {
	router := newQueueRouter(&musicHandlerServiceMock{})

	if recorder := requestQueue(router, "/music/playlists/abc/queue"); recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}
