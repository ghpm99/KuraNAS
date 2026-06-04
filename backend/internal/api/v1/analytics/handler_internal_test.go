package analytics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type serviceStub struct {
	err            error
	capturedPeriod string
	capturedLimit  int
}

func (s *serviceStub) GetStorage(period string) (StorageStatsDto, error) {
	s.capturedPeriod = period
	return StorageStatsDto{}, s.err
}
func (s *serviceStub) GetTimeSeries(period string) ([]TimeSeriesPointDto, error) {
	s.capturedPeriod = period
	return nil, s.err
}
func (s *serviceStub) GetTypes() ([]TypeBreakdownDto, error) { return nil, s.err }
func (s *serviceStub) GetExtensions(limit int) ([]ExtensionDto, error) {
	s.capturedLimit = limit
	return nil, s.err
}
func (s *serviceStub) GetRecentFiles(limit int) ([]RecentFileDto, error) {
	s.capturedLimit = limit
	return nil, s.err
}
func (s *serviceStub) GetTopFolders(limit int) ([]FolderUsageDto, error) {
	s.capturedLimit = limit
	return nil, s.err
}
func (s *serviceStub) GetHotFolders(period string, limit int) ([]HotFolderDto, error) {
	s.capturedPeriod = period
	s.capturedLimit = limit
	return nil, s.err
}
func (s *serviceStub) GetDuplicatesSummary() (DuplicatesSummaryDto, error) {
	return DuplicatesSummaryDto{}, s.err
}
func (s *serviceStub) GetDuplicateGroups(limit int) ([]DuplicateGroupDto, error) {
	s.capturedLimit = limit
	return nil, s.err
}
func (s *serviceStub) GetLibrary() (LibraryDto, error)       { return LibraryDto{}, s.err }
func (s *serviceStub) GetProcessing() (ProcessingDto, error) { return ProcessingDto{}, s.err }
func (s *serviceStub) GetHealth() (HealthDto, error)         { return HealthDto{}, s.err }
func (s *serviceStub) GetInsights(period string) ([]string, error) {
	s.capturedPeriod = period
	return []string{}, s.err
}

func doRequest(t *testing.T, register func(*gin.Engine, *Handler), url string) (*serviceStub, *httptest.ResponseRecorder) {
	t.Helper()
	stub := &serviceStub{}
	handler := NewHandler(stub)
	router := gin.New()
	register(router, handler)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return stub, w
}

func TestHandlerStorageSuccessAndPeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub, w := doRequest(t, func(r *gin.Engine, h *Handler) {
		r.GET("/analytics/storage", h.GetStorageHandler)
	}, "/analytics/storage?period=24h")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if stub.capturedPeriod != "24h" {
		t.Fatalf("expected period 24h, got %s", stub.capturedPeriod)
	}
}

func TestHandlerStorageDefaultPeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub, w := doRequest(t, func(r *gin.Engine, h *Handler) {
		r.GET("/analytics/storage", h.GetStorageHandler)
	}, "/analytics/storage")
	if w.Code != http.StatusOK || stub.capturedPeriod != "7d" {
		t.Fatalf("expected default period 7d and 200, got %s / %d", stub.capturedPeriod, w.Code)
	}
}

func TestHandlerInvalidPeriodReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &serviceStub{err: ErrInvalidPeriod}
	handler := NewHandler(stub)
	router := gin.New()
	router.GET("/analytics/storage", handler.GetStorageHandler)
	req := httptest.NewRequest(http.MethodGet, "/analytics/storage?period=bad", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandlerInternalErrorReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &serviceStub{err: errors.New("db")}
	handler := NewHandler(stub)
	router := gin.New()
	router.GET("/analytics/types", handler.GetTypesHandler)
	req := httptest.NewRequest(http.MethodGet, "/analytics/types", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandlerLimitParsing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		query string
		want  int
	}{
		{"", 12},            // default
		{"?limit=5", 5},     // valid
		{"?limit=0", 12},    // non-positive -> default
		{"?limit=abc", 12},  // invalid -> default
		{"?limit=999", 100}, // above max -> capped
	}
	for _, tc := range cases {
		stub, w := doRequest(t, func(r *gin.Engine, h *Handler) {
			r.GET("/analytics/extensions", h.GetExtensionsHandler)
		}, "/analytics/extensions"+tc.query)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for %q, got %d", tc.query, w.Code)
		}
		if stub.capturedLimit != tc.want {
			t.Fatalf("query %q: expected limit %d, got %d", tc.query, tc.want, stub.capturedLimit)
		}
	}
}

func TestHandlerRemainingEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	endpoints := []struct {
		register func(*gin.Engine, *Handler)
		url      string
	}{
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetTimeSeriesHandler) }, "/x?period=30d"},
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetRecentFilesHandler) }, "/x"},
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetTopFoldersHandler) }, "/x"},
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetHotFoldersHandler) }, "/x?period=90d&limit=4"},
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetDuplicatesHandler) }, "/x"},
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetDuplicateGroupsHandler) }, "/x"},
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetLibraryHandler) }, "/x"},
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetProcessingHandler) }, "/x"},
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetHealthHandler) }, "/x"},
		{func(r *gin.Engine, h *Handler) { r.GET("/x", h.GetInsightsHandler) }, "/x"},
	}
	for _, ep := range endpoints {
		_, w := doRequest(t, ep.register, ep.url)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", ep.url, w.Code)
		}
	}
}
