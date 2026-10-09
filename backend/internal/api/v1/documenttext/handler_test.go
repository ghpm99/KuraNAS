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
	gotQuery  string
	gotPage   int
	gotSize   int
	wasCalled bool
}

func (service *stubService) SearchDocuments(query string, page int, pageSize int) (utils.PaginationResponse[DocumentSearchResultDto], error) {
	service.wasCalled = true
	service.gotQuery, service.gotPage, service.gotSize = query, page, pageSize
	return service.page, service.err
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
	if service.gotQuery != "contrato" || service.gotPage != 2 || service.gotSize != 10 {
		t.Fatalf("unexpected call %q %d %d", service.gotQuery, service.gotPage, service.gotSize)
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	item := payload["items"].([]any)[0].(map[string]any)
	for _, field := range []string{"file_id", "name", "path", "parent_path", "format", "size", "updated_at", "snippet"} {
		if _, hasField := item[field]; !hasField {
			t.Fatalf("missing field %q in %v", field, item)
		}
	}
}

func TestSearchDocumentsHandlerAppliesDefaultsAndCapsPageSize(t *testing.T) {
	service := &stubService{}

	performSearchRequest(NewHandler(service), "q=contrato")
	if service.gotPage != 1 || service.gotSize != defaultSearchPageSize {
		t.Fatalf("unexpected defaults %d/%d", service.gotPage, service.gotSize)
	}

	performSearchRequest(NewHandler(service), "q=contrato&page_size=400")
	if service.gotSize != maxSearchPageSize {
		t.Fatalf("page size not capped: %d", service.gotSize)
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
