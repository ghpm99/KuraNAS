package files

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func serveTopFilesBySize(t *testing.T, service *filesHandlerServiceFuncMock, url string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/top-files-by-size", handler.GetTopFilesBySizeHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, url, nil))
	return recorder
}

func TestTopFilesBySizeWithoutPageKeepsFlatArray(t *testing.T) {
	requestedLimit := 0
	service := &filesHandlerServiceFuncMock{
		getTopFilesBySizeFn: func(limit int) ([]FileDto, error) {
			requestedLimit = limit
			return []FileDto{{ID: 7, Name: "big.bin", Size: 99}}, nil
		},
		getTopFilesBySizePageFn: func(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			t.Fatal("legacy request must not reach the paginated service")
			return utils.PaginationResponse[FileDto]{}, nil
		},
	}

	recorder := serveTopFilesBySize(t, service, "/files/top-files-by-size?limit=3")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var files []FileDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &files); err != nil {
		t.Fatalf("response without page must be a flat array: %v body=%s", err, recorder.Body.String())
	}
	if len(files) != 1 || files[0].ID != 7 || files[0].Name != "big.bin" || files[0].Size != 99 || requestedLimit != 3 {
		t.Fatalf("unexpected legacy response: %+v limit=%d", files, requestedLimit)
	}
}

func TestTopFilesBySizeWithPageReturnsPaginatedEnvelope(t *testing.T) {
	var requestedPage, requestedPageSize int
	service := &filesHandlerServiceFuncMock{
		getTopFilesBySizePageFn: func(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			requestedPage, requestedPageSize = page, pageSize
			return utils.PaginationResponse[FileDto]{
				Items: []FileDto{
					{ID: 7, Name: "big.bin", Path: "/data/big.bin", ParentPath: "/data", Format: ".bin", Type: File, Size: 99, Tier: TierHot},
					{ID: 8, Name: "small.bin", Path: "/data/small.bin", ParentPath: "/data", Format: ".bin", Type: File, Size: 5, Tier: TierCold},
				},
				Pagination: utils.Pagination{Page: 2, PageSize: 2, HasNext: true, HasPrev: true},
			}, nil
		},
	}

	recorder := serveTopFilesBySize(t, service, "/files/top-files-by-size?page=2&page_size=2")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var envelope utils.PaginationResponse[FileDto]
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, recorder.Body.String())
	}
	if envelope.Pagination != (utils.Pagination{Page: 2, PageSize: 2, HasNext: true, HasPrev: true}) {
		t.Fatalf("unexpected pagination: %+v", envelope.Pagination)
	}
	if len(envelope.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(envelope.Items))
	}
	biggest := envelope.Items[0]
	if biggest.ID != 7 || biggest.Name != "big.bin" || biggest.ParentPath == "" || biggest.Format != ".bin" || biggest.Size != 99 {
		t.Fatalf("unexpected first item: %+v", biggest)
	}
	if requestedPage != 2 || requestedPageSize != 2 {
		t.Fatalf("unexpected forwarded pagination: page=%d size=%d", requestedPage, requestedPageSize)
	}
}

func TestTopFilesBySizeWithPageUsesDefaultPageSize(t *testing.T) {
	var requestedPageSize int
	service := &filesHandlerServiceFuncMock{
		getTopFilesBySizePageFn: func(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			requestedPageSize = pageSize
			return utils.PaginationResponse[FileDto]{Items: []FileDto{}}, nil
		},
	}

	recorder := serveTopFilesBySize(t, service, "/files/top-files-by-size?page=1")

	if recorder.Code != http.StatusOK || requestedPageSize != defaultTopFilesBySizePageSize {
		t.Fatalf("expected 200 with default page size, got %d size=%d", recorder.Code, requestedPageSize)
	}
}

func TestTopFilesBySizeWithInvalidPaginationReturns400(t *testing.T) {
	for _, url := range []string{
		"/files/top-files-by-size?page=abc",
		"/files/top-files-by-size?page=1&page_size=abc",
	} {
		service := &filesHandlerServiceFuncMock{
			getTopFilesBySizePageFn: func(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
				t.Fatalf("%s: service must not be called", url)
				return utils.PaginationResponse[FileDto]{}, nil
			},
		}
		recorder := serveTopFilesBySize(t, service, url)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", url, recorder.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] == "" {
			t.Fatalf("%s: expected error body, got %s", url, recorder.Body.String())
		}
	}
}

func TestTopFilesBySizeWithPageServiceFailureReturns500(t *testing.T) {
	service := &filesHandlerServiceFuncMock{
		getTopFilesBySizePageFn: func(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
			return utils.PaginationResponse[FileDto]{}, errors.New("boom")
		},
	}

	recorder := serveTopFilesBySize(t, service, "/files/top-files-by-size?page=1")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
