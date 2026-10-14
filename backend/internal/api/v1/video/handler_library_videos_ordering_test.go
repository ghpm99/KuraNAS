package video

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type libraryVideosCapturingService struct {
	videoHandlerServiceMock
	capturedRequest LibraryVideosRequest
}

func (service *libraryVideosCapturingService) ListLibraryVideos(request LibraryVideosRequest) (utils.PaginationResponse[VideoFileDto], error) {
	service.capturedRequest = request
	return utils.PaginationResponse[VideoFileDto]{}, nil
}

func TestListLibraryVideosHandlerDecodesSortAndOrderIntoServiceRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &libraryVideosCapturingService{}
	router := gin.New()
	router.GET("/video/library/files", (&Handler{service: service}).ListLibraryVideosHandler)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/video/library/files?sort=size&order=asc&query=abc", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", recorder.Code)
	}
	expectedOrdering := LibraryVideoOrdering{Sort: LibraryVideoSortSize, Direction: SortDirectionAscending}
	if service.capturedRequest.Ordering != expectedOrdering || service.capturedRequest.SearchQuery != "abc" {
		t.Fatalf("unexpected request: %+v", service.capturedRequest)
	}
}

func TestListLibraryVideosHandlerDefaultsToRecentDescending(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &libraryVideosCapturingService{}
	router := gin.New()
	router.GET("/video/library/files", (&Handler{service: service}).ListLibraryVideosHandler)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/video/library/files", nil))

	expectedOrdering := LibraryVideoOrdering{Sort: LibraryVideoSortRecent, Direction: SortDirectionDescending}
	if service.capturedRequest.Ordering != expectedOrdering {
		t.Fatalf("unexpected ordering: %+v", service.capturedRequest.Ordering)
	}
}

func TestListLibraryVideosHandlerPassesFiltersToServiceRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &libraryVideosCapturingService{}
	router := gin.New()
	router.GET("/video/library/files", (&Handler{service: service}).ListLibraryVideosHandler)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/video/library/files?format=MP4,.mkv&resolution=fhd&duration=long", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", recorder.Code)
	}
	filters := service.capturedRequest.Filters
	if len(filters.Formats) != 2 || filters.Formats[0] != ".mp4" || filters.Formats[1] != ".mkv" {
		t.Fatalf("unexpected formats: %v", filters.Formats)
	}
	if filters.Resolution != LibraryVideoResolutionFHD || filters.Duration != LibraryVideoDurationLong {
		t.Fatalf("unexpected filters: %+v", filters)
	}
}

func TestListLibraryVideosHandlerRejectsInvalidFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, rawQuery := range []string{"format=exe", "resolution=8k", "duration=eternal"} {
		service := &libraryVideosCapturingService{}
		router := gin.New()
		router.GET("/video/library/files", (&Handler{service: service}).ListLibraryVideosHandler)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/video/library/files?"+rawQuery, nil))

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: want 400, got %d", rawQuery, recorder.Code)
		}
	}
}
