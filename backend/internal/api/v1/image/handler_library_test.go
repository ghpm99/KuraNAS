package image

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type fakeLibraryService struct {
	err           error
	page          LibraryPageDto
	count         LibraryCountDto
	buckets       []LibraryTimelineBucketDto
	listRequest   LibraryListRequest
	filterGotten  LibraryFilter
	callCount     int
	folderPage    utils.PaginationResponse[LibraryFolderDto]
	folderRequest LibraryFolderRequest
	neighbors     LibraryNeighborsDto
	neighborsReq  LibraryNeighborsRequest
}

func (f *fakeLibraryService) ListLibraryNeighbors(request LibraryNeighborsRequest) (LibraryNeighborsDto, error) {
	f.callCount++
	f.neighborsReq = request
	return f.neighbors, f.err
}

func (f *fakeLibraryService) ListLibraryImages(request LibraryListRequest) (LibraryPageDto, error) {
	f.callCount++
	f.listRequest = request
	return f.page, f.err
}

func (f *fakeLibraryService) CountLibraryImages(filter LibraryFilter) (LibraryCountDto, error) {
	f.callCount++
	f.filterGotten = filter
	return f.count, f.err
}

func (f *fakeLibraryService) ListLibraryTimeline(filter LibraryFilter) ([]LibraryTimelineBucketDto, error) {
	f.callCount++
	f.filterGotten = filter
	return f.buckets, f.err
}

func (f *fakeLibraryService) ListLibraryFolders(request LibraryFolderRequest) (utils.PaginationResponse[LibraryFolderDto], error) {
	f.callCount++
	f.folderRequest = request
	return f.folderPage, f.err
}

func serveLibraryRequest(service *fakeLibraryService, requestURL string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewLibraryHandler(service, &imageLoggerMock{})
	router := gin.New()
	router.GET("/image/library", handler.ListLibraryImagesHandler)
	router.GET("/image/library/neighbors/:file_id", handler.ListLibraryNeighborsHandler)
	router.GET("/image/library/count", handler.CountLibraryImagesHandler)
	router.GET("/image/library/folders", handler.ListLibraryFoldersHandler)
	router.GET("/image/library/timeline", handler.ListLibraryTimelineHandler)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, requestURL, nil))
	return recorder
}

func TestListLibraryImagesHandlerDecodesEveryParameter(t *testing.T) {
	service := &fakeLibraryService{page: LibraryPageDto{Items: []LibraryItemDto{{FileID: 1}}, NextCursor: "abc", HasNext: true, PageSize: 25}}
	cursor := LibraryCursor{FileID: 11}.Encode()
	query := url.Values{
		"q":            {"  Beach "},
		"category":     {"photo", "Portrait"},
		"starred":      {"true"},
		"format":       {"jpg", ".PNG"},
		"taken_from":   {"2020-01-02"},
		"taken_to":     {"2021-03-04"},
		"camera":       {"Canon EOS 5D"},
		"folder":       {"/photos"},
		"sort":         {"taken_at"},
		"order":        {"desc"},
		"cursor":       {cursor},
		"taken_before": {"2019-06-07T08:09:10Z"},
		"page_size":    {"25"},
	}

	recorder := serveLibraryRequest(service, "/image/library?"+query.Encode())

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", recorder.Code, recorder.Body.String())
	}
	request := service.listRequest
	filter := request.Filter
	if filter.NameQuery != "beach" || !filter.OnlyStarred || filter.Camera != "Canon EOS 5D" || filter.Folder != "/photos" {
		t.Fatalf("unexpected filter %+v", filter)
	}
	if len(filter.Categories) != 2 || filter.Categories[0] != ClassificationCategoryPhoto || filter.Categories[1] != ClassificationCategoryPortrait {
		t.Fatalf("unexpected categories %v", filter.Categories)
	}
	if len(filter.Formats) != 2 || filter.Formats[0] != ".jpg" || filter.Formats[1] != ".png" {
		t.Fatalf("unexpected formats %v", filter.Formats)
	}
	if !filter.TakenFrom.Equal(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected taken_from %v", filter.TakenFrom)
	}
	if !filter.TakenTo.Equal(time.Date(2021, 3, 4, 23, 59, 59, 999999000, time.UTC)) {
		t.Fatalf("taken_to date-only must cover the whole day, got %v", filter.TakenTo)
	}
	if request.Cursor == nil || request.Cursor.FileID != 11 || request.TakenBefore == nil || !request.TakenBefore.Equal(time.Date(2019, 6, 7, 8, 9, 10, 0, time.UTC)) {
		t.Fatalf("unexpected cursor/taken_before %+v", request)
	}
	if request.PageSize != 25 || request.Page != 1 || request.Sort != LibrarySortTakenAt || request.Order != LibrarySortOrderDesc {
		t.Fatalf("unexpected paging %+v", request)
	}

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"items", "next_cursor", "has_next", "page_size"} {
		if _, hasKey := body[key]; !hasKey {
			t.Fatalf("response missing %q: %s", key, recorder.Body.String())
		}
	}
}

