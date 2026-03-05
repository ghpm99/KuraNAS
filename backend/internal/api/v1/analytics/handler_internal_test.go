package analytics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type serviceStub struct {
	response OverviewDto
	err      error
	period   string
}

func (stub *serviceStub) GetOverview(period string) (OverviewDto, error) {
	stub.period = period
	return stub.response, stub.err
}

func TestHandlerGetOverview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		stub := &serviceStub{response: OverviewDto{Period: "7d"}}
		handler := NewHandler(stub)
		router := gin.New()
		router.GET("/analytics/overview", handler.GetOverviewHandler)

		req := httptest.NewRequest(http.MethodGet, "/analytics/overview?period=24h", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if stub.period != "24h" {
			t.Fatalf("expected period 24h, got %s", stub.period)
		}
	})

	t.Run("invalid period", func(t *testing.T) {
		stub := &serviceStub{err: ErrInvalidPeriod}
		handler := NewHandler(stub)
		router := gin.New()
		router.GET("/analytics/overview", handler.GetOverviewHandler)

		req := httptest.NewRequest(http.MethodGet, "/analytics/overview?period=bad", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		stub := &serviceStub{err: errors.New("db")}
		handler := NewHandler(stub)
		router := gin.New()
		router.GET("/analytics/overview", handler.GetOverviewHandler)

		req := httptest.NewRequest(http.MethodGet, "/analytics/overview", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}
