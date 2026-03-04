package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/diary"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/music"
	"nas-go/api/internal/api/v1/updater"
	"nas-go/api/internal/api/v1/video"

	"github.com/gin-gonic/gin"
)

func buildRouteContext() *AppContext {
	return &AppContext{
		Files:                &FileContext{Handler: files.NewHandler(nil, nil, nil)},
		Diary:                &DiaryContext{Handler: diary.NewHandler(nil, nil)},
		Music:                &MusicContext{Handler: music.NewHandler(nil, nil)},
		Video:                &VideoContext{Handler: video.NewHandler(nil, nil)},
		ConfigurationHandler: configuration.NewHandler(nil),
		UpdateHandler:        updater.NewHandler(updater.NewService(), nil),
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

func TestSetUpRouterAndRegisterRoutes(t *testing.T) {
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
		{method: http.MethodGet, path: "/api/v1/diary/summary"},
		{method: http.MethodGet, path: "/api/v1/music/playlists/"},
		{method: http.MethodGet, path: "/api/v1/video/playlists/unassigned"},
		{method: http.MethodGet, path: "/api/v1/configuration/about"},
		{method: http.MethodPost, path: "/api/v1/update/apply"},
	}
	for _, check := range checks {
		if !routeExists(routes, check.method, check.path) {
			t.Fatalf("expected route %s %s to exist", check.method, check.path)
		}
	}
}

func TestRegisterCorsRoutes(t *testing.T) {
	router := SetUpRouter()
	registerCorsRoutes(router)
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
}
