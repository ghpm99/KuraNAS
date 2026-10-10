package video

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/internal/api/v1/clientidentity"

	"github.com/gin-gonic/gin"
)

func performPlaybackStateRequest(t *testing.T, target string, headers map[string]string) (*httptest.ResponseRecorder, *clientIDCapturingServiceMock) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	serviceMock := &clientIDCapturingServiceMock{}
	handler := NewHandler(serviceMock, &videoFilesServiceMock{}, &videoRecentServiceMock{}, &videoLoggerMock{})
	router := gin.New()
	router.GET("/video/playback/state", handler.GetPlaybackStateHandler)

	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.RemoteAddr = "10.9.8.7:4321"
	for name, headerValue := range headers {
		request.Header.Set(name, headerValue)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder, serviceMock
}

type clientIDCapturingServiceMock struct {
	videoHandlerServiceMock
	receivedClientID string
}

func (m *clientIDCapturingServiceMock) GetPlaybackState(clientID string) (PlaybackSessionDto, error) {
	m.receivedClientID = clientID
	return PlaybackSessionDto{}, nil
}

func TestVideoPlaybackStateUsesHeaderClientID(t *testing.T) {
	recorder, serviceMock := performPlaybackStateRequest(t, "/video/playback/state?client_id=query-client-1",
		map[string]string{clientidentity.Header: "header-client-1"})
	if recorder.Code != http.StatusOK || serviceMock.receivedClientID != "header-client-1" {
		t.Fatalf("status=%d clientID=%q", recorder.Code, serviceMock.receivedClientID)
	}
}

func TestVideoPlaybackStateUsesQueryClientID(t *testing.T) {
	recorder, serviceMock := performPlaybackStateRequest(t, "/video/playback/state?client_id=query-client-1", nil)
	if recorder.Code != http.StatusOK || serviceMock.receivedClientID != "query-client-1" {
		t.Fatalf("status=%d clientID=%q", recorder.Code, serviceMock.receivedClientID)
	}
}

func TestVideoPlaybackStateFallsBackToClientIP(t *testing.T) {
	recorder, serviceMock := performPlaybackStateRequest(t, "/video/playback/state", nil)
	if recorder.Code != http.StatusOK || serviceMock.receivedClientID != "10.9.8.7" {
		t.Fatalf("status=%d clientID=%q", recorder.Code, serviceMock.receivedClientID)
	}
}

func TestVideoPlaybackStateRejectsMalformedClientID(t *testing.T) {
	recorder, _ := performPlaybackStateRequest(t, "/video/playback/state", map[string]string{clientidentity.Header: "short"})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", recorder.Code)
	}
}

type continueWatchingCapturingServiceMock struct {
	videoHandlerServiceMock
	receivedClientID string
	receivedLimit    int
}

func (m *continueWatchingCapturingServiceMock) GetContinueWatching(clientID string, limit int) ([]ContinueWatchingItemDto, error) {
	m.receivedClientID = clientID
	m.receivedLimit = limit
	return []ContinueWatchingItemDto{{Video: VideoFileDto{ID: 3}, PositionSeconds: 8, DurationSeconds: 40}}, nil
}

func performContinueWatchingRequest(t *testing.T, service ServiceInterface, target string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &videoFilesServiceMock{}, &videoRecentServiceMock{}, &videoLoggerMock{})
	router := gin.New()
	router.GET("/video/continue", handler.GetContinueWatchingHandler)

	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.RemoteAddr = "10.9.8.7:4321"
	for name, headerValue := range headers {
		request.Header.Set(name, headerValue)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestContinueWatchingUsesHeaderClientIDAndLimit(t *testing.T) {
	serviceMock := &continueWatchingCapturingServiceMock{}
	recorder := performContinueWatchingRequest(t, serviceMock, "/video/continue?limit=5",
		map[string]string{clientidentity.Header: "header-client-1"})
	if recorder.Code != http.StatusOK || serviceMock.receivedClientID != "header-client-1" || serviceMock.receivedLimit != 5 {
		t.Fatalf("status=%d clientID=%q limit=%d", recorder.Code, serviceMock.receivedClientID, serviceMock.receivedLimit)
	}
	var body []ContinueWatchingItemDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || len(body) != 1 || body[0].Video.ID != 3 || body[0].PositionSeconds != 8 || body[0].DurationSeconds != 40 {
		t.Fatalf("unexpected body %s err=%v", recorder.Body.String(), err)
	}
}

func TestContinueWatchingFallsBackToClientIPAndDefaultLimit(t *testing.T) {
	serviceMock := &continueWatchingCapturingServiceMock{}
	recorder := performContinueWatchingRequest(t, serviceMock, "/video/continue", nil)
	if recorder.Code != http.StatusOK || serviceMock.receivedClientID != "10.9.8.7" || serviceMock.receivedLimit != 24 {
		t.Fatalf("status=%d clientID=%q limit=%d", recorder.Code, serviceMock.receivedClientID, serviceMock.receivedLimit)
	}
}

func TestContinueWatchingRejectsInvalidLimitAndReportsServiceError(t *testing.T) {
	for _, target := range []string{"/video/continue?limit=0", "/video/continue?limit=101", "/video/continue?limit=abc"} {
		if recorder := performContinueWatchingRequest(t, &videoHandlerServiceMock{}, target, nil); recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", target, recorder.Code)
		}
	}
	if recorder := performContinueWatchingRequest(t, &videoHandlerErrServiceMock{}, "/video/continue", nil); recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