func TestListLibraryImagesHandlerDefaults(t *testing.T) {
	service := &fakeLibraryService{}
	recorder := serveLibraryRequest(service, "/image/library")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	request := service.listRequest
	if request.Sort != LibrarySortTakenAt || request.Order != LibrarySortOrderDesc || request.PageSize != defaultLibraryPageSize || request.Cursor != nil {
		t.Fatalf("unexpected defaults %+v", request)
	}
}

func TestListLibraryImagesHandlerDefaultOrderPerSort(t *testing.T) {
	service := &fakeLibraryService{}
	serveLibraryRequest(service, "/image/library?sort=name&page=2")
	if service.listRequest.Order != LibrarySortOrderAsc || service.listRequest.Page != 2 {
		t.Fatalf("unexpected request %+v", service.listRequest)
	}
	serveLibraryRequest(service, "/image/library?sort=size")
	if service.listRequest.Order != LibrarySortOrderDesc {
		t.Fatalf("unexpected request %+v", service.listRequest)
	}
}

func TestLibraryHandlersRejectInvalidParameters(t *testing.T) {
	invalidQueries := map[string]string{
		"cursor":         "cursor=%21%21",
		"category":       "category=nope",
		"format":         "format=exe",
		"taken_from":     "taken_from=yesterday",
		"taken_to":       "taken_to=2021-13-45",
		"taken_before":   "taken_before=soon",
		"sort":           "sort=color",
		"order":          "order=sideways",
		"starred":        "starred=maybe",
		"query too long": "q=" + url.QueryEscape(repeatLetter('a', maxLibraryNameQueryLength+1)),
		"cursor w/ name": "sort=name&cursor=" + LibraryCursor{FileID: 1}.Encode(),
		"seek w/ size":   "sort=size&taken_before=2020-01-01",
		"cursor w/ asc":  "order=asc&cursor=" + LibraryCursor{FileID: 1}.Encode(),
		"page":           "page=abc",
	}

	for name, rawQuery := range invalidQueries {
		for _, route := range []string{"/image/library", "/image/library/count", "/image/library/timeline"} {
			if route != "/image/library" && !libraryFilterParameterNames[name] {
				continue
			}
			service := &fakeLibraryService{}
			recorder := serveLibraryRequest(service, route+"?"+rawQuery)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("%s %s: status = %d body %s", route, name, recorder.Code, recorder.Body.String())
			}
			if service.callCount != 0 {
				t.Fatalf("%s %s: service must not be called", route, name)
			}
			var body map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] == "" {
				t.Fatalf("%s %s: expected translated error body, got %s", route, name, recorder.Body.String())
			}
		}
	}
}

var libraryFilterParameterNames = map[string]bool{
	"category": true, "format": true, "taken_from": true, "taken_to": true, "starred": true, "query too long": true,
}

func repeatLetter(letter byte, count int) string {
	letters := make([]byte, count)
	for position := range letters {
		letters[position] = letter
	}
	return string(letters)
}

func TestCountLibraryImagesHandler(t *testing.T) {
	service := &fakeLibraryService{count: LibraryCountDto{Total: 77}}
	recorder := serveLibraryRequest(service, "/image/library/count?category=photo&starred=true")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var body map[string]int
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["total"] != 77 {
		t.Fatalf("body = %s err %v", recorder.Body.String(), err)
	}
	if !service.filterGotten.OnlyStarred || len(service.filterGotten.Categories) != 1 {
		t.Fatalf("unexpected filter %+v", service.filterGotten)
	}
}

func TestListLibraryTimelineHandler(t *testing.T) {
	service := &fakeLibraryService{buckets: []LibraryTimelineBucketDto{{Year: 2022, Month: 5, Count: 3}}}
	recorder := serveLibraryRequest(service, "/image/library/timeline?camera=Canon")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var body []map[string]int
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || len(body) != 1 || body[0]["year"] != 2022 || body[0]["month"] != 5 || body[0]["count"] != 3 {
		t.Fatalf("body = %s err %v", recorder.Body.String(), err)
	}
	if service.filterGotten.Camera != "Canon" {
		t.Fatalf("unexpected filter %+v", service.filterGotten)
	}
}

func TestLibraryHandlersReturn500OnServiceError(t *testing.T) {
	for _, route := range []string{"/image/library", "/image/library/count", "/image/library/timeline"} {
		recorder := serveLibraryRequest(&fakeLibraryService{err: errors.New("db down")}, route)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("%s: status = %d", route, recorder.Code)
		}
	}
}

