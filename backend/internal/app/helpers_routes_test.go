package app

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nas-go/api/internal/api/v1/accesscontrol"
	"nas-go/api/internal/api/v1/analytics"
	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/music"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/api/v1/search"
	"nas-go/api/internal/api/v1/updater"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/api/v1/watchfolders"
	"nas-go/api/internal/config"

	"github.com/gin-gonic/gin"
)

func buildRouteContext() *AppContext {
	return &AppContext{
		Files:         &FileContext{Handler: files.NewHandler(nil, nil, nil)},
		Jobs:          &JobsContext{Handler: jobs.NewHandler(nil)},
		Diary:         &DiaryContext{Handler: diary.NewHandler(nil, nil)},
		Music:         &MusicContext{Handler: music.NewHandler(nil, nil, nil)},
		Video:         &VideoContext{Handler: video.NewHandler(nil, nil, nil)},
		Analytics:     &AnalyticsContext{Handler: analytics.NewHandler(nil)},
		Configuration: &ConfigurationContext{Handler: configuration.NewHandler(nil, nil)},
		Search:        &SearchContext{Handler: search.NewHandler(nil)},
		Notifications: &NotificationContext{Handler: notifications.NewHandler(nil)},
		WatchFolders:  &WatchFoldersContext{Handler: watchfolders.NewHandler(nil, nil)},
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
		{method: http.MethodGet, path: "/api/v1/analytics/storage"},
		{method: http.MethodGet, path: "/api/v1/analytics/health"},
		{method: http.MethodGet, path: "/api/v1/analytics/duplicates/groups"},
		{method: http.MethodGet, path: "/api/v1/configuration/about"},
		{method: http.MethodGet, path: "/api/v1/configuration/settings"},
		{method: http.MethodPut, path: "/api/v1/configuration/settings"},
		{method: http.MethodPost, path: "/api/v1/update/apply"},
		{method: http.MethodGet, path: "/api/v1/search/global"},
		{method: http.MethodGet, path: "/api/v1/notifications"},
		{method: http.MethodGet, path: "/api/v1/notifications/unread-count"},
		{method: http.MethodGet, path: "/api/v1/notifications/:id"},
		{method: http.MethodPut, path: "/api/v1/notifications/:id/read"},
		{method: http.MethodPut, path: "/api/v1/notifications/read-all"},
		{method: http.MethodGet, path: "/api/v1/watch-folders"},
		{method: http.MethodGet, path: "/api-docs/openapi.json"},
		{method: http.MethodGet, path: "/swagger/*any"},
	}
	for _, check := range checks {
		if !routeExists(routes, check.method, check.path) {
			t.Fatalf("expected route %s %s to exist", check.method, check.path)
		}
	}
}

func TestRegisterSwaggerRoutes(t *testing.T) {
	router := SetUpRouter()
	registerSwaggerRoutes(router)

	specReq := httptest.NewRequest(http.MethodGet, "/api-docs/openapi.json", nil)
	specWriter := httptest.NewRecorder()
	router.ServeHTTP(specWriter, specReq)

	if specWriter.Code != http.StatusOK {
		t.Fatalf("expected openapi spec route to return 200, got %d", specWriter.Code)
	}
	if got := specWriter.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected openapi spec response as json, got %q", got)
	}
	if !strings.Contains(specWriter.Body.String(), "\"openapi\": \"3.0.3\"") {
		t.Fatalf("expected openapi version in response body")
	}

	uiReq := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	uiWriter := httptest.NewRecorder()
	router.ServeHTTP(uiWriter, uiReq)

	if uiWriter.Code != http.StatusOK {
		t.Fatalf("expected swagger ui route to return 200, got %d", uiWriter.Code)
	}
	if body := uiWriter.Body.String(); !strings.Contains(body, "Swagger UI") {
		t.Fatalf("expected swagger ui html response")
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

func setupDistDir(t *testing.T) {
	t.Helper()
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
	jsPath := filepath.Join(assetsDir, "vendor-mui-abc123.js")
	if err := os.WriteFile(jsPath, []byte("console.log('mui')"), 0644); err != nil {
		t.Fatalf("failed to write js asset: %v", err)
	}
}

func TestRegisterReactRoutes_NoRouteServesIndexAndAssetsRouteIsRegistered(t *testing.T) {
	setupDistDir(t)

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

func TestRegisterReactRoutes_IndexHasNoCacheHeader(t *testing.T) {
	setupDistDir(t)

	router := SetUpRouter()
	registerReactRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/some/route", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("expected Cache-Control 'no-cache' for index, got %q", got)
	}
}

func TestRegisterReactRoutes_AssetsHaveImmutableCacheHeader(t *testing.T) {
	setupDistDir(t)

	router := SetUpRouter()
	registerReactRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/assets/vendor-mui-abc123.js", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for asset, got %d", w.Code)
	}
	expected := "public, max-age=31536000, immutable"
	if got := w.Header().Get("Cache-Control"); got != expected {
		t.Fatalf("expected Cache-Control %q for assets, got %q", expected, got)
	}
}

