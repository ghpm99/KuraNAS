package files

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func (m *filesRepoMock) SearchActiveFiles(query FileSearchQuery) (utils.PaginationResponse[FileModel], error) {
	if m.searchActiveFilesFn != nil {
		return m.searchActiveFilesFn(query)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}

type searchServiceMock struct {
	filesHandlerServiceMock
	receivedParams []FileSearchParams
	searchFailure  error
}

func (m *searchServiceMock) SearchFilesByName(params FileSearchParams) (utils.PaginationResponse[FileDto], error) {
	m.receivedParams = append(m.receivedParams, params)
	if m.searchFailure != nil {
		return utils.PaginationResponse[FileDto]{}, m.searchFailure
	}
	return m.listingPage(params.Page, params.PageSize)
}

func performSearchRequest(service *searchServiceMock, route string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/search", handler.SearchFilesHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
	return recorder
}

func TestSearchHandlerDecodesAllParams(t *testing.T) {
	service := &searchServiceMock{}
	recorder := performSearchRequest(service, "/files/search?q=%20relatorio%20&parent_id=7&recursive=false&page=3&page_size=40")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	expected := FileSearchParams{Query: "relatorio", ParentID: 7, IsRecursive: false, Page: 3, PageSize: 40, Filter: FileSearchFilter{Sort: SearchSortRelevance, Kinds: []FileSearchKind{}}}
	if len(service.receivedParams) != 1 || !reflect.DeepEqual(service.receivedParams[0], expected) {
		t.Fatalf("expected %+v, got %+v", expected, service.receivedParams)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid body: %v", err)
	}
	if _, hasItems := body["items"]; !hasItems {
		t.Fatalf("expected paginated shape, got %s", recorder.Body.String())
	}
	if _, hasPagination := body["pagination"]; !hasPagination {
		t.Fatalf("expected paginated shape, got %s", recorder.Body.String())
	}
}

func TestSearchHandlerDefaultsToGlobalRecursiveSearch(t *testing.T) {
	service := &searchServiceMock{}
	recorder := performSearchRequest(service, "/files/search?q=foto")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	expected := FileSearchParams{Query: "foto", ParentID: 0, IsRecursive: true, Page: 1, PageSize: 15, Filter: FileSearchFilter{Sort: SearchSortRelevance, Kinds: []FileSearchKind{}}}
	if !reflect.DeepEqual(service.receivedParams[0], expected) {
		t.Fatalf("expected %+v, got %+v", expected, service.receivedParams[0])
	}
}

func TestSearchHandlerRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name  string
		route string
	}{
		{name: "missing query", route: "/files/search"},
		{name: "one character query", route: "/files/search?q=a"},
		{name: "whitespace padded one character query", route: "/files/search?q=%20a%20"},
		{name: "invalid parent", route: "/files/search?q=ab&parent_id=abc"},
		{name: "negative parent", route: "/files/search?q=ab&parent_id=-1"},
		{name: "invalid recursive", route: "/files/search?q=ab&recursive=maybe"},
		{name: "invalid page", route: "/files/search?q=ab&page=x"},
		{name: "unknown kind", route: "/files/search?q=ab&kind=spreadsheet"},
		{name: "invalid modified_from", route: "/files/search?q=ab&modified_from=10-01-2026"},
		{name: "invalid modified_to", route: "/files/search?q=ab&modified_to=ontem"},
		{name: "inverted dates", route: "/files/search?q=ab&modified_from=2026-02-01&modified_to=2026-01-01"},
		{name: "invalid min_size", route: "/files/search?q=ab&min_size=big"},
		{name: "negative max_size", route: "/files/search?q=ab&max_size=-1"},
		{name: "inverted sizes", route: "/files/search?q=ab&min_size=10&max_size=5"},
		{name: "invalid tier", route: "/files/search?q=ab&tier=lukewarm"},
		{name: "invalid starred", route: "/files/search?q=ab&starred=yes"},
		{name: "invalid sort", route: "/files/search?q=ab&sort=color"},
		{name: "invalid order", route: "/files/search?q=ab&order=sideways"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := &searchServiceMock{}
			recorder := performSearchRequest(service, tc.route)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d body=%s", recorder.Code, recorder.Body.String())
			}
			if len(service.receivedParams) != 0 {
				t.Fatalf("service must not be called, got %+v", service.receivedParams)
			}
			var body map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] == "" {
				t.Fatalf("expected error body, got %s", recorder.Body.String())
			}
		})
	}
}

func TestSearchHandlerAcceptsTwoMultibyteCharacters(t *testing.T) {
	service := &searchServiceMock{}
	recorder := performSearchRequest(service, "/files/search?q=%C3%A7%C3%A3")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestSearchHandlerMapsServiceFailures(t *testing.T) {
	tests := []struct {
		name    string
		failure error
		code    int
	}{
		{name: "parent not found", failure: ErrFileNotFound, code: http.StatusNotFound},
		{name: "internal", failure: errors.New("boom"), code: http.StatusInternalServerError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := performSearchRequest(&searchServiceMock{searchFailure: tc.failure}, "/files/search?q=ab&parent_id=1")
			if recorder.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, recorder.Code)
			}
			var body map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] == "" {
				t.Fatalf("expected error body, got %s", recorder.Body.String())
			}
		})
	}
}

func TestBuildDescendantsPathPrefixAppendsSeparatorOnce(t *testing.T) {
	separator := string(filepath.Separator)
	if prefix := buildDescendantsPathPrefix(separator + "srv"); prefix != separator+"srv"+separator {
		t.Fatalf("unexpected prefix %q", prefix)
	}
	if prefix := buildDescendantsPathPrefix(separator); prefix != separator {
		t.Fatalf("unexpected root prefix %q", prefix)
	}
}

