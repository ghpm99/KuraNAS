package music

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newSummaryRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&musicHandlerServiceMock{}, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.GET("/music/library/albums/:key", handler.GetLibraryAlbumSummaryHandler)
	router.GET("/music/library/artists/:key", handler.GetLibraryArtistSummaryHandler)
	router.GET("/music/library/artists/:key/albums", handler.GetLibraryAlbumsByArtistHandler)
	router.GET("/music/library/genres/:key", handler.GetLibraryGenreSummaryHandler)
	router.GET("/music/library/folders/:key", handler.GetLibraryFolderSummaryHandler)
	return router
}

func TestCatalogSummaryHandlersReturnSmallPayloads(t *testing.T) {
	router := newSummaryRouter()
	cases := []struct {
		path       string
		wantFields []string
	}{
		{"/music/library/albums/a", []string{"key", "name", "artist", "year", "track_count", "total_length_seconds", "disc_count"}},
		{"/music/library/artists/a", []string{"key", "name", "track_count", "album_count", "total_length_seconds"}},
		{"/music/library/genres/g", []string{"key", "name", "track_count", "total_length_seconds"}},
		{"/music/library/folders/f", []string{"key", "name", "track_count", "total_length_seconds"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, testCase.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
			}
			payload := map[string]any{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(payload) != len(testCase.wantFields) {
				t.Fatalf("payload = %v, want exactly fields %v", payload, testCase.wantFields)
			}
			for _, field := range testCase.wantFields {
				if _, exists := payload[field]; !exists {
					t.Fatalf("missing field %q in %v", field, payload)
				}
			}
		})
	}
}

func TestCatalogSummaryHandlersAnswerNotFoundForUnknownGroup(t *testing.T) {
	router := newSummaryRouter()
	for _, path := range []string{
		"/music/library/albums/missing",
		"/music/library/artists/missing",
		"/music/library/genres/missing",
		"/music/library/folders/missing",
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d", path, recorder.Code)
		}
	}
}

func TestArtistAlbumsHandlerReturnsPageAndReportsFailures(t *testing.T) {
	router := newSummaryRouter()

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/music/library/artists/a/albums?page=1&page_size=10", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/music/library/artists/broken/albums", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("failure status = %d", recorder.Code)
	}
}
