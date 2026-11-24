package documenttext

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func performListRequest(handler *Handler, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/documents", handler.ListDocumentsHandler)
	router.GET("/documents/count", handler.CountDocumentsHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func TestListDocumentsHandlerFixesResponseFieldNames(t *testing.T) {
	moment := time.Date(2024, 3, 5, 10, 0, 0, 0, time.UTC)
	service := &stubService{documentPage: utils.PaginationResponse[DocumentItemDto]{
		Items:      []DocumentItemDto{{ID: 7, Name: "a.pdf", Format: ".pdf", ParentPath: "/Casa", Size: 10, ModifiedAt: moment, CreatedAt: moment, Starred: true, IndexState: StatePartial}},
		Pagination: utils.Pagination{Page: 1, PageSize: 50, HasNext: true},
	}}

	recorder := performListRequest(NewHandler(service), "/documents")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Items      []map[string]any `json:"items"`
		Pagination map[string]any   `json:"pagination"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	item := payload.Items[0]
	if len(item) != 9 || item["id"] != float64(7) || item["name"] != "a.pdf" || item["format"] != ".pdf" || item["parent_path"] != "/Casa" ||
		item["size"] != float64(10) || item["modified_at"] != "2024-03-05T10:00:00Z" || item["created_at"] != "2024-03-05T10:00:00Z" ||
		item["starred"] != true || item["index_state"] != "partial" {
		t.Fatalf("unexpected item %v", item)
	}
	if payload.Pagination["has_next"] != true || payload.Pagination["page_size"] != float64(50) {
		t.Fatalf("unexpected pagination %v", payload.Pagination)
	}
}

func TestListDocumentsHandlerPassesParsedParameters(t *testing.T) {
	service := &stubService{}

	recorder := performListRequest(NewHandler(service), "/documents?format=pdf&folder_id=4&modified_from=2024-03-01&modified_to=2024-03-31&favorites_only=true&index_state=queued&page=2&page_size=10&sort=size&order=asc")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	params := service.gotListParams
	if params.Filter.Format != GroupPDF || params.FolderID != 4 || params.Filter.ModifiedFrom == nil || params.Filter.ModifiedTo == nil ||
		!params.Filter.FavoritesOnly || params.Filter.IndexState != StateQueued || params.Page != 2 || params.PageSize != 10 ||
		params.Sort != SortSize || params.Order != OrderAscending {
		t.Fatalf("unexpected params %+v", params)
	}
}

func TestListDocumentsHandlerDefaultsSortAndOrder(t *testing.T) {
	testCases := map[string]struct {
		target    string
		wantSort  string
		wantOrder string
	}{
		"nothing given":    {"/documents", SortModifiedAt, OrderDescending},
		"name":             {"/documents?sort=name", SortName, OrderAscending},
		"size":             {"/documents?sort=size", SortSize, OrderDescending},
		"explicit reverse": {"/documents?sort=name&order=desc", SortName, OrderDescending},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := &stubService{}
			performListRequest(NewHandler(service), testCase.target)
			if service.gotListParams.Sort != testCase.wantSort || service.gotListParams.Order != testCase.wantOrder {
				t.Fatalf("got %s %s", service.gotListParams.Sort, service.gotListParams.Order)
			}
		})
	}
}

func TestDocumentHandlersRejectInvalidValuesWithoutCallingService(t *testing.T) {
	targets := map[string]string{
		"format":         "format=video",
		"folder not int": "folder_id=abc",
		"folder zero":    "folder_id=0",
		"date":           "modified_from=01/03/2024",
		"inverted range": "modified_from=2024-03-10&modified_to=2024-03-01",
		"favorites":      "favorites_only=yes",
		"index state":    "index_state=done",
	}
	for name, rawQuery := range targets {
		for _, route := range []string{"/documents", "/documents/count"} {
			t.Run(route+" "+name, func(t *testing.T) {
				service := &stubService{documentPage: utils.PaginationResponse[DocumentItemDto]{Items: []DocumentItemDto{{ID: 1}}}}
				recorder := performListRequest(NewHandler(service), route+"?"+rawQuery)
				if recorder.Code != http.StatusBadRequest || service.gotListParams.Page != 0 || service.gotCountFolder != 0 {
					t.Fatalf("status %d params %+v", recorder.Code, service.gotListParams)
				}
				var payload map[string]string
				if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil || payload["error"] == "" {
					t.Fatalf("expected error message, got %s", recorder.Body.String())
				}
			})
		}
	}
}

func TestListDocumentsHandlerRejectsInvalidSortOrderAndPagination(t *testing.T) {
	for _, rawQuery := range []string{"sort=color", "order=sideways", "page=abc", "page_size=x"} {
		recorder := performListRequest(NewHandler(&stubService{}), "/documents?"+rawQuery)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: status %d", rawQuery, recorder.Code)
		}
	}
}

func TestDocumentHandlersMapFolderNotFoundAndFailures(t *testing.T) {
	for _, route := range []string{"/documents", "/documents/count"} {
		notFound := performListRequest(NewHandler(&stubService{err: ErrFolderNotFound}), route+"?folder_id=9")
		if notFound.Code != http.StatusNotFound {
			t.Fatalf("%s: status %d", route, notFound.Code)
		}
		failure := performListRequest(NewHandler(&stubService{err: errors.New("boom")}), route)
		if failure.Code != http.StatusInternalServerError || failure.Body.String() == "" {
			t.Fatalf("%s: status %d", route, failure.Code)
		}
	}
}

func TestCountDocumentsHandlerFixesResponseFieldNames(t *testing.T) {
	service := &stubService{documentCount: DocumentCountDto{Total: 1234}}

	recorder := performListRequest(NewHandler(service), "/documents/count?format=text&folder_id=3")

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK || len(payload) != 1 || payload["total"] != float64(1234) {
		t.Fatalf("status %d payload %v", recorder.Code, payload)
	}
	if service.gotCountFilter.Format != GroupText || service.gotCountFolder != 3 {
		t.Fatalf("unexpected call %+v %d", service.gotCountFilter, service.gotCountFolder)
	}
}
