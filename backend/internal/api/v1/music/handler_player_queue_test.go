package music

import (
	"encoding/json"
	"nas-go/api/internal/api/v1/clientidentity"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type playerClientRecordingServiceMock struct {
	musicHandlerServiceMock
	receivedClientIDs []string
	receivedRequest   ReplacePlayerQueueRequest
	replaceFailure    error
}

func (m *playerClientRecordingServiceMock) GetPlayerState(clientID string) (PlayerStateDto, error) {
	m.receivedClientIDs = append(m.receivedClientIDs, clientID)
	return PlayerStateDto{ClientID: clientID}, nil
}

func (m *playerClientRecordingServiceMock) GetPlayerQueue(clientID string) (PlayerQueueDto, error) {
	m.receivedClientIDs = append(m.receivedClientIDs, clientID)
	return PlayerQueueDto{Items: []MusicQueueEntryDto{}, CurrentIndex: 2}, nil
}

func (m *playerClientRecordingServiceMock) ReplacePlayerQueue(clientID string, request ReplacePlayerQueueRequest) error {
	m.receivedClientIDs = append(m.receivedClientIDs, clientID)
	m.receivedRequest = request
	return m.replaceFailure
}

func newPlayerStateRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.GET("/music/player-state/", handler.GetPlayerStateHandler)
	router.GET("/music/player-state/queue", handler.GetPlayerQueueHandler)
	router.PUT("/music/player-state/queue", handler.ReplacePlayerQueueHandler)
	router.POST("/music/player-state/queue", handler.ReplacePlayerQueueHandler)
	return router
}

func performPlayerRequest(router *gin.Engine, method string, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.RemoteAddr = "10.0.0.7:5555"
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestPlayerEndpointsPreferDeclaredClientIDOverIP(t *testing.T) {
	service := &playerClientRecordingServiceMock{}
	router := newPlayerStateRouter(service)
	deviceHeaders := map[string]string{clientidentity.Header: "3f2c1a9e-7b44-4c1d-9a55-0123456789ab"}

	performPlayerRequest(router, http.MethodGet, "/music/player-state/", "", deviceHeaders)
	performPlayerRequest(router, http.MethodGet, "/music/player-state/", "", nil)
	performPlayerRequest(router, http.MethodGet, "/music/player-state/queue?client_id=queryclient-0001", "", nil)

	want := []string{"3f2c1a9e-7b44-4c1d-9a55-0123456789ab", "10.0.0.7", "queryclient-0001"}
	for index, wantID := range want {
		if service.receivedClientIDs[index] != wantID {
			t.Fatalf("client ids = %v, want %v", service.receivedClientIDs, want)
		}
	}
}

func TestPlayerEndpointsRejectMalformedClientID(t *testing.T) {
	router := newPlayerStateRouter(&playerClientRecordingServiceMock{})
	malformedIDs := []string{"short", "has spaces in it", strings.Repeat("a", 65), "semi;colon-12345"}

	for _, malformedID := range malformedIDs {
		recorder := performPlayerRequest(router, http.MethodGet, "/music/player-state/", "", map[string]string{clientidentity.Header: malformedID})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("client id %q answered %d", malformedID, recorder.Code)
		}
	}
}

func TestReplacePlayerQueueDecodesFileIDsAndCurrentIndex(t *testing.T) {
	for _, method := range []string{http.MethodPut, http.MethodPost} {
		service := &playerClientRecordingServiceMock{}
		router := newPlayerStateRouter(service)

		recorder := performPlayerRequest(router, method, "/music/player-state/queue", `{"file_ids":[4,9,4],"current_index":1}`, nil)

		if recorder.Code != http.StatusNoContent {
			t.Fatalf("%s status = %d body=%s", method, recorder.Code, recorder.Body.String())
		}
		if len(service.receivedRequest.FileIDs) != 3 || service.receivedRequest.FileIDs[1] != 9 || service.receivedRequest.CurrentIndex != 1 {
			t.Fatalf("%s decoded request = %+v", method, service.receivedRequest)
		}
	}
}

func TestReplacePlayerQueueRejectsMalformedBodyAndInvalidQueue(t *testing.T) {
	router := newPlayerStateRouter(&playerClientRecordingServiceMock{})
	if recorder := performPlayerRequest(router, http.MethodPut, "/music/player-state/queue", `{"file_ids":"x"}`, nil); recorder.Code != http.StatusBadRequest {
		t.Fatalf("malformed body answered %d", recorder.Code)
	}

	invalidQueueRouter := newPlayerStateRouter(&playerClientRecordingServiceMock{replaceFailure: ErrInvalidPlayerQueue})
	if recorder := performPlayerRequest(invalidQueueRouter, http.MethodPut, "/music/player-state/queue", `{"file_ids":[1],"current_index":5}`, nil); recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid queue answered %d", recorder.Code)
	}

	failingRouter := newPlayerStateRouter(&musicHandlerErrServiceMock{})
	if recorder := performPlayerRequest(failingRouter, http.MethodPut, "/music/player-state/queue", `{"file_ids":[1],"current_index":0}`, nil); recorder.Code != http.StatusInternalServerError {
		t.Fatalf("service failure answered %d", recorder.Code)
	}
	if recorder := performPlayerRequest(failingRouter, http.MethodGet, "/music/player-state/queue", "", nil); recorder.Code != http.StatusInternalServerError {
		t.Fatalf("service failure on get answered %d", recorder.Code)
	}
}

func TestGetPlayerQueueReturnsItemsAndCurrentIndex(t *testing.T) {
	router := newPlayerStateRouter(&playerClientRecordingServiceMock{})

	recorder := performPlayerRequest(router, http.MethodGet, "/music/player-state/queue", "", nil)

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, hasItems := body["items"].([]any); !hasItems {
		t.Fatalf("body missing items array: %v", body)
	}
	if body["current_index"] != float64(2) {
		t.Fatalf("current_index = %v", body["current_index"])
	}
}