func TestWebDAVRouteIsGatedByConfig(t *testing.T) {
	setAllowedOriginsForTest(t)
	previous := config.AppConfig.EnableWebDAV
	t.Cleanup(func() { config.AppConfig.EnableWebDAV = previous })

	// Default (disabled): /dav does not exist.
	config.AppConfig.EnableWebDAV = false
	router := SetUpRouter()
	RegisterRoutes(router, buildRouteContext())
	for _, route := range router.Routes() {
		if strings.HasPrefix(route.Path, "/dav") {
			t.Fatalf("expected no /dav route with WebDAV disabled, found %s %s", route.Method, route.Path)
		}
	}

	// Enabled: /dav answers WebDAV verbs.
	config.AppConfig.EnableWebDAV = true
	enabledRouter := SetUpRouter()
	RegisterRoutes(enabledRouter, buildRouteContext())

	found := false
	for _, route := range enabledRouter.Routes() {
		if route.Path == "/dav/*path" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected /dav/*path route with WebDAV enabled")
	}
}

// denyAllAccessControl implements accesscontrol.ServiceInterface refusing
// every non-loopback IP, to prove protected routes answer 403.
type denyAllAccessControl struct{}

func (denyAllAccessControl) GetAllowedIPs() ([]accesscontrol.AllowedIPDto, error) { return nil, nil }
func (denyAllAccessControl) CreateAllowedIP(dto accesscontrol.CreateAllowedIPDto) (accesscontrol.AllowedIPDto, error) {
	return accesscontrol.AllowedIPDto{}, nil
}
func (denyAllAccessControl) UpdateAllowedIP(id int, dto accesscontrol.UpdateAllowedIPDto) (accesscontrol.AllowedIPDto, error) {
	return accesscontrol.AllowedIPDto{}, nil
}
func (denyAllAccessControl) DeleteAllowedIP(id int) error   { return nil }
func (denyAllAccessControl) IsAllowed(addr netip.Addr) bool { return false }
func (denyAllAccessControl) Reload() error                  { return nil }

func TestWebDAVSitsBehindTheIPWhitelist(t *testing.T) {
	setAllowedOriginsForTest(t)
	previous := config.AppConfig.EnableWebDAV
	t.Cleanup(func() { config.AppConfig.EnableWebDAV = previous })
	config.AppConfig.EnableWebDAV = true

	context := buildRouteContext()
	context.AccessControl = &AccessControlContext{Service: denyAllAccessControl{}}

	router := SetUpRouter()
	RegisterRoutes(router, context)

	for _, method := range []string{http.MethodGet, http.MethodPut, "PROPFIND", "MKCOL", http.MethodDelete} {
		request := httptest.NewRequest(method, "/dav/qualquer", nil)
		request.RemoteAddr = "192.168.1.99:50000" // non-loopback, not whitelisted
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("%s /dav: expected 403 from the whitelist, got %d", method, recorder.Code)
		}
	}
}
