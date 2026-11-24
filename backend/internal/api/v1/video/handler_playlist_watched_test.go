package video

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/internal/api/v1/clientidentity"

	"github.com/gin-gonic/gin"
)

type playlistWatchedCapturingServiceMock struct {
	videoHandlerServiceMock
	receivedClientID   string
	receivedPlaylistID int
	receivedSeason     *int
	receivedWatched    bool
	returnedError      error
}

func (m *playlistWatchedCapturingServiceMock) SetPlaylistWatched(clientID string, playlistID int, season *int, isWatched bool) (int, error) {
	m.receivedClientID = clientID
	m.receivedPlaylistID = playlistID
	m.receivedSeason = season
	m.receivedWatched = isWatched
	return 12, m.returnedError
}

func performSetPlaylistWatchedRequest(t *testing.T, service ServiceInterface, target string, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &videoFilesServiceMock{}, &videoRecentServiceMock{}, &videoLoggerMock{})
	router := gin.New()
	router.PUT("/video/playlists/:id/watched", handler.SetPlaylistWatchedHandler)

	request := httptest.NewRequest(http.MethodPut, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.RemoteAddr = "10.9.8.7:4321"
	for name, headerValue := range headers {
		request.Header.Set(name, headerValue)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestSetPlaylistWatchedDecodesBodyAndUsesHeaderClientID(t *testing.T) {
	serviceMock := &playlistWatchedCapturingServiceMock{}
	recorder := performSetPlaylistWatchedRequest(t, serviceMock, "/video/playlists/5/watched", `{"watched":true,"season":2}`,
		map[string]string{clientidentity.Header: "header-client-1"})

	if recorder.Code != http.StatusOK || serviceMock.receivedClientID != "header-client-1" || serviceMock.receivedPlaylistID != 5 ||
		serviceMock.receivedSeason == nil || *serviceMock.receivedSeason != 2 || !serviceMock.receivedWatched {
		t.Fatalf("status=%d %+v", recorder.Code, serviceMock)
	}
	var body map[string]int
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["updated"] != 12 {
		t.Fatalf("unexpected body %s err=%v", recorder.Body.String(), err)
	}
}

func TestSetPlaylistWatchedWithoutSeasonFallsBackToClientIP(t *testing.T) {
	serviceMock := &playlistWatchedCapturingServiceMock{receivedWatched: true}
	recorder := performSetPlaylistWatchedRequest(t, serviceMock, "/video/playlists/5/watched", `{"watched":false}`, nil)

	if recorder.Code != http.StatusOK || serviceMock.receivedClientID != "10.9.8.7" || serviceMock.receivedSeason != nil || serviceMock.receivedWatched {
		t.Fatalf("status=%d %+v", recorder.Code, serviceMock)
	}
}

func TestSetPlaylistWatchedRejectsInvalidRequests(t *testing.T) {
	cases := []struct {
		name    string
		target  string
		body    string
		headers map[string]string
	}{
		{"missing watched", "/video/playlists/5/watched", `{}`, nil},
		{"malformed json", "/video/playlists/5/watched", `{`, nil},
		{"negative season", "/video/playlists/5/watched", `{"watched":true,"season":-1}`, nil},
		{"non numeric season", "/video/playlists/5/watched", `{"watched":true,"season":"a"}`, nil},
		{"non numeric playlist id", "/video/playlists/abc/watched", `{"watched":true}`, nil},
		{"malformed client id", "/video/playlists/5/watched", `{"watched":true}`, map[string]string{clientidentity.Header: "short"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			serviceMock := &playlistWatchedCapturingServiceMock{}
			recorder := performSetPlaylistWatchedRequest(t, serviceMock, testCase.target, testCase.body, testCase.headers)
			if recorder.Code != http.StatusBadRequest || serviceMock.receivedPlaylistID != 0 {
				t.Fatalf("status=%d service called=%+v", recorder.Code, serviceMock)
			}
		})
	}
}

func TestSetPlaylistWatchedMapsOwnerPlaylistTo409(t *testing.T) {
	serviceMock := &playlistWatchedCapturingServiceMock{returnedError: ErrPlaylistNotAutomaticSeries}
	recorder := performSetPlaylistWatchedRequest(t, serviceMock, "/video/playlists/5/watched", `{"watched":true}`, nil)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("status=%d, want 409", recorder.Code)
	}
}
