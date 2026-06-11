package accesscontrol

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newWhitelistedRouter(t *testing.T, cidrs ...string) (*gin.Engine, ServiceInterface) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	service := NewService(newFakeRepository())
	for _, cidr := range cidrs {
		if _, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: cidr}); err != nil {
			t.Fatalf("seed whitelist with %q: %v", cidr, err)
		}
	}

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		t.Fatalf("SetTrustedProxies: %v", err)
	}
	router.Use(NewMiddleware(service))
	router.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	return router, service
}

func requestFrom(router *gin.Engine, remoteAddr string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = remoteAddr
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestMiddlewareLoopbackAlwaysPasses(t *testing.T) {
	router, _ := newWhitelistedRouter(t) // empty whitelist

	for _, remote := range []string{"127.0.0.1:5000", "127.0.0.53:5000", "[::1]:5000"} {
		if rec := requestFrom(router, remote, nil); rec.Code != http.StatusOK {
			t.Fatalf("loopback %q must always pass, got %d", remote, rec.Code)
		}
	}
}

func TestMiddlewareBlocksUnknownIPWithI18nBody(t *testing.T) {
	router, _ := newWhitelistedRouter(t)

	rec := requestFrom(router, "192.168.1.66:4321", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("unknown IP must get 403, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode 403 body: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("403 body must carry a translated error message")
	}
	if body["ip"] != "192.168.1.66" {
		t.Fatalf("403 body must echo the requester IP, got %q", body["ip"])
	}
}

func TestMiddlewareAllowsRegisteredCIDRRange(t *testing.T) {
	router, _ := newWhitelistedRouter(t, "192.168.1.0/24")

	if rec := requestFrom(router, "192.168.1.42:4321", nil); rec.Code != http.StatusOK {
		t.Fatalf("IP inside registered /24 must pass, got %d", rec.Code)
	}
	if rec := requestFrom(router, "192.168.2.42:4321", nil); rec.Code != http.StatusForbidden {
		t.Fatalf("IP outside registered /24 must be blocked, got %d", rec.Code)
	}
}

func TestMiddlewareDisabledEntryDoesNotPass(t *testing.T) {
	router, service := newWhitelistedRouter(t, "192.168.1.42")

	entries, err := service.GetAllowedIPs()
	if err != nil || len(entries) != 1 {
		t.Fatalf("seeded entry missing: %v", err)
	}
	disabled := false
	if _, err := service.UpdateAllowedIP(entries[0].ID, UpdateAllowedIPDto{Enabled: &disabled}); err != nil {
		t.Fatalf("disable entry: %v", err)
	}

	if rec := requestFrom(router, "192.168.1.42:4321", nil); rec.Code != http.StatusForbidden {
		t.Fatalf("disabled entry must not grant access, got %d", rec.Code)
	}
}

func TestMiddlewareForgedProxyHeadersDoNotBypass(t *testing.T) {
	router, _ := newWhitelistedRouter(t, "192.168.1.42")

	headers := map[string]string{
		"X-Forwarded-For": "192.168.1.42",
		"X-Real-IP":       "127.0.0.1",
	}
	if rec := requestFrom(router, "10.9.9.9:4321", headers); rec.Code != http.StatusForbidden {
		t.Fatalf("forged proxy headers must not bypass the whitelist, got %d", rec.Code)
	}
}

func TestMiddlewareIPv4MappedClientMatchesIPv4Entry(t *testing.T) {
	router, _ := newWhitelistedRouter(t, "192.168.1.42")

	// An IPv4 client arriving through an IPv6 dual-stack socket.
	if rec := requestFrom(router, "[::ffff:192.168.1.42]:4321", nil); rec.Code != http.StatusOK {
		t.Fatalf("IPv4-mapped client must match its IPv4 entry, got %d", rec.Code)
	}
}

func TestMiddlewareUnparsableRemoteAddrIsBlocked(t *testing.T) {
	router, _ := newWhitelistedRouter(t, "192.168.1.0/24")

	if rec := requestFrom(router, "garbage", nil); rec.Code != http.StatusForbidden {
		t.Fatalf("unparsable remote addr must be blocked, got %d", rec.Code)
	}
}
