package files

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func (m *filesRepoMock) SearchActiveFilesByName(namePattern string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.searchByNameFn != nil {
		return m.searchByNameFn(namePattern, page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}

func (m *filesRepoMock) SearchActiveFilesByNameUnderPath(pathPrefix string, namePattern string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.searchUnderPathFn != nil {
		return m.searchUnderPathFn(pathPrefix, namePattern, page, pageSize)
	}
	return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
}

func (m *filesRepoMock) SearchActiveChildrenByName(parentPath string, namePattern string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	if m.searchChildrenFn != nil {
		return m.searchChildrenFn(parentPath, namePattern, page, pageSize)
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
	expected := FileSearchParams{Query: "relatorio", ParentID: 7, IsRecursive: false, Page: 3, PageSize: 40}
	if len(service.receivedParams) != 1 || service.receivedParams[0] != expected {
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
	expected := FileSearchParams{Query: "foto", ParentID: 0, IsRecursive: true, Page: 1, PageSize: 15}
	if service.receivedParams[0] != expected {
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

func TestBuildContainsNamePatternEscapesLikeWildcards(t *testing.T) {
	tests := map[string]string{
		"relatorio": "%relatorio%",
		"100%":      `%100\%%`,
		"a_b":       `%a\_b%`,
		`c\d`:       `%c\\d%`,
	}
	for query, expected := range tests {
		if pattern := buildContainsNamePattern(query); pattern != expected {
			t.Fatalf("query %q: expected %q, got %q", query, expected, pattern)
		}
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

func TestServiceSearchFilesByNameChoosesQueryVariant(t *testing.T) {
	separator := string(filepath.Separator)
	folderPath := separator + "srv" + separator + "docs"
	var calledVariants []string
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			return searchFolderModel(id, folderPath, Directory), true, nil
		},
		searchByNameFn: func(namePattern string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			calledVariants = append(calledVariants, "global:"+namePattern)
			return utils.PaginationResponse[FileModel]{Items: []FileModel{sampleModel(1, "a", File)}}, nil
		},
		searchUnderPathFn: func(pathPrefix string, namePattern string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			calledVariants = append(calledVariants, "under:"+pathPrefix+"|"+namePattern)
			return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
		},
		searchChildrenFn: func(parentPath string, namePattern string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			calledVariants = append(calledVariants, "children:"+parentPath+"|"+namePattern)
			return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
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

	expected := []string{
		`global:%x\_%`,
		"under:" + folderPath + separator + `|%x\_%`,
		"children:" + folderPath + `|%x\_%`,
	}
	if len(calledVariants) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, calledVariants)
	}
	for index := range expected {
		if calledVariants[index] != expected[index] {
			t.Fatalf("expected %v, got %v", expected, calledVariants)
		}
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
		searchByNameFn: func(string, int, int) (utils.PaginationResponse[FileModel], error) {
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

func TestRepositorySearchQueriesBindPatternAndPagination(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchActiveFilesByNameQuery)).
		WithArgs("%ab%", 11, 10).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 1, "ab", "/tmp/ab"))
	mock.ExpectRollback()
	globalPage, err := repo.SearchActiveFilesByName("%ab%", 2, 10)
	if err != nil || len(globalPage.Items) != 1 {
		t.Fatalf("global search: items=%v err=%v", globalPage.Items, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchActiveFilesByNameUnderPathQuery)).
		WithArgs("/tmp/", "%ab%", 11, 0).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 1, "ab", "/tmp/ab"))
	mock.ExpectRollback()
	if underPage, err := repo.SearchActiveFilesByNameUnderPath("/tmp/", "%ab%", 1, 10); err != nil || len(underPage.Items) != 1 {
		t.Fatalf("under path search: items=%v err=%v", underPage.Items, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchActiveChildrenByNameQuery)).
		WithArgs("/tmp", "%ab%", 11, 0).
		WillReturnError(errors.New("children failed"))
	mock.ExpectRollback()
	if _, err := repo.SearchActiveChildrenByName("/tmp", "%ab%", 1, 10); err == nil {
		t.Fatalf("expected children search error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
