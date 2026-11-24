package video

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type seasonCapturingServiceMock struct {
	videoHandlerServiceMock
	requestedSeason *int
}

func (mock *seasonCapturingServiceMock) GetPlaylistItemsPage(clientID string, playlistID int, season *int, page int, pageSize int) (utils.PaginationResponse[VideoPlaylistItemDto], error) {
	mock.requestedSeason = season
	return utils.PaginationResponse[VideoPlaylistItemDto]{Items: []VideoPlaylistItemDto{}}, nil
}

func newSeasonItemsRouter(serviceMock *seasonCapturingServiceMock) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(serviceMock, &videoFilesServiceMock{}, &videoRecentServiceMock{}, &videoLoggerMock{})
	router := gin.New()
	router.GET("/video/playlists/:id/items", handler.GetPlaylistItemsPageHandler)
	return router
}

func TestGetPlaylistItemsPageHandlerRejectsInvalidSeason(t *testing.T) {
	useTranslationLocale(t, "pt-BR")
	for _, rawSeason := range []string{"abc", "-1", "1.5", ""} {
		serviceMock := &seasonCapturingServiceMock{}
		router := newSeasonItemsRouter(serviceMock)

		request := httptest.NewRequest(http.MethodGet, "/video/playlists/1/items?season="+rawSeason, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("season=%q: status = %d, want 400", rawSeason, recorder.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("season=%q: body: %v", rawSeason, err)
		}
		if body["error"] != "Parâmetro de temporada inválido" {
			t.Fatalf("season=%q: error not translated: %q", rawSeason, body["error"])
		}
		if serviceMock.requestedSeason != nil {
			t.Fatalf("season=%q: service must not be called", rawSeason)
		}
	}
}

func TestGetPlaylistItemsPageHandlerPassesSeasonToService(t *testing.T) {
	for _, testCase := range []struct {
		query string
		want  *int
	}{
		{"", nil},
		{"?season=0", intPointer(0)},
		{"?season=2", intPointer(2)},
	} {
		serviceMock := &seasonCapturingServiceMock{}
		router := newSeasonItemsRouter(serviceMock)

		request := httptest.NewRequest(http.MethodGet, "/video/playlists/1/items"+testCase.query, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%q: status = %d", testCase.query, recorder.Code)
		}
		if (serviceMock.requestedSeason == nil) != (testCase.want == nil) ||
			(testCase.want != nil && *serviceMock.requestedSeason != *testCase.want) {
			t.Fatalf("%q: season = %v, want %v", testCase.query, serviceMock.requestedSeason, testCase.want)
		}
	}
}
