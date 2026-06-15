package ingest

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeYtDlpService struct {
	status    YtDlpStatusDto
	updateErr error
}

func (s *fakeYtDlpService) Status() YtDlpStatusDto { return s.status }
func (s *fakeYtDlpService) Update() error          { return s.updateErr }

func newYtDlpRouter(service YtDlpServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewYtDlpHandler(service)
	router := gin.New()
	router.GET("/ingest/ytdlp/status", handler.GetStatusHandler)
	router.POST("/ingest/ytdlp/update", handler.UpdateHandler)
	return router
}

func TestYtDlpStatusHandler(t *testing.T) {
	router := newYtDlpRouter(&fakeYtDlpService{status: YtDlpStatusDto{Installed: true, CurrentVersion: "2024.08.06", LatestVersion: "2024.09.01", UpdateAvailable: true}})
	rec := doRequest(router, http.MethodGet, "/ingest/ytdlp/status", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "2024.09.01") {
		t.Fatalf("status: code %d body %s", rec.Code, rec.Body.String())
	}
}

func TestYtDlpUpdateHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := newYtDlpRouter(&fakeYtDlpService{})
		rec := doRequest(router, http.MethodPost, "/ingest/ytdlp/update", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})
	t.Run("failure", func(t *testing.T) {
		router := newYtDlpRouter(&fakeYtDlpService{updateErr: errString("boom")})
		rec := doRequest(router, http.MethodPost, "/ingest/ytdlp/update", "")
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}
