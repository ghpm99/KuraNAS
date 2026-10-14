package video

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type customPlaylistServiceStub struct {
	ServiceInterface
	createdRequest CreateCustomPlaylistRequest
	deletedID      int
	createResult   VideoPlaylistDto
	failure        error
}

func (s *customPlaylistServiceStub) CreateCustomPlaylist(request CreateCustomPlaylistRequest) (VideoPlaylistDto, error) {
	s.createdRequest = request
	return s.createResult, s.failure
}

func (s *customPlaylistServiceStub) DeleteCustomPlaylist(playlistID int) error {
	s.deletedID = playlistID
	return s.failure
}

func customPlaylistRouter(stub *customPlaylistServiceStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(stub, nil, nil, nil)
	router := gin.New()
	router.POST("/video/playlists", handler.CreateCustomPlaylistHandler)
	router.DELETE("/video/playlists/:id", handler.DeleteCustomPlaylistHandler)
	return router
}

func performCustomPlaylistRequest(router *gin.Engine, method string, path string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestCreateCustomPlaylistHandlerDecodesBodyAndReturns201(t *testing.T) {
	stub := &customPlaylistServiceStub{createResult: VideoPlaylistDto{ID: 7, Name: "Favoritos", Type: "custom"}}
	recorder := performCustomPlaylistRequest(customPlaylistRouter(stub), http.MethodPost, "/video/playlists", `{"name":"Favoritos","video_id":42}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}
	if stub.createdRequest.Name != "Favoritos" || stub.createdRequest.VideoID == nil || *stub.createdRequest.VideoID != 42 {
		t.Fatalf("decoded request = %+v", stub.createdRequest)
	}
	var responseBody VideoPlaylistDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil || responseBody.ID != 7 {
		t.Fatalf("unexpected body %s (%v)", recorder.Body.String(), err)
	}
}

func TestCreateCustomPlaylistHandlerWithoutVideoLeavesVideoIDNil(t *testing.T) {
	stub := &customPlaylistServiceStub{}
	performCustomPlaylistRequest(customPlaylistRouter(stub), http.MethodPost, "/video/playlists", `{"name":"Sem video"}`)
	if stub.createdRequest.VideoID != nil {
		t.Fatalf("video id should be nil, got %v", *stub.createdRequest.VideoID)
	}
}

func TestCreateCustomPlaylistHandlerErrorStatuses(t *testing.T) {
	cases := []struct {
		name   string
		body   string
		err    error
		status int
	}{
		{"malformed body", `{`, nil, http.StatusBadRequest},
		{"empty name", `{"name":""}`, ErrPlaylistNameRequired, http.StatusBadRequest},
		{"long name", `{"name":"x"}`, ErrPlaylistNameTooLong, http.StatusBadRequest},
		{"missing video", `{"name":"x","video_id":1}`, sql.ErrNoRows, http.StatusNotFound},
		{"unexpected failure", `{"name":"x"}`, errors.New("boom"), http.StatusInternalServerError},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := performCustomPlaylistRequest(customPlaylistRouter(&customPlaylistServiceStub{failure: testCase.err}), http.MethodPost, "/video/playlists", testCase.body)
			if recorder.Code != testCase.status {
				t.Fatalf("status = %d, want %d (%s)", recorder.Code, testCase.status, recorder.Body.String())
			}
			if testCase.status == http.StatusInternalServerError && strings.Contains(recorder.Body.String(), "boom") {
				t.Fatalf("raw error leaked: %s", recorder.Body.String())
			}
		})
	}
}

func TestDeleteCustomPlaylistHandlerStatuses(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		err    error
		status int
	}{
		{"success", "/video/playlists/9", nil, http.StatusNoContent},
		{"invalid id", "/video/playlists/abc", nil, http.StatusBadRequest},
		{"not found", "/video/playlists/9", sql.ErrNoRows, http.StatusNotFound},
		{"automatic playlist", "/video/playlists/9", ErrPlaylistAutomaticNotDeletable, http.StatusConflict},
		{"unexpected failure", "/video/playlists/9", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			stub := &customPlaylistServiceStub{failure: testCase.err}
			recorder := performCustomPlaylistRequest(customPlaylistRouter(stub), http.MethodDelete, testCase.path, "")
			if recorder.Code != testCase.status {
				t.Fatalf("status = %d, want %d", recorder.Code, testCase.status)
			}
			if testCase.status == http.StatusNoContent && stub.deletedID != 9 {
				t.Fatalf("deleted id = %d, want 9", stub.deletedID)
			}
		})
	}
}
