package documenttext

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func performDocumentRequest(handler *Handler, method string, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/documents/:file_id/index-state", handler.GetIndexStateHandler)
	router.POST("/documents/:file_id/reindex", handler.ReindexDocumentHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(method, target, nil))
	return recorder
}

func TestGetIndexStateHandlerFixesResponseFieldNames(t *testing.T) {
	service := &stubService{indexState: IndexStateDto{FileID: 5, State: StatePartial, IndexedBytes: 2097152}}

	recorder := performDocumentRequest(NewHandler(service), http.MethodGet, "/documents/5/index-state")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload) != 5 || payload["file_id"] != float64(5) || payload["state"] != "partial" || payload["indexed_bytes"] != float64(2097152) || payload["reason"] != "" || payload["error"] != "" {
		t.Fatalf("unexpected payload %v", payload)
	}
}

func TestGetIndexStateHandlerStatusMapping(t *testing.T) {
	testCases := map[string]struct {
		target     string
		err        error
		wantStatus int
	}{
		"not a document":  {"/documents/5/index-state", ErrDocumentNotFound, http.StatusNotFound},
		"service failure": {"/documents/5/index-state", errors.New("boom"), http.StatusInternalServerError},
		"invalid id":      {"/documents/abc/index-state", nil, http.StatusBadRequest},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			recorder := performDocumentRequest(NewHandler(&stubService{err: testCase.err}), http.MethodGet, testCase.target)
			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status %d, want %d", recorder.Code, testCase.wantStatus)
			}
		})
	}
}

func TestReindexDocumentHandlerAcceptsWithJobID(t *testing.T) {
	service := &stubService{jobID: 77}

	recorder := performDocumentRequest(NewHandler(service), http.MethodPost, "/documents/5/reindex")

	if recorder.Code != http.StatusAccepted || service.reindexedFileID != 5 {
		t.Fatalf("status %d fileID %d", recorder.Code, service.reindexedFileID)
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload) != 1 || payload["job_id"] != float64(77) {
		t.Fatalf("unexpected payload %v", payload)
	}
}

func TestReindexDocumentHandlerStatusMapping(t *testing.T) {
	testCases := map[string]struct {
		target     string
		err        error
		wantStatus int
	}{
		"not a document":  {"/documents/5/reindex", ErrDocumentNotFound, http.StatusNotFound},
		"service failure": {"/documents/5/reindex", errors.New("boom"), http.StatusInternalServerError},
		"invalid id":      {"/documents/0/reindex", nil, http.StatusBadRequest},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			recorder := performDocumentRequest(NewHandler(&stubService{err: testCase.err}), http.MethodPost, testCase.target)
			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status %d, want %d", recorder.Code, testCase.wantStatus)
			}
		})
	}
}
