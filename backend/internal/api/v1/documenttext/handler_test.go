package documenttext

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type stubService struct {
	page      utils.PaginationResponse[DocumentSearchResultDto]
	err       error
	gotSearch DocumentSearchParams
	wasCalled bool

	indexState      IndexStateDto
	documentPage    utils.PaginationResponse[DocumentItemDto]
	documentCount   DocumentCountDto
	gotListParams   DocumentListParams
	gotCountFilter  DocumentFilter
	gotCountFolder  int
	jobID           int
	reindexedFileID int
}

func (service *stubService) SearchDocuments(params DocumentSearchParams) (utils.PaginationResponse[DocumentSearchResultDto], error) {
	service.wasCalled = true
	service.gotSearch = params
	return service.page, service.err
}

func (service *stubService) GetIndexState(fileID int) (IndexStateDto, error) {
	return service.indexState, service.err
}

func (service *stubService) ReindexDocument(fileID int) (int, error) {
	service.reindexedFileID = fileID
	return service.jobID, service.err
}

func (service *stubService) ListDocuments(params DocumentListParams) (utils.PaginationResponse[DocumentItemDto], error) {
	service.gotListParams = params
	return service.documentPage, service.err
}

func (service *stubService) CountDocuments(filter DocumentFilter, folderID int) (DocumentCountDto, error) {
	service.gotCountFilter, service.gotCountFolder = filter, folderID
	return service.documentCount, service.err
}

func (service *stubService) SearchTopDocuments(string, int) ([]DocumentSearchResultDto, error) {
	return nil, nil
}

func performSearchRequest(handler *Handler, rawQuery string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/documents/search", handler.SearchDocumentsHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/documents/search?"+rawQuery, nil))
	return recorder
}

func TestSearchDocumentsHandlerReturnsPaginatedItems(t *testing.T) {
	service := &stubService{page: utils.PaginationResponse[DocumentSearchResultDto]{
		Items:      []DocumentSearchResultDto{{FileID: 4, Name: "a.pdf", Snippet: "trecho"}},
		Pagination: utils.Pagination{Page: 2, PageSize: 10, HasNext: true, HasPrev: true},
	}}

	recorder := performSearchRequest(NewHandler(service), "q=contrato&page=2&page_size=10")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	if service.gotSearch.Query != "contrato" || service.gotSearch.Page != 2 || service.gotSearch.PageSize != 10 {
		t.Fatalf("unexpected call %+v", service.gotSearch)
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	item := payload["items"].([]any)[0].(map[string]any)
	for _, field := range []string{"file_id", "name", "path", "parent_path", "format", "size", "updated_at", "snippet", "page"} {
		if _, hasField := item[field]; !hasField {
			t.Fatalf("missing field %q in %v", field, item)
		}
	}
	if item["page"] != nil || item["snippet"] != "trecho" {
		t.Fatalf("unexpected snippet/page %v", item)
	}
}

func TestSearchDocumentsHandlerSerializesPageNumber(t *testing.T) {
	page := 7
	service := &stubService{page: utils.PaginationResponse[DocumentSearchResultDto]{
		Items: []DocumentSearchResultDto{{FileID: 4, Name: "a.pdf", Snippet: "trecho", Page: &page}},
	}}

	recorder := performSearchRequest(NewHandler(service), "q=contrato")

	var payload struct {
		Items []struct {
			Snippet string `json:"snippet"`
			Page    *int   `json:"page"`
		} `json:"items"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Items) != 1 || payload.Items[0].Page == nil || *payload.Items[0].Page != 7 || payload.Items[0].Snippet != "trecho" {
		t.Fatalf("unexpected payload %s", recorder.Body.String())
	}
}

func TestSearchDocumentsHandlerPassesFiltersToService(t *testing.T) {
	service := &stubService{}

	recorder := performSearchRequest(NewHandler(service), "q=contrato&format=pdf&folder_id=9&favorites_only=true&index_state=name_only&modified_from=2026-01-01&modified_to=2026-02-01")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	params := service.gotSearch
	if params.Filter.Format != GroupPDF || params.FolderID != 9 || !params.Filter.FavoritesOnly || params.Filter.IndexState != StateNameOnly || params.Filter.ModifiedFrom == nil || params.Filter.ModifiedTo == nil {
		t.Fatalf("unexpected params %+v", params)
	}
}

func TestSearchDocumentsHandlerUnknownFolderIsNotFound(t *testing.T) {
	recorder := performSearchRequest(NewHandler(&stubService{err: ErrFolderNotFound}), "q=contrato&folder_id=9")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestSearchDocumentsHandlerAppliesDefaultsAndCapsPageSize(t *testing.T) {
	service := &stubService{}

	performSearchRequest(NewHandler(service), "q=contrato")
	if service.gotSearch.Page != 1 || service.gotSearch.PageSize != defaultSearchPageSize {
		t.Fatalf("unexpected defaults %+v", service.gotSearch)
	}

	performSearchRequest(NewHandler(service), "q=contrato&page_size=400")
	if service.gotSearch.PageSize != maxSearchPageSize {
		t.Fatalf("page size not capped: %d", service.gotSearch.PageSize)
	}
}

func TestSearchDocumentsHandlerRejectsInvalidRequests(t *testing.T) {
	testCases := map[string]string{
		"missing query":    "page=1",
		"blank query":      "q=%20%20",
		"single character": "q=a",
		"overlong query":   "q=" + url.QueryEscape(strings.Repeat("a", maxQueryLength+1)),
		"non numeric page": "q=contrato&page=abc",
		"non numeric size": "q=contrato&page_size=x",
		"invalid format":   "q=contrato&format=zip",
		"invalid folder":   "q=contrato&folder_id=abc",
		"invalid date":     "q=contrato&modified_from=ontem",
		"inverted range":   "q=contrato&modified_from=2026-02-01&modified_to=2026-01-01",
		"invalid favorite": "q=contrato&favorites_only=talvez",
		"invalid state":    "q=contrato&index_state=lento",
	}

	for name, rawQuery := range testCases {
		t.Run(name, func(t *testing.T) {
			service := &stubService{}
			recorder := performSearchRequest(NewHandler(service), rawQuery)
			if recorder.Code != http.StatusBadRequest || service.wasCalled {
				t.Fatalf("expected 400 without service call, got %d called=%v", recorder.Code, service.wasCalled)
			}
		})
	}
}

func TestSearchDocumentsHandlerServiceFailureIsInternalError(t *testing.T) {
	recorder := performSearchRequest(NewHandler(&stubService{err: errors.New("boom")}), "q=contrato")

	if recorder.Code != http.StatusInternalServerError || strings.Contains(recorder.Body.String(), "boom") {
		t.Fatalf("expected sanitized 500, got %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestSearchDocumentsHandlerWithoutServiceIsInternalError(t *testing.T) {
	recorder := performSearchRequest(NewHandler(nil), "q=contrato")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
