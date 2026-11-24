package ingest

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type fakeDownloadsService struct {
	retried     DownloadDto
	retryErr    error
	retriedID   int
	response    utils.PaginationResponse[DownloadDto]
	err         error
	stateFilter string
	page        int
	pageSize    int
	wasCalled   bool
}

func (f *fakeDownloadsService) ListDownloads(stateFilter string, page int, pageSize int) (utils.PaginationResponse[DownloadDto], error) {
	f.wasCalled = true
	f.stateFilter, f.page, f.pageSize = stateFilter, page, pageSize
	return f.response, f.err
}

func (f *fakeDownloadsService) RetryDownload(downloadID int) (DownloadDto, error) {
	f.retriedID = downloadID
	return f.retried, f.retryErr
}

func newDownloadsRouter(service DownloadsServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ingest/downloads", NewDownloadsHandler(service).ListDownloadsHandler)
	router.POST("/ingest/downloads/:id/retry", NewDownloadsHandler(service).RetryDownloadHandler)
	return router
}

func TestListDownloadsHandlerResponseFieldNames(t *testing.T) {
	jobID, fileID, eta := 345, 9, 95
	title, speed := "Video", int64(1048576)
	createdAt := time.Date(2026, 11, 16, 20, 15, 0, 0, time.UTC)
	service := &fakeDownloadsService{response: utils.PaginationResponse[DownloadDto]{
		Items: []DownloadDto{{
			ID: 12, JobID: &jobID, URL: "https://x.test/v", Title: &title, Preset: "video_1080",
			TargetRootID: 2, TargetRootName: "Videos", Subfolder: "Shows", State: "running", Progress: 42,
			SpeedBytesPerSecond: &speed, ETASeconds: &eta, FileID: &fileID, CreatedAt: createdAt,
		}},
		Pagination: utils.Pagination{Page: 1, PageSize: 20, HasNext: true},
	}}
	rec := doRequest(newDownloadsRouter(service), http.MethodGet, "/ingest/downloads?state=active&page=1&page_size=20", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if service.stateFilter != "active" || service.page != 1 || service.pageSize != 20 {
		t.Fatalf("service called with %q %d %d", service.stateFilter, service.page, service.pageSize)
	}
	var body struct {
		Items      []map[string]any `json:"items"`
		Pagination map[string]any   `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	assertExactFieldNames(t, body.Pagination, []string{"page", "page_size", "has_next", "has_prev"})
	if len(body.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(body.Items))
	}
	assertExactFieldNames(t, body.Items[0], []string{
		"id", "job_id", "url", "title", "preset", "target_root_id", "target_root_name", "subfolder",
		"state", "progress", "speed_bytes_per_second", "eta_seconds", "file_id", "error", "created_at", "ended_at",
	})
	if body.Items[0]["ended_at"] != nil || body.Items[0]["file_id"] != float64(9) || body.Items[0]["target_root_name"] != "Videos" {
		t.Fatalf("unexpected item values: %v", body.Items[0])
	}
}

func TestListDownloadsHandlerEmptyItemsIsArray(t *testing.T) {
	service := &fakeDownloadsService{response: utils.PaginationResponse[DownloadDto]{Items: []DownloadDto{}}}
	rec := doRequest(newDownloadsRouter(service), http.MethodGet, "/ingest/downloads", "")

	var body struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Items == nil {
		t.Fatalf("items must decode to an empty array: %v %s", err, rec.Body.String())
	}
	if service.page != 1 || service.pageSize != defaultDownloadsPageSize {
		t.Fatalf("defaults = %d %d", service.page, service.pageSize)
	}
}

func TestListDownloadsHandlerInvalidState(t *testing.T) {
	service := &fakeDownloadsService{err: ErrInvalidDownloadStateFilter}
	rec := doRequest(newDownloadsRouter(service), http.MethodGet, "/ingest/downloads?state=bogus", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListDownloadsHandlerInvalidPagination(t *testing.T) {
	service := &fakeDownloadsService{}
	rec := doRequest(newDownloadsRouter(service), http.MethodGet, "/ingest/downloads?page=abc", "")
	if rec.Code != http.StatusBadRequest || service.wasCalled {
		t.Fatalf("expected 400 without calling service, got %d", rec.Code)
	}
}

func TestListDownloadsHandlerServiceError(t *testing.T) {
	service := &fakeDownloadsService{err: errors.New("db down")}
	rec := doRequest(newDownloadsRouter(service), http.MethodGet, "/ingest/downloads", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

type fakeDownloadLister struct {
	downloads []DownloadDto
	err       error
}

func (f fakeDownloadLister) GetDownload(int) (DownloadDto, error) {
	return DownloadDto{}, f.err
}

func (f fakeDownloadLister) ListDownloads(string, int, int) ([]DownloadDto, error) {
	return f.downloads, f.err
}

func TestDownloadsServiceSetsHasNextAndTrimsExtraRow(t *testing.T) {
	service := NewDownloadsService(fakeDownloadLister{downloads: []DownloadDto{{ID: 3}, {ID: 2}, {ID: 1}}}, nil)
	response, err := service.ListDownloads("", 1, 2)
	if err != nil {
		t.Fatalf("ListDownloads: %v", err)
	}
	if len(response.Items) != 2 || !response.Pagination.HasNext || response.Pagination.HasPrev {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestDownloadsServiceNilRowsBecomeEmptyItems(t *testing.T) {
	service := NewDownloadsService(fakeDownloadLister{}, nil)
	response, err := service.ListDownloads("finished", 2, 5)
	if err != nil || response.Items == nil || response.Pagination.HasNext || !response.Pagination.HasPrev {
		t.Fatalf("unexpected response: %+v err=%v", response, err)
	}
}

func TestDownloadsServiceRejectsInvalidState(t *testing.T) {
	service := NewDownloadsService(fakeDownloadLister{}, nil)
	if _, err := service.ListDownloads("bogus", 1, 5); !errors.Is(err, ErrInvalidDownloadStateFilter) {
		t.Fatalf("err = %v, want ErrInvalidDownloadStateFilter", err)
	}
}

func TestRetryDownloadHandlerResponseFieldNames(t *testing.T) {
	jobID := 346
	createdAt := time.Date(2026, 11, 16, 20, 15, 0, 0, time.UTC)
	service := &fakeDownloadsService{retried: DownloadDto{
		ID: 13, JobID: &jobID, URL: "https://x.test/v", Preset: "video_1080",
		TargetRootID: 2, TargetRootName: "Videos", Subfolder: "Shows", State: "queued", CreatedAt: createdAt,
	}}
	rec := doRequest(newDownloadsRouter(service), http.MethodPost, "/ingest/downloads/12/retry", "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if service.retriedID != 12 {
		t.Fatalf("retried id = %d, want 12", service.retriedID)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, field := range []string{"id", "job_id", "url", "title", "preset", "target_root_id", "target_root_name",
		"subfolder", "state", "progress", "speed_bytes_per_second", "eta_seconds", "file_id", "error", "created_at", "ended_at"} {
		if _, exists := body[field]; !exists {
			t.Fatalf("field %q missing in %s", field, rec.Body.String())
		}
	}
	if body["id"] != float64(13) || body["state"] != "queued" || body["url"] != "https://x.test/v" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRetryDownloadHandlerErrorStatuses(t *testing.T) {
	testCases := []struct {
		name       string
		path       string
		err        error
		wantStatus int
	}{
		{"missing", "/ingest/downloads/12/retry", ErrDownloadNotFound, http.StatusNotFound},
		{"not retryable", "/ingest/downloads/12/retry", ErrDownloadNotRetryable, http.StatusConflict},
		{"destination gone", "/ingest/downloads/12/retry", ErrInvalidTarget, http.StatusBadRequest},
		{"jobs unavailable", "/ingest/downloads/12/retry", ErrJobsUnavailable, http.StatusServiceUnavailable},
		{"unexpected", "/ingest/downloads/12/retry", errors.New("db down"), http.StatusInternalServerError},
		{"invalid id", "/ingest/downloads/abc/retry", nil, http.StatusBadRequest},
		{"zero id", "/ingest/downloads/0/retry", nil, http.StatusBadRequest},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := &fakeDownloadsService{retryErr: testCase.err}
			rec := doRequest(newDownloadsRouter(service), http.MethodPost, testCase.path, "")
			if rec.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, testCase.wantStatus, rec.Body.String())
			}
		})
	}
}
