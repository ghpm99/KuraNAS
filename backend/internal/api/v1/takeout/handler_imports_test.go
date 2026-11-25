package takeout

import (
	"encoding/json"
	"errors"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func performListImportsRequest(service ServiceInterface, query string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/takeout/imports", NewHandler(service, &handlerLoggerMock{}).ListImportsHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/takeout/imports"+query, nil))
	return recorder
}

func TestListImportsHandlerExposesEveryFieldAndPagination(t *testing.T) {
	jobID := 9
	endedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	var receivedPage, receivedPageSize int
	service := &handlerServiceMock{listFn: func(page int, pageSize int) (utils.PaginationResponse[TakeoutImportDto], error) {
		receivedPage, receivedPageSize = page, pageSize
		return utils.PaginationResponse[TakeoutImportDto]{
			Items: []TakeoutImportDto{{
				ID: 1, JobID: &jobID, FileName: "takeout.zip", Size: 2048, State: "completed",
				TotalEntries: 10, ProcessedEntries: 10, Imported: 6, AlreadyExisting: 2, Ignored: 1,
				FailedEntries: 1, Error: "", CreatedAt: endedAt, EndedAt: &endedAt,
			}},
			Pagination: utils.Pagination{Page: 2, PageSize: 5, HasNext: true, HasPrev: true},
		}, nil
	}}

	recorder := performListImportsRequest(service, "?page=2&page_size=5")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if receivedPage != 2 || receivedPageSize != 5 {
		t.Fatalf("service received page=%d page_size=%d", receivedPage, receivedPageSize)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	items := body["items"].([]any)
	importItem := items[0].(map[string]any)
	for _, fieldName := range []string{
		"id", "job_id", "file_name", "size", "state", "total_entries", "processed_entries",
		"imported", "already_existing", "ignored", "failed_entries", "error", "created_at", "ended_at",
	} {
		if _, exists := importItem[fieldName]; !exists {
			t.Errorf("missing field %q in %v", fieldName, importItem)
		}
	}
	pagination := body["pagination"].(map[string]any)
	for _, fieldName := range []string{"page", "page_size", "has_next", "has_prev"} {
		if _, exists := pagination[fieldName]; !exists {
			t.Errorf("missing pagination field %q", fieldName)
		}
	}
}

func TestListImportsHandlerRejectsInvalidPagination(t *testing.T) {
	recorder := performListImportsRequest(&handlerServiceMock{}, "?page=abc")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestListImportsHandlerHidesRawErrorOnFailure(t *testing.T) {
	service := &handlerServiceMock{listFn: func(int, int) (utils.PaginationResponse[TakeoutImportDto], error) {
		return utils.PaginationResponse[TakeoutImportDto]{}, errors.New("pq: secret detail")
	}}

	recorder := performListImportsRequest(service, "")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	if body["error"] != i18n.GetMessage("TAKEOUT_IMPORTS_LIST_ERROR") {
		t.Fatalf("error = %q", body["error"])
	}
}
