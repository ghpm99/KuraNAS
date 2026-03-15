package search

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type searchServiceMock struct {
	searchGlobalFn func(query string, limit int) (GlobalSearchResponseDto, error)
}

func (m *searchServiceMock) SearchGlobal(query string, limit int) (GlobalSearchResponseDto, error) {
	return m.searchGlobalFn(query, limit)
}

func TestSearchHandlerReturnsResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&searchServiceMock{
		searchGlobalFn: func(query string, limit int) (GlobalSearchResponseDto, error) {
			if query != "mix" || limit != 4 {
				t.Fatalf("unexpected search input query=%q limit=%d", query, limit)
			}
			return GlobalSearchResponseDto{Query: query, Files: []FileResultDto{{ID: 1, Name: "song"}}}, nil
		},
	})

	router := gin.New()
	router.GET("/search/global", handler.SearchGlobalHandler)

	req := httptest.NewRequest(http.MethodGet, "/search/global?q=mix&limit=4", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if body := w.Body.String(); body == "" || body == "{}" {
		t.Fatalf("expected payload body, got %q", body)
	}
}

func TestSearchHandlerValidatesLimitAndServiceAvailability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	routerInvalid := gin.New()
	routerInvalid.GET("/search/global", NewHandler(&searchServiceMock{
		searchGlobalFn: func(string, int) (GlobalSearchResponseDto, error) {
			t.Fatal("service should not be called for invalid limit")
			return GlobalSearchResponseDto{}, nil
		},
	}).SearchGlobalHandler)

	reqInvalid := httptest.NewRequest(http.MethodGet, "/search/global?limit=oops", nil)
	wInvalid := httptest.NewRecorder()
	routerInvalid.ServeHTTP(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", wInvalid.Code)
	}

	routerNil := gin.New()
	routerNil.GET("/search/global", NewHandler(nil).SearchGlobalHandler)
	reqNil := httptest.NewRequest(http.MethodGet, "/search/global", nil)
	wNil := httptest.NewRecorder()
	routerNil.ServeHTTP(wNil, reqNil)
	if wNil.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", wNil.Code)
	}
}

func TestSearchHandlerReturnsServerErrorOnServiceFailureAndDefaultLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false
	handler := NewHandler(&searchServiceMock{
		searchGlobalFn: func(query string, limit int) (GlobalSearchResponseDto, error) {
			called = true
			if limit != defaultSearchLimit {
				t.Fatalf("expected default limit %d, got %d", defaultSearchLimit, limit)
			}
			return GlobalSearchResponseDto{}, errors.New("boom")
		},
	})

	router := gin.New()
	router.GET("/search/global", handler.SearchGlobalHandler)

	req := httptest.NewRequest(http.MethodGet, "/search/global?q=test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !called {
		t.Fatalf("expected service to be called")
	}
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
