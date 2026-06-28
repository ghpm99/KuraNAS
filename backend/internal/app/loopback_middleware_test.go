package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func runLoopbackMiddleware(remoteAddr string) int {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	ctx.Request = req

	loopbackOnlyMiddleware()(ctx)
	return rec.Code
}

func TestLoopbackMiddlewareAllowsLoopback(t *testing.T) {
	if code := runLoopbackMiddleware("127.0.0.1:54321"); code != http.StatusOK {
		t.Fatalf("expected loopback IPv4 to pass, got %d", code)
	}
	if code := runLoopbackMiddleware("[::1]:54321"); code != http.StatusOK {
		t.Fatalf("expected loopback IPv6 to pass, got %d", code)
	}
}

func TestLoopbackMiddlewareRejectsRemote(t *testing.T) {
	if code := runLoopbackMiddleware("192.168.1.10:54321"); code != http.StatusForbidden {
		t.Fatalf("expected remote IP to be forbidden, got %d", code)
	}
	if code := runLoopbackMiddleware("garbage"); code != http.StatusForbidden {
		t.Fatalf("expected unparsable addr to be forbidden, got %d", code)
	}
}

func TestEnvRoutesRegistered(t *testing.T) {
	router := SetUpRouter()
	RegisterRoutes(router, buildRouteContext())
	routes := router.Routes()

	expected := [][2]string{
		{http.MethodGet, "/api/v1/configuration/env"},
		{http.MethodPut, "/api/v1/configuration/env"},
		{http.MethodPost, "/api/v1/configuration/env/test-db"},
		{http.MethodPost, "/api/v1/configuration/env/test-path"},
	}
	for _, route := range expected {
		if !routeExists(routes, route[0], route[1]) {
			t.Fatalf("expected route %s %s to be registered", route[0], route[1])
		}
	}
}
