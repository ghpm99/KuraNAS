package documenttext

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func performTextRequest(handler *Handler, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/documents/:file_id/text", handler.GetDocumentTextHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func TestGetDocumentTextHandlerFixesResponseFieldNames(t *testing.T) {
	service := &stubService{documentText: DocumentTextDto{Text: "olá mundo", IsPartial: true}}

	recorder := performTextRequest(NewHandler(service), "/documents/5/text")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload) != 2 || payload["text"] != "olá mundo" || payload["is_partial"] != true {
		t.Fatalf("unexpected payload %v", payload)
	}
}

func TestGetDocumentTextHandlerStatusMapping(t *testing.T) {
	testCases := map[string]struct {
		target     string
		err        error
		wantStatus int
	}{
		"not a document":  {"/documents/5/text", ErrDocumentNotFound, http.StatusNotFound},
		"service failure": {"/documents/5/text", errors.New("boom"), http.StatusInternalServerError},
		"invalid id":      {"/documents/abc/text", nil, http.StatusBadRequest},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			recorder := performTextRequest(NewHandler(&stubService{err: testCase.err}), testCase.target)
			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status %d, want %d", recorder.Code, testCase.wantStatus)
			}
		})
	}
}
