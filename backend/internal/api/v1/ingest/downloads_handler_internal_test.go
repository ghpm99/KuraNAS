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

func newDownloadsRouter(service DownloadsServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ingest/downloads", NewDownloadsHandler(service).ListDownloadsHandler)
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

func (f fakeDownloadLister) ListDownloads(string, int, int) ([]DownloadDto, error) {
	return f.downloads, f.err
}

func TestDownloadsServiceSetsHasNextAndTrimsExtraRow(t *testing.T) {
	service := NewDownloadsService(fakeDownloadLister{downloads: []DownloadDto{{ID: 3}, {ID: 2}, {ID: 1}}})
	response, err := service.ListDownloads("", 1, 2)
	if err != nil {
		t.Fatalf("ListDownloads: %v", err)
	}
	if len(response.Items) != 2 || !response.Pagination.HasNext || response.Pagination.HasPrev {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestDownloadsServiceNilRowsBecomeEmptyItems(t *testing.T) {
	service := NewDownloadsService(fakeDownloadLister{})
	response, err := service.ListDownloads("finished", 2, 5)
	if err != nil || response.Items == nil || response.Pagination.HasNext || !response.Pagination.HasPrev {
		t.Fatalf("unexpected response: %+v err=%v", response, err)
	}
}

func TestDownloadsServiceRejectsInvalidState(t *testing.T) {
	service := NewDownloadsService(fakeDownloadLister{})
	if _, err := service.ListDownloads("bogus", 1, 5); !errors.Is(err, ErrInvalidDownloadStateFilter) {
		t.Fatalf("err = %v, want ErrInvalidDownloadStateFilter", err)
	}
}
