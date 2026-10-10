package music

import (
	"encoding/json"
	"nas-go/api/internal/api/v1/clientidentity"
	"net/http"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type playRecordingServiceMock struct {
	musicHandlerServiceMock
	receivedClientID string
	receivedRequest  RecordPlayRequest
	receivedPeriod   PlayPeriod
	recordFailure    error
}

func (m *playRecordingServiceMock) RecordPlay(clientID string, request RecordPlayRequest) error {
	m.receivedClientID = clientID
	m.receivedRequest = request
	return m.recordFailure
}

func (m *playRecordingServiceMock) GetMostPlayedTracks(period PlayPeriod, page int, pageSize int) (utils.PaginationResponse[MusicPlayedTrackDto], error) {
	m.receivedPeriod = period
	return m.musicHandlerServiceMock.GetMostPlayedTracks(period, page, pageSize)
}

func newPlayHistoryRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
	router := gin.New()
	router.POST("/music/plays", handler.RecordPlayHandler)
	router.GET("/music/library/most-played", handler.GetMostPlayedTracksHandler)
	router.GET("/music/library/recent-plays", handler.GetRecentlyPlayedTracksHandler)
	return router
}

func TestRecordPlayDecodesFileIDAndPlayedSecondsWithDeclaredClient(t *testing.T) {
	service := &playRecordingServiceMock{}
	router := newPlayHistoryRouter(service)

	recorder := performPlayerRequest(router, http.MethodPost, "/music/plays", `{"file_id":12,"played_seconds":31}`,
		map[string]string{clientidentity.Header: "3f2c1a9e-7b44-4c1d-9a55-0123456789ab"})

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.receivedRequest.FileID != 12 || service.receivedRequest.PlayedSeconds != 31 {
		t.Fatalf("decoded request = %+v", service.receivedRequest)
	}
	if service.receivedClientID != "3f2c1a9e-7b44-4c1d-9a55-0123456789ab" {
		t.Fatalf("client id = %q", service.receivedClientID)
	}
}

func TestRecordPlayFallsBackToClientIPAndRejectsBadInput(t *testing.T) {
	service := &playRecordingServiceMock{}
	router := newPlayHistoryRouter(service)

	performPlayerRequest(router, http.MethodPost, "/music/plays", `{"file_id":1,"played_seconds":30}`, nil)
	if service.receivedClientID != "10.0.0.7" {
		t.Fatalf("client id = %q", service.receivedClientID)
	}
	if recorder := performPlayerRequest(router, http.MethodPost, "/music/plays", `{"file_id":"x"}`, nil); recorder.Code != http.StatusBadRequest {
		t.Fatalf("malformed body answered %d", recorder.Code)
	}
	if recorder := performPlayerRequest(router, http.MethodPost, "/music/plays", `{}`, map[string]string{clientidentity.Header: "short"}); recorder.Code != http.StatusBadRequest {
		t.Fatalf("malformed client id answered %d", recorder.Code)
	}
}

func TestRecordPlayMapsServiceFailures(t *testing.T) {
	invalidRouter := newPlayHistoryRouter(&playRecordingServiceMock{recordFailure: ErrInvalidPlayRequest})
	if recorder := performPlayerRequest(invalidRouter, http.MethodPost, "/music/plays", `{"file_id":0}`, nil); recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid play answered %d", recorder.Code)
	}
	failingRouter := newPlayHistoryRouter(&musicHandlerErrServiceMock{})
	if recorder := performPlayerRequest(failingRouter, http.MethodPost, "/music/plays", `{"file_id":1}`, nil); recorder.Code != http.StatusInternalServerError {
		t.Fatalf("service failure answered %d", recorder.Code)
	}
}

func TestPlayHistoryListsDefaultToAllPeriodAndPassThePeriod(t *testing.T) {
	service := &playRecordingServiceMock{}
	router := newPlayHistoryRouter(service)

	recorder := performPlayerRequest(router, http.MethodGet, "/music/library/most-played", "", nil)
	if recorder.Code != http.StatusOK || service.receivedPeriod != PlayPeriodAll {
		t.Fatalf("status = %d period = %q", recorder.Code, service.receivedPeriod)
	}
	var body struct {
		Items []struct {
			PlayCount int `json:"play_count"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || len(body.Items) != 1 || body.Items[0].PlayCount != 3 {
		t.Fatalf("body = %s err=%v", recorder.Body.String(), err)
	}

	performPlayerRequest(router, http.MethodGet, "/music/library/most-played?period=30d", "", nil)
	if service.receivedPeriod != PlayPeriodLast30Days {
		t.Fatalf("period = %q", service.receivedPeriod)
	}
	if recorder := performPlayerRequest(router, http.MethodGet, "/music/library/recent-plays?page=1&page_size=5", "", nil); recorder.Code != http.StatusOK {
		t.Fatalf("recent plays answered %d", recorder.Code)
	}
}

func TestPlayHistoryListsRejectBadPaginationAndMapFailures(t *testing.T) {
	router := newPlayHistoryRouter(&musicHandlerServiceMock{})
	for _, path := range []string{"/music/library/most-played?page=x", "/music/library/recent-plays?page_size=x"} {
		if recorder := performPlayerRequest(router, http.MethodGet, path, "", nil); recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s answered %d", path, recorder.Code)
		}
	}
	failingRouter := newPlayHistoryRouter(&musicHandlerErrServiceMock{})
	for _, path := range []string{"/music/library/most-played", "/music/library/recent-plays"} {
		if recorder := performPlayerRequest(failingRouter, http.MethodGet, path, "", nil); recorder.Code != http.StatusInternalServerError {
			t.Fatalf("%s answered %d", path, recorder.Code)
		}
	}
}
