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
