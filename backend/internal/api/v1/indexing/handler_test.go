package indexing

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type serviceFake struct {
	receivedRequest ReindexRequestDto
	jobID           int
	err             error
}

func (fake *serviceFake) Reindex(request ReindexRequestDto) (int, error) {
	fake.receivedRequest = request
	return fake.jobID, fake.err
}

func newHandlerRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, nil)
	router := gin.New()
	router.POST("/indexing/reindex", handler.ReindexHandler)
	router.POST("/files/update", handler.UpdateFilesHandler)
	return router
}

func postReindex(router *gin.Engine, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/indexing/reindex", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v (%s)", err, recorder.Body.String())
	}
	return body
}

func TestReindexHandlerAcceptsEachScopeWithJobID(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		expected ReindexRequestDto
	}{
		{"all", `{"scope": "all"}`, ReindexRequestDto{Scope: "all"}},
		{"root", `{"scope": "root", "root_id": 2}`, ReindexRequestDto{Scope: "root", RootID: 2}},
		{"folder", `{"scope": "folder", "folder_id": 345}`, ReindexRequestDto{Scope: "folder", FolderID: 345}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			service := &serviceFake{jobID: 42}
			recorder := postReindex(newHandlerRouter(service), testCase.body)

			if recorder.Code != http.StatusAccepted {
				t.Fatalf("expected 202, got %d: %s", recorder.Code, recorder.Body.String())
			}
			if body := decodeBody(t, recorder); body["job_id"] != float64(42) {
				t.Fatalf("expected job_id 42, got %v", body)
			}
			if service.receivedRequest != testCase.expected {
				t.Fatalf("handler decoded %+v, want %+v", service.receivedRequest, testCase.expected)
			}
		})
	}
}

func TestReindexHandlerMapsErrorsToStatus(t *testing.T) {
	cases := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{"invalid scope", ErrInvalidScope, http.StatusBadRequest},
		{"identifier required", ErrIdentifierRequired, http.StatusBadRequest},
		{"root not found", ErrRootNotFound, http.StatusNotFound},
		{"folder not found", ErrFolderNotFound, http.StatusNotFound},
		{"unexpected failure", errors.New("db down"), http.StatusInternalServerError},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := postReindex(newHandlerRouter(&serviceFake{err: testCase.err}), `{"scope": "folder", "folder_id": 1}`)

			if recorder.Code != testCase.expectedStatus {
				t.Fatalf("expected %d, got %d", testCase.expectedStatus, recorder.Code)
			}
			if message, _ := decodeBody(t, recorder)["error"].(string); message == "" || strings.Contains(message, "db down") {
				t.Fatalf("error must be a translated message without the raw cause, got %q", message)
			}
		})
	}
}

func TestReindexHandlerAnswersConflictWithRunningJobID(t *testing.T) {
	recorder := postReindex(newHandlerRouter(&serviceFake{err: &AlreadyRunningError{JobID: 77}}), `{"scope": "all"}`)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}
	body := decodeBody(t, recorder)
	if body["job_id"] != float64(77) || body["error"] == "" {
		t.Fatalf("expected error and job_id 77, got %v", body)
	}
}

func TestReindexHandlerRejectsMalformedBody(t *testing.T) {
	recorder := postReindex(newHandlerRouter(&serviceFake{}), `not json`)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func postFilesUpdate(router *gin.Engine, form url.Values) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/files/update", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestUpdateFilesHandlerReindexesEverythingAndReturnsJobID(t *testing.T) {
	service := &serviceFake{jobID: 9}
	recorder := postFilesUpdate(newHandlerRouter(service), url.Values{"data": {"manual-rescan"}})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if decodeBody(t, recorder)["job_id"] != float64(9) {
		t.Fatalf("expected job_id 9, got %s", recorder.Body.String())
	}
	if service.receivedRequest.Scope != ScopeAll {
		t.Fatalf("files/update must reindex everything, got %+v", service.receivedRequest)
	}
}

func TestUpdateFilesHandlerReturnsRunningJobID(t *testing.T) {
	recorder := postFilesUpdate(newHandlerRouter(&serviceFake{err: &AlreadyRunningError{JobID: 5}}), url.Values{"data": {"x"}})

	if recorder.Code != http.StatusOK || decodeBody(t, recorder)["job_id"] != float64(5) {
		t.Fatalf("expected 200 with job_id 5, got %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestUpdateFilesHandlerRequiresData(t *testing.T) {
	recorder := postFilesUpdate(newHandlerRouter(&serviceFake{}), url.Values{})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestUpdateFilesHandlerReportsFailure(t *testing.T) {
	recorder := postFilesUpdate(newHandlerRouter(&serviceFake{err: errors.New("db down")}), url.Values{"data": {"x"}})
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