func searchFolderModel(id int, path string, fileType FileType) FileModel {
	model := sampleModel(id, filepath.Base(path), fileType)
	model.Path = path
	return model
}

func TestSearchHandlerDecodesFilterParams(t *testing.T) {
	service := &searchServiceMock{}
	route := "/files/search?q=ab&kind=image&kind=folder&modified_from=2026-01-01&modified_to=2026-01-31&min_size=10&max_size=20&tier=cold&starred=true&sort=size&order=asc"
	recorder := performSearchRequest(service, route)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	modifiedFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedTo := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	minSize, maxSize := int64(10), int64(20)
	expected := FileSearchFilter{
		Kinds:        []FileSearchKind{SearchKindImage, SearchKindFolder},
		ModifiedFrom: &modifiedFrom,
		ModifiedTo:   &modifiedTo,
		MinSize:      &minSize,
		MaxSize:      &maxSize,
		Tier:         TierCold,
		OnlyStarred:  true,
		Sort:         SearchSortSize,
		Order:        SearchOrderAscending,
	}
	if !reflect.DeepEqual(service.receivedParams[0].Filter, expected) {
		t.Fatalf("expected %+v, got %+v", expected, service.receivedParams[0].Filter)
	}
}

func TestServiceSearchFilesByNameChoosesScope(t *testing.T) {
	separator := string(filepath.Separator)
	folderPath := separator + "srv" + separator + "docs"
	var receivedQueries []FileSearchQuery
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			return searchFolderModel(id, folderPath, Directory), true, nil
		},
		searchActiveFilesFn: func(query FileSearchQuery) (utils.PaginationResponse[FileModel], error) {
			receivedQueries = append(receivedQueries, query)
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(1, "a", File)}}, nil
		},
	}
	service := newFilesServiceForTest(t, repo)

	paramsList := []FileSearchParams{
		{Query: "x_", Page: 1, PageSize: 10},
		{Query: "x_", ParentID: 5, IsRecursive: true, Page: 1, PageSize: 10},
		{Query: "x_", ParentID: 5, IsRecursive: false, Page: 1, PageSize: 10},
	}
	for _, params := range paramsList {
		if _, err := service.SearchFilesByName(params); err != nil {
			t.Fatalf("search %+v: %v", params, err)
		}
	}

	expected := []FileSearchQuery{
		{Query: "x_", Scope: SearchScopeGlobal, Page: 1, PageSize: 10},
		{Query: "x_", Scope: SearchScopeDescendants, ScopePath: folderPath + separator, Page: 1, PageSize: 10},
		{Query: "x_", Scope: SearchScopeChildren, ScopePath: folderPath, Page: 1, PageSize: 10},
	}
	if !reflect.DeepEqual(receivedQueries, expected) {
		t.Fatalf("expected %+v, got %+v", expected, receivedQueries)
	}
}

func TestServiceSearchFilesByNameRejectsUnusableParent(t *testing.T) {
	deletedFolder := searchFolderModel(5, "/srv/gone", Directory)
	deletedFolder.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}
	tests := []struct {
		name  string
		model FileModel
		found bool
		err   error
		want  error
	}{
		{name: "missing", found: false, want: ErrFileNotFound},
		{name: "deleted", model: deletedFolder, found: true, want: ErrFileNotFound},
		{name: "not a directory", model: searchFolderModel(5, "/srv/a.txt", File), found: true, want: ErrFileNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := newFilesServiceForTest(t, &filesRepoMock{
				getFileByIDFn: func(id int) (FileModel, bool, error) { return tc.model, tc.found, tc.err },
			})
			_, err := service.SearchFilesByName(FileSearchParams{Query: "ab", ParentID: 5, IsRecursive: true, Page: 1, PageSize: 10})
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestServiceSearchFilesByNamePropagatesRepositoryFailures(t *testing.T) {
	failure := errors.New("db down")
	service := newFilesServiceForTest(t, &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) { return FileModel{}, false, failure },
		searchActiveFilesFn: func(FileSearchQuery) (utils.PaginationResponse[FileModel], error) {
			return utils.PaginationResponse[FileModel]{}, failure
		},
	})
	if _, err := service.SearchFilesByName(FileSearchParams{Query: "ab", Page: 1, PageSize: 10}); !errors.Is(err, failure) {
		t.Fatalf("global search must propagate failure, got %v", err)
	}
	if _, err := service.SearchFilesByName(FileSearchParams{Query: "ab", ParentID: 1, Page: 1, PageSize: 10}); !errors.Is(err, failure) {
		t.Fatalf("parent lookup must propagate failure, got %v", err)
	}
}

func TestRepositorySearchActiveFilesBindsBuiltStatementAndPagination(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	searchQuery := FileSearchQuery{Query: "ab", Page: 2, PageSize: 10}
	statement, _, err := buildSearchFilesQuery(searchQuery)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(statement)).
		WithArgs("%ab%", sqlmock.AnyArg(), "ab", "ab%", "%ab%", 11, 10).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 1, "ab", "/tmp/ab"))
	mock.ExpectRollback()
	page, err := repo.SearchActiveFiles(searchQuery)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("search: items=%v err=%v", page.Items, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(statement)).WillReturnError(errors.New("search failed"))
	mock.ExpectRollback()
	if _, err := repo.SearchActiveFiles(searchQuery); err == nil {
		t.Fatalf("expected search error")
	}

	if _, err := repo.SearchActiveFiles(FileSearchQuery{Query: "   ", Page: 1, PageSize: 10}); err == nil {
		t.Fatalf("expected an error for a query without terms")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
