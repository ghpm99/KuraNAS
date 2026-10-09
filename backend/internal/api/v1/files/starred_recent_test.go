package files

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

type starredRecentServiceMock struct {
	filesHandlerServiceMock
	requestedStarred [][2]int
	requestedRecent  [][2]int
	starredPage      utils.PaginationResponse[FileDto]
	recentPage       utils.PaginationResponse[FileDto]
	listingError     error
}

func (m *starredRecentServiceMock) GetStarredFiles(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	m.requestedStarred = append(m.requestedStarred, [2]int{page, pageSize})
	return m.starredPage, m.listingError
}

func (m *starredRecentServiceMock) GetRecentlyAccessedFiles(page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	m.requestedRecent = append(m.requestedRecent, [2]int{page, pageSize})
	return m.recentPage, m.listingError
}

func newStarredRecentRouter(service *starredRecentServiceMock) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/starred", handler.GetStarredFilesHandler)
	router.GET("/files/recent-files", handler.GetRecentlyAccessedFilesHandler)
	return router
}

func TestStarredAndRecentHandlersPassDecodedPagination(t *testing.T) {
	service := &starredRecentServiceMock{
		starredPage: utils.PaginationResponse[FileDto]{Items: []FileDto{{ID: 4, Name: "fav", Path: "/tmp/fav"}}},
		recentPage:  utils.PaginationResponse[FileDto]{Items: []FileDto{{ID: 5, Name: "seen", Path: "/tmp/seen"}}},
	}
	router := newStarredRecentRouter(service)

	for _, route := range []string{"/files/starred?page=3&page_size=7", "/files/recent-files?page=3&page_size=7"} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d body=%s", route, recorder.Code, recorder.Body.String())
		}
		var body utils.PaginationResponse[FileDto]
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || len(body.Items) != 1 {
			t.Fatalf("%s: unexpected body %s err=%v", route, recorder.Body.String(), err)
		}
	}

	expectedPage := [2]int{3, 7}
	if len(service.requestedStarred) != 1 || service.requestedStarred[0] != expectedPage {
		t.Fatalf("starred received %v, expected %v", service.requestedStarred, expectedPage)
	}
	if len(service.requestedRecent) != 1 || service.requestedRecent[0] != expectedPage {
		t.Fatalf("recent received %v, expected %v", service.requestedRecent, expectedPage)
	}
}

func TestStarredAndRecentHandlersDefaultAndInvalidPagination(t *testing.T) {
	service := &starredRecentServiceMock{}
	router := newStarredRecentRouter(service)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/files/starred", nil))
	if recorder.Code != http.StatusOK || service.requestedStarred[0] != [2]int{1, 15} {
		t.Fatalf("expected default page 1/15, got code=%d %v", recorder.Code, service.requestedStarred)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/files/recent-files?page=abc", nil))
	if recorder.Code != http.StatusBadRequest || len(service.requestedRecent) != 0 {
		t.Fatalf("expected 400 without service call, got %d", recorder.Code)
	}
}

func TestStarredAndRecentHandlersMapServiceErrorToInternal(t *testing.T) {
	service := &starredRecentServiceMock{listingError: errors.New("db down")}
	router := newStarredRecentRouter(service)

	for _, route := range []string{"/files/starred", "/files/recent-files"} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("%s: expected 500, got %d", route, recorder.Code)
		}
	}
}

func TestServiceStarredAndRecentConvertPagesAndPropagateErrors(t *testing.T) {
	service := newFilesServiceForTest(t, &filesRepoMock{
		getStarredFilesFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(8, "fav", File)}}, nil
		},
		getRecentlyAccessedFilesFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("recent failed")
		},
	})

	starred, err := service.GetStarredFiles(1, 10)
	if err != nil || len(starred.Items) != 1 || starred.Items[0].ID != 8 {
		t.Fatalf("unexpected starred page %+v err=%v", starred, err)
	}
	if _, err := service.GetRecentlyAccessedFiles(1, 10); err == nil {
		t.Fatalf("expected recent error")
	}

	failing := newFilesServiceForTest(t, &filesRepoMock{
		getStarredFilesFn: func(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, errors.New("starred failed")
		},
	})
	if _, err := failing.GetStarredFiles(1, 10); err == nil {
		t.Fatalf("expected starred error")
	}
	if recent, err := failing.GetRecentlyAccessedFiles(1, 10); err != nil || len(recent.Items) != 0 {
		t.Fatalf("expected empty recent page, got %+v err=%v", recent, err)
	}
}

func TestRepositoryStarredAndRecentQueries(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetStarredFilesQuery)).
		WithArgs(11, 10).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 1, "a", "/tmp/a"))
	mock.ExpectRollback()
	starred, err := repo.GetStarredFiles(2, 10)
	if err != nil || len(starred.Items) != 1 {
		t.Fatalf("GetStarredFiles failed len=%d err=%v", len(starred.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetRecentlyAccessedFilesQuery)).
		WithArgs(11, 0).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 2, "b", "/tmp/b"))
	mock.ExpectRollback()
	recent, err := repo.GetRecentlyAccessedFiles(1, 10)
	if err != nil || len(recent.Items) != 1 {
		t.Fatalf("GetRecentlyAccessedFiles failed len=%d err=%v", len(recent.Items), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