func TestListLibraryFoldersHandlerDecodesParentAndPagination(t *testing.T) {
	service := &fakeLibraryService{folderPage: utils.PaginationResponse[LibraryFolderDto]{
		Items:      []LibraryFolderDto{{Path: "/photos/trip", Name: "trip", ImageCount: 4, CoverFileID: 9}},
		Pagination: utils.Pagination{Page: 2, PageSize: 10, HasNext: true},
	}}
	query := url.Values{"parent": {"  /photos "}, "page": {"2"}, "page_size": {"10"}}

	recorder := serveLibraryRequest(service, "/image/library/folders?"+query.Encode())

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", recorder.Code, recorder.Body.String())
	}
	if service.folderRequest.ParentPath != "/photos" || service.folderRequest.Page != 2 || service.folderRequest.PageSize != 10 {
		t.Fatalf("unexpected request %+v", service.folderRequest)
	}
	var body struct {
		Items      []map[string]any `json:"items"`
		Pagination map[string]any   `json:"pagination"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"path", "name", "image_count", "cover_file_id"} {
		if _, hasKey := body.Items[0][key]; !hasKey {
			t.Fatalf("folder missing %q: %s", key, recorder.Body.String())
		}
	}
	if body.Pagination["has_next"] != true {
		t.Fatalf("pagination: %s", recorder.Body.String())
	}
}

func TestListLibraryFoldersHandlerDefaultsToRoots(t *testing.T) {
	service := &fakeLibraryService{}
	recorder := serveLibraryRequest(service, "/image/library/folders")

	if recorder.Code != http.StatusOK || service.folderRequest.ParentPath != "" || service.folderRequest.PageSize != defaultLibraryFolderPageSize {
		t.Fatalf("status %d request %+v", recorder.Code, service.folderRequest)
	}
}

func TestListLibraryFoldersHandlerRejectsInvalidInput(t *testing.T) {
	tooLong := url.Values{"parent": {strings.Repeat("a", maxLibraryFolderPathLength+1)}}
	for name, requestURL := range map[string]string{
		"parent too long": "/image/library/folders?" + tooLong.Encode(),
		"parent with NUL": "/image/library/folders?parent=%2Fa%00b",
		"bad pagination":  "/image/library/folders?page=abc",
	} {
		service := &fakeLibraryService{}
		recorder := serveLibraryRequest(service, requestURL)
		if recorder.Code != http.StatusBadRequest || service.callCount != 0 {
			t.Fatalf("%s: status %d calls %d", name, recorder.Code, service.callCount)
		}
	}
}

func TestListLibraryFoldersHandlerMapsServiceErrorToInternal(t *testing.T) {
	service := &fakeLibraryService{err: errors.New("boom")}
	recorder := serveLibraryRequest(service, "/image/library/folders")
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestListLibraryNeighborsHandlerDecodesFileFilterAndCount(t *testing.T) {
	service := &fakeLibraryService{neighbors: LibraryNeighborsDto{
		Before: []LibraryItemDto{{FileID: 8}},
		After:  []LibraryItemDto{{FileID: 6}},
	}}

	recorder := serveLibraryRequest(service, "/image/library/neighbors/7?starred=true&folder=/photos&count=5")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", recorder.Code, recorder.Body.String())
	}
	request := service.neighborsReq
	if request.FileID != 7 || request.Count != 5 || !request.Filter.OnlyStarred || request.Filter.Folder != "/photos" {
		t.Fatalf("unexpected request %+v", request)
	}
	var body map[string][]LibraryItemDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body["before"]) != 1 || body["before"][0].FileID != 8 || len(body["after"]) != 1 || body["after"][0].FileID != 6 {
		t.Fatalf("response must expose before and after, got %s", recorder.Body.String())
	}
}

func TestListLibraryNeighborsHandlerDefaultsCount(t *testing.T) {
	service := &fakeLibraryService{}
	recorder := serveLibraryRequest(service, "/image/library/neighbors/7")

	if recorder.Code != http.StatusOK || service.neighborsReq.Count != defaultLibraryNeighborCount {
		t.Fatalf("status = %d request %+v", recorder.Code, service.neighborsReq)
	}
}

func TestListLibraryNeighborsHandlerRejectsInvalidInput(t *testing.T) {
	for _, requestURL := range []string{
		"/image/library/neighbors/abc",
		"/image/library/neighbors/0",
		"/image/library/neighbors/7?count=0",
		"/image/library/neighbors/7?count=51",
		"/image/library/neighbors/7?count=x",
		"/image/library/neighbors/7?category=nope",
	} {
		service := &fakeLibraryService{}
		recorder := serveLibraryRequest(service, requestURL)
		if recorder.Code != http.StatusBadRequest || service.callCount != 0 {
			t.Fatalf("%s: status = %d calls = %d", requestURL, recorder.Code, service.callCount)
		}
	}
}

func TestListLibraryNeighborsHandlerMapsServiceErrors(t *testing.T) {
	notFound := serveLibraryRequest(&fakeLibraryService{err: fmt.Errorf("wrapped: %w", sql.ErrNoRows)}, "/image/library/neighbors/7")
	if notFound.Code != http.StatusNotFound {
		t.Fatalf("missing image status = %d", notFound.Code)
	}

	internal := serveLibraryRequest(&fakeLibraryService{err: errors.New("db down")}, "/image/library/neighbors/7")
	if internal.Code != http.StatusInternalServerError {
		t.Fatalf("internal status = %d", internal.Code)
	}
}
