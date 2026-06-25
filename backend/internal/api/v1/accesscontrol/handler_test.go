package accesscontrol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newHandlerRouter(t *testing.T) (*gin.Engine, ServiceInterface) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	service := NewService(newFakeRepository())
	handler := NewHandler(service, nil)

	router := gin.New()
	group := router.Group("/access-control")
	group.GET("/ips", handler.GetAllowedIPsHandler)
	group.POST("/ips", handler.CreateAllowedIPHandler)
	group.PUT("/ips/:id", handler.UpdateAllowedIPHandler)
	group.DELETE("/ips/:id", handler.DeleteAllowedIPHandler)
	group.GET("/client-ip", handler.GetClientIPHandler)
	return router, service
}

func doJSON(router *gin.Engine, method, url string, payload any) *httptest.ResponseRecorder {
	var body *bytes.Buffer
	if payload != nil {
		raw, _ := json.Marshal(payload)
		body = bytes.NewBuffer(raw)
	} else {
		body = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:9999"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAllowedIPsCRUDHandlers(t *testing.T) {
	router, _ := newHandlerRouter(t)

	// Create
	rec := doJSON(router, http.MethodPost, "/access-control/ips", CreateAllowedIPDto{CIDR: "192.168.1.10", Label: "notebook"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d (%s)", rec.Code, rec.Body.String())
	}
	var created AllowedIPDto
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.CIDR != "192.168.1.10/32" || !created.Enabled {
		t.Fatalf("unexpected created entry: %+v", created)
	}

	// Invalid CIDR → 400
	if rec := doJSON(router, http.MethodPost, "/access-control/ips", CreateAllowedIPDto{CIDR: "bogus"}); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid cidr: expected 400, got %d", rec.Code)
	}
	// Duplicate → 400
	if rec := doJSON(router, http.MethodPost, "/access-control/ips", CreateAllowedIPDto{CIDR: "192.168.1.10/32"}); rec.Code != http.StatusBadRequest {
		t.Fatalf("duplicate: expected 400, got %d", rec.Code)
	}
	// Missing body → 400
	if rec := doJSON(router, http.MethodPost, "/access-control/ips", nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("empty body: expected 400, got %d", rec.Code)
	}

	// List
	rec = doJSON(router, http.MethodGet, "/access-control/ips", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", rec.Code)
	}
	var listed []AllowedIPDto
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil || len(listed) != 1 {
		t.Fatalf("expected 1 listed entry, got %v (%v)", listed, err)
	}

	// Update
	disabled := false
	rec = doJSON(router, http.MethodPut, fmt.Sprintf("/access-control/ips/%d", created.ID), UpdateAllowedIPDto{Enabled: &disabled})
	if rec.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var updated AllowedIPDto
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil || updated.Enabled {
		t.Fatalf("expected disabled entry, got %+v (%v)", updated, err)
	}

	// Update errors: bad id, not found
	if rec := doJSON(router, http.MethodPut, "/access-control/ips/abc", UpdateAllowedIPDto{}); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id: expected 400, got %d", rec.Code)
	}
	if rec := doJSON(router, http.MethodPut, "/access-control/ips/999", UpdateAllowedIPDto{}); rec.Code != http.StatusNotFound {
		t.Fatalf("missing id: expected 404, got %d", rec.Code)
	}

	// Delete
	if rec := doJSON(router, http.MethodDelete, fmt.Sprintf("/access-control/ips/%d", created.ID), nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", rec.Code)
	}
	if rec := doJSON(router, http.MethodDelete, fmt.Sprintf("/access-control/ips/%d", created.ID), nil); rec.Code != http.StatusNotFound {
		t.Fatalf("delete again: expected 404, got %d", rec.Code)
	}
	if rec := doJSON(router, http.MethodDelete, "/access-control/ips/abc", nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("delete bad id: expected 400, got %d", rec.Code)
	}
}

// TestCreateAllowedIPHandlerDecodesPayload pins the request seam: it proves the
// handler decodes the exact JSON the frontend sends (service/accessControl.ts →
// POST /access-control/ips) into CreateAllowedIPDto — including the optional
// label — round-tripping through the real service. A json tag drift fails here
// instead of breaking the frontend integration silently.
func TestCreateAllowedIPHandlerDecodesPayload(t *testing.T) {
	router, _ := newHandlerRouter(t)

	rec := doJSON(router, http.MethodPost, "/access-control/ips", map[string]any{
		"cidr":  "10.0.0.5",
		"label": "notebook-sala",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", rec.Code, rec.Body.String())
	}

	var created AllowedIPDto
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.CIDR != "10.0.0.5/32" {
		t.Fatalf("cidr did not decode/normalize, got %q", created.CIDR)
	}
	if created.Label != "notebook-sala" {
		t.Fatalf("label did not decode, got %q", created.Label)
	}
}

func TestGetClientIPHandlerEchoesConnectionIP(t *testing.T) {
	router, _ := newHandlerRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/access-control/client-ip", nil)
	req.RemoteAddr = "[::ffff:192.168.1.77]:1234"
	// Forged proxy header must not change the reported IP.
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["ip"] != "192.168.1.77" {
		t.Fatalf("expected unmapped connection IP, got %q", body["ip"])
	}
}
