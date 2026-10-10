package video

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type playbackStateCapturingServiceMock struct {
	videoHandlerServiceMock
	receivedClientID string
	receivedRequest  UpdatePlaybackStateRequest
}

func (m *playbackStateCapturingServiceMock) UpdatePlaybackState(clientID string, req UpdatePlaybackStateRequest) (VideoPlaybackStateDto, error) {
	m.receivedClientID = clientID
	m.receivedRequest = req
	return VideoPlaybackStateDto{}, nil
}

func TestPostPlaybackStateAliasDecodesBodyAndQueryClientID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	serviceMock := &playbackStateCapturingServiceMock{}
	handler := NewHandler(serviceMock, &videoFilesServiceMock{}, &videoRecentServiceMock{}, &videoLoggerMock{})
	router := gin.New()
	router.POST("/video/playback/state", handler.UpdatePlaybackStateHandler)

	body := `{"playlist_id":4,"video_id":9,"current_time":42.5,"duration":600,"is_paused":true,"completed":false}`
	request := httptest.NewRequest(http.MethodPost, "/video/playback/state?client_id=beacon-client-1", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	received := serviceMock.receivedRequest
	if serviceMock.receivedClientID != "beacon-client-1" {
		t.Fatalf("clientID=%q", serviceMock.receivedClientID)
	}
	if received.PlaylistID == nil || *received.PlaylistID != 4 ||
		received.VideoID == nil || *received.VideoID != 9 ||
		received.CurrentTime == nil || *received.CurrentTime != 42.5 ||
		received.Duration == nil || *received.Duration != 600 ||
		received.IsPaused == nil || !*received.IsPaused ||
		received.Completed == nil || *received.Completed {
		t.Fatalf("decoded request mismatch: %+v", received)
	}
}
