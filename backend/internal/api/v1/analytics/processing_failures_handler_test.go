package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func registerProcessingFailures(router *gin.Engine, handler *Handler) {
	router.GET("/x", handler.GetProcessingFailuresHandler)
}

func TestHandlerProcessingFailuresEnvelopeAndItemKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &serviceStub{failuresToReturn: utils.PaginationResponse[ProcessingFailureDto]{
		Items:      []ProcessingFailureDto{{FileID: 3, Name: "a.jpg", Stage: "thumbnail", Error: "msg", FailedAt: time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)}},
		Pagination: utils.Pagination{Page: 2, PageSize: 5, HasNext: true, HasPrev: true},
	}}
	router := gin.New()
	registerProcessingFailures(router, NewHandler(stub))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/x?page=2&page_size=5&stage=thumbnail", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var envelope map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	assertKeys(t, "envelope", []string{"items", "pagination"}, envelope)
	assertKeys(t, "pagination", []string{"has_next", "has_prev", "page", "page_size"}, envelope["pagination"])
	assertKeys(t, "item", []string{"error", "failed_at", "file_id", "name", "stage"}, firstElement(t, envelope["items"]))
	if stub.capturedStage != ProcessingStageThumbnail || stub.capturedPage != 2 || stub.capturedPageSize != 5 {
		t.Fatalf("unexpected args: %q %d %d", stub.capturedStage, stub.capturedPage, stub.capturedPageSize)
	}
}

func TestHandlerProcessingFailuresDefaultsAndEmptyItems(t *testing.T) {
	stub, recorder := doRequest(t, registerProcessingFailures, "/x")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if stub.capturedStage != "" || stub.capturedPage != 1 || stub.capturedPageSize != defaultProcessingFailuresPageSize {
		t.Fatalf("unexpected defaults: %q %d %d", stub.capturedStage, stub.capturedPage, stub.capturedPageSize)
	}
}

func TestHandlerProcessingFailuresRejectsInvalidInput(t *testing.T) {
	for _, url := range []string{"/x?stage=bogus", "/x?page=abc", "/x?page_size=abc"} {
		stub, recorder := doRequest(t, registerProcessingFailures, url)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", url, recorder.Code)
		}
		if stub.calledFailuresPage {
			t.Fatalf("%s: service must not be called", url)
		}
		var body map[string]string
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] == "" {
			t.Fatalf("%s: expected error body, got %s", url, recorder.Body.String())
		}
	}
}

func TestHandlerProcessingFailuresServiceFailureDoesNotLeak(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerProcessingFailures(router, NewHandler(&serviceStub{err: http.ErrAbortHandler}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/x", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), http.ErrAbortHandler.Error()) {
		t.Fatalf("raw error leaked: %s", recorder.Body.String())
	}
}
