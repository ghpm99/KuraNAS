package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newPaginationContext(rawQuery string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/?"+rawQuery, nil)
	return ctx, recorder
}

func TestParsePaginationNormalizesValues(t *testing.T) {
	tests := []struct {
		name             string
		rawQuery         string
		expectedPage     int
		expectedPageSize int
	}{
		{"defaults", "", 1, 15},
		{"explicit values", "page=3&page_size=40", 3, 40},
		{"page below one becomes one", "page=0", 1, 15},
		{"negative page becomes one", "page=-4", 1, 15},
		{"page size zero becomes one", "page_size=0", 1, 1},
		{"negative page size becomes one", "page_size=-9", 1, 1},
		{"page size above max is clamped", "page_size=100000", 1, MaxPageSize},
		{"page size at max is kept", "page_size=500", 1, 500},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, recorder := newPaginationContext(testCase.rawQuery)

			page, pageSize, isValid := ParsePagination(ctx, 15)

			if !isValid {
				t.Fatalf("expected valid pagination")
			}
			if page != testCase.expectedPage || pageSize != testCase.expectedPageSize {
				t.Fatalf("expected page=%d pageSize=%d, got page=%d pageSize=%d", testCase.expectedPage, testCase.expectedPageSize, page, pageSize)
			}
			if ctx.IsAborted() || recorder.Body.Len() != 0 {
				t.Fatalf("valid pagination must not write a response")
			}
		})
	}
}

func TestParsePaginationRejectsNonNumericValues(t *testing.T) {
	for _, rawQuery := range []string{"page=abc", "page_size=xyz", "page=1&page_size=1.5"} {
		t.Run(rawQuery, func(t *testing.T) {
			ctx, recorder := newPaginationContext(rawQuery)

			_, _, isValid := ParsePagination(ctx, 15)

			if isValid {
				t.Fatalf("expected invalid pagination")
			}
			if recorder.Code != http.StatusBadRequest || !ctx.IsAborted() {
				t.Fatalf("expected aborted 400, got %d", recorder.Code)
			}
		})
	}
}
