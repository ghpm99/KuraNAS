package video

import (
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
