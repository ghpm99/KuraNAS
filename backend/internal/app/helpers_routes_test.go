package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/internal/api/v1/analytics"
	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/music"
	"nas-go/api/internal/api/v1/updater"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/config"

	"github.com/gin-gonic/gin"
)

func buildRouteContext() *AppContext {
	return &AppContext{
		Files:         &FileContext{Handler: files.NewHandler(nil, nil, nil)},
		Jobs:          &JobsContext{Handler: jobs.NewHandler(nil)},
		Diary:         &DiaryContext{Handler: diary.NewHandler(nil, nil)},
		Music:         &MusicContext{Handler: music.NewHandler(nil, nil)},
		Video:         &VideoContext{Handler: video.NewHandler(nil, nil)},
		Analytics:     &AnalyticsContext{Handler: analytics.NewHandler(nil)},
		Configuration: &ConfigurationContext{Handler: configuration.NewHandler(nil, nil)},
		UpdateHandler: updater.NewHandler(updater.NewService(), nil),
	}
}

func routeExists(routes gin.RoutesInfo, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

func setAllowedOriginsForTest(t *testing.T) {
	t.Helper()
	originalAllowedOrigins := config.AppConfig.AllowedOrigins
	config.AppConfig.AllowedOrigins = "https://github.com,http://localhost:5173"
	t.Cleanup(func() {
		config.AppConfig.AllowedOrigins = originalAllowedOrigins
	})
}

func TestSetUpRouterAndRegisterRoutes(t *testing.T) {
	setAllowedOriginsForTest(t)
	router := SetUpRouter()
	RegisterRoutes(router, buildRouteContext())

	routes := router.Routes()
	if len(routes) == 0 {
		t.Fatalf("expected registered routes")
	}

	checks := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/files/"},
		{method: http.MethodGet, path: "/api/v1/jobs/:id"},
		{method: http.MethodGet, path: "/api/v1/jobs"},
		{method: http.MethodGet, path: "/api/v1/jobs/:id/steps"},
		{method: http.MethodPost, path: "/api/v1/jobs/:id/cancel"},
		{method: http.MethodGet, path: "/api/v1/diary/summary"},
		{method: http.MethodGet, path: "/api/v1/music/playlists/"},
		{method: http.MethodGet, path: "/api/v1/video/playlists/unassigned"},
		{method: http.MethodGet, path: "/api/v1/analytics/overview"},
		{method: http.MethodGet, path: "/api/v1/configuration/about"},
		{method: http.MethodGet, path: "/api/v1/configuration/settings"},
		{method: http.MethodPut, path: "/api/v1/configuration/settings"},
		{method: http.MethodPost, path: "/api/v1/update/apply"},
	}
	for _, check := range checks {
		if !routeExists(routes, check.method, check.path) {
			t.Fatalf("expected route %s %s to exist", check.method, check.path)
		}
	}
}

func TestRegisterCorsRoutes(t *testing.T) {
	setAllowedOriginsForTest(t)

	router := SetUpRouter()
	registerCorsRoutes(router, buildRouteContext())
	router.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://github.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentials header to be true, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://github.com" {
		t.Fatalf("expected allowed origin header, got %q", got)
	}

	reqDenied := httptest.NewRequest(http.MethodGet, "/ping", nil)
	reqDenied.Header.Set("Origin", "https://example.com")
	wDenied := httptest.NewRecorder()
	router.ServeHTTP(wDenied, reqDenied)
	if wDenied.Code != http.StatusOK {
		t.Fatalf("expected 200 for denied-origin request too, got %d", wDenied.Code)
	}
	if got := wDenied.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("expected denied-origin request to not include credentials header, got %q", got)
	}
	if got := wDenied.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected denied-origin request to not include origin header, got %q", got)
	}

	reqPreflight := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	reqPreflight.Header.Set("Origin", "https://github.com")
	reqPreflight.Header.Set("Access-Control-Request-Method", http.MethodGet)
	wPreflight := httptest.NewRecorder()
	router.ServeHTTP(wPreflight, reqPreflight)
	if wPreflight.Code != http.StatusNoContent {
		t.Fatalf("expected preflight 204, got %d", wPreflight.Code)
	}
}

func TestRegisterReactRoutes_NoRouteServesIndexAndAssetsRouteIsRegistered(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmpRoot := t.TempDir()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	distDir := "dist"
	assetsDir := filepath.Join(distDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("failed to create dist assets dir: %v", err)
	}
	indexPath := filepath.Join(distDir, "index.html")
	indexContent := []byte("<html><body>kuranas</body></html>")
	if err := os.WriteFile(indexPath, indexContent, 0644); err != nil {
		t.Fatalf("failed to write dist index: %v", err)
	}

	router := SetUpRouter()
	registerReactRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/some/unknown/route", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected NoRoute to serve index with 200, got %d", w.Code)
	}
	if body := w.Body.String(); body == "" {
		t.Fatalf("expected index response body for NoRoute")
	}
}
