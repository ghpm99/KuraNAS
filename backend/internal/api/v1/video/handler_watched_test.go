package video

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/internal/api/v1/clientidentity"

	"github.com/gin-gonic/gin"
)

type watchedCapturingServiceMock struct {
	videoHandlerServiceMock
	receivedClientID string
	receivedVideoID  int
	receivedWatched  bool
}

func (m *watchedCapturingServiceMock) SetVideoWatched(clientID string, videoID int, isWatched bool) error {
	m.receivedClientID = clientID
	m.receivedVideoID = videoID
	m.receivedWatched = isWatched
	return nil
}

func performSetWatchedRequest(t *testing.T, service ServiceInterface, target string, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &videoFilesServiceMock{}, &videoRecentServiceMock{}, &videoLoggerMock{})
	router := gin.New()
	router.PUT("/video/progress/:file_id/watched", handler.SetVideoWatchedHandler)

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

func TestSetVideoWatchedDecodesBodyAndUsesHeaderClientID(t *testing.T) {
	serviceMock := &watchedCapturingServiceMock{}
	recorder := performSetWatchedRequest(t, serviceMock, "/video/progress/7/watched", `{"watched":true}`,
		map[string]string{clientidentity.Header: "header-client-1"})
	if recorder.Code != http.StatusOK || serviceMock.receivedClientID != "header-client-1" || serviceMock.receivedVideoID != 7 || !serviceMock.receivedWatched {
		t.Fatalf("status=%d %+v", recorder.Code, serviceMock)
	}
	var body map[string]bool
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || !body["success"] {
		t.Fatalf("unexpected body %s err=%v", recorder.Body.String(), err)
	}
}

func TestSetVideoWatchedDecodesFalseAndFallsBackToClientIP(t *testing.T) {
	serviceMock := &watchedCapturingServiceMock{receivedWatched: true}
	recorder := performSetWatchedRequest(t, serviceMock, "/video/progress/7/watched", `{"watched":false}`, nil)
	if recorder.Code != http.StatusOK || serviceMock.receivedClientID != "10.9.8.7" || serviceMock.receivedWatched {
		t.Fatalf("status=%d %+v", recorder.Code, serviceMock)
	}
}

func TestSetVideoWatchedRejectsInvalidRequests(t *testing.T) {
	cases := []struct {
		name    string
		target  string
		body    string
		headers map[string]string
	}{
		{"missing watched", "/video/progress/7/watched", `{}`, nil},
		{"malformed json", "/video/progress/7/watched", `{`, nil},
		{"non numeric file id", "/video/progress/abc/watched", `{"watched":true}`, nil},
		{"malformed client id", "/video/progress/7/watched", `{"watched":true}`, map[string]string{clientidentity.Header: "short"}},
	}
	for _, testCase := range cases {
		if recorder := performSetWatchedRequest(t, &videoHandlerServiceMock{}, testCase.target, testCase.body, testCase.headers); recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", testCase.name, recorder.Code)
		}
	}
}

func TestSetVideoWatchedMapsServiceErrors(t *testing.T) {
	if recorder := performSetWatchedRequest(t, &videoHandlerErrServiceMock{}, "/video/progress/7/watched", `{"watched":true}`, nil); recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	if recorder := performSetWatchedRequest(t, &videoNotFoundServiceMock{}, "/video/progress/7/watched", `{"watched":true}`, nil); recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

type videoNotFoundServiceMock struct {
	videoHandlerServiceMock
}

func (m *videoNotFoundServiceMock) SetVideoWatched(clientID string, videoID int, isWatched bool) error {
	return sql.ErrNoRows
}
