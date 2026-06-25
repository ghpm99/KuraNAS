package storageroots

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// serviceStub lets each test pin the service outcome to exercise every
// handler/error-mapping branch without touching the filesystem.
type serviceStub struct {
	roots []StorageRootDto
	dto   StorageRootDto
	err   error

	capturedCreate   *CreateStorageRootDto
	capturedUpdate   *UpdateStorageRootDto
	capturedUpdateID int
}

func (s *serviceStub) GetRoots() ([]StorageRootDto, error) { return s.roots, s.err }
func (s *serviceStub) CreateRoot(request CreateStorageRootDto) (StorageRootDto, error) {
	s.capturedCreate = &request
	return s.dto, s.err
}
func (s *serviceStub) UpdateRoot(id int, request UpdateStorageRootDto) (StorageRootDto, error) {
	s.capturedUpdate = &request
	s.capturedUpdateID = id
	return s.dto, s.err
}
func (s *serviceStub) DeleteRoot(id int) error { return s.err }
func (s *serviceStub) ReloadRegistry() error   { return s.err }

func newRootsRouter(stub *serviceStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(stub, nil)

	router := gin.New()
	group := router.Group("/storage-roots")
	group.GET("", handler.GetStorageRootsHandler)
	group.POST("", handler.CreateStorageRootHandler)
	group.PUT("/:id", handler.UpdateStorageRootHandler)
	group.DELETE("/:id", handler.DeleteStorageRootHandler)
	return router
}

func doRootsJSON(router *gin.Engine, method, url string, payload any) *httptest.ResponseRecorder {
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

func TestGetStorageRootsHandler(t *testing.T) {
	router := newRootsRouter(&serviceStub{roots: []StorageRootDto{{ID: 1, Label: "primary"}}})

	rec := doRootsJSON(router, http.MethodGet, "/storage-roots", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var payload []StorageRootDto
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(payload) != 1 || payload[0].Label != "primary" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestCreateStorageRootHandlerSuccess(t *testing.T) {
	router := newRootsRouter(&serviceStub{dto: StorageRootDto{ID: 2, Label: "midia"}})

	rec := doRootsJSON(router, http.MethodPost, "/storage-roots", CreateStorageRootDto{Path: "/x"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", rec.Code, rec.Body.String())
	}
}

// TestCreateStorageRootHandlerDecodesPayload pins the request seam: it proves
// the handler decodes the exact JSON the frontend sends (service/storageRoots.ts
// → POST /storage-roots) into CreateStorageRootDto, including the optional
// *bool enabled. A json tag drift fails here instead of breaking the frontend
// integration silently.
func TestCreateStorageRootHandlerDecodesPayload(t *testing.T) {
	stub := &serviceStub{dto: StorageRootDto{ID: 2}}
	router := newRootsRouter(stub)

	enabled := false
	rec := doRootsJSON(router, http.MethodPost, "/storage-roots", map[string]any{
		"path":    "/mnt/midia",
		"label":   "Mídia",
		"enabled": enabled,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", rec.Code, rec.Body.String())
	}
	if stub.capturedCreate == nil {
		t.Fatal("service did not receive the create request")
	}
	got := *stub.capturedCreate
	if got.Path != "/mnt/midia" || got.Label != "Mídia" {
		t.Fatalf("path/label did not decode: %+v", got)
	}
	if got.Enabled == nil || *got.Enabled != false {
		t.Fatalf("enabled did not decode as a *bool: %v", got.Enabled)
	}
}

// TestUpdateStorageRootHandlerDecodesPayload proves the PUT body decodes into
// UpdateStorageRootDto and the id path param is forwarded to the service.
func TestUpdateStorageRootHandlerDecodesPayload(t *testing.T) {
	stub := &serviceStub{dto: StorageRootDto{ID: 5}}
	router := newRootsRouter(stub)

	rec := doRootsJSON(router, http.MethodPut, "/storage-roots/5", map[string]any{"label": "renomeado", "enabled": true})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if stub.capturedUpdateID != 5 {
		t.Fatalf("expected id 5 forwarded, got %d", stub.capturedUpdateID)
	}
	if stub.capturedUpdate == nil || stub.capturedUpdate.Label != "renomeado" {
		t.Fatalf("label did not decode: %+v", stub.capturedUpdate)
	}
	if stub.capturedUpdate.Enabled == nil || *stub.capturedUpdate.Enabled != true {
		t.Fatalf("enabled did not decode as a *bool: %v", stub.capturedUpdate.Enabled)
	}
}

func TestCreateStorageRootHandlerBadJSON(t *testing.T) {
	router := newRootsRouter(&serviceStub{})

	req := httptest.NewRequest(http.MethodPost, "/storage-roots", bytes.NewBufferString("{not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateStorageRootHandlerSuccess(t *testing.T) {
	router := newRootsRouter(&serviceStub{dto: StorageRootDto{ID: 2, Label: "renamed"}})

	rec := doRootsJSON(router, http.MethodPut, "/storage-roots/2", UpdateStorageRootDto{Label: "renamed"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestUpdateStorageRootHandlerInvalidID(t *testing.T) {
	router := newRootsRouter(&serviceStub{})

	rec := doRootsJSON(router, http.MethodPut, "/storage-roots/abc", UpdateStorageRootDto{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateStorageRootHandlerBadJSON(t *testing.T) {
	router := newRootsRouter(&serviceStub{})

	req := httptest.NewRequest(http.MethodPut, "/storage-roots/2", bytes.NewBufferString("{not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteStorageRootHandlerSuccess(t *testing.T) {
	router := newRootsRouter(&serviceStub{})

	rec := doRootsJSON(router, http.MethodDelete, "/storage-roots/3", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestDeleteStorageRootHandlerInvalidID(t *testing.T) {
	router := newRootsRouter(&serviceStub{})

	rec := doRootsJSON(router, http.MethodDelete, "/storage-roots/0", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestServiceErrorMapping(t *testing.T) {
	cases := []struct {
		err  error
		code int
	}{
		{ErrRootNotFound, http.StatusNotFound},
		{ErrInvalidRootPath, http.StatusBadRequest},
		{ErrInvalidRootLabel, http.StatusBadRequest},
		{ErrOverlappingRoot, http.StatusBadRequest},
		{ErrDuplicateRoot, http.StatusBadRequest},
		{ErrPrimaryRootImmutable, http.StatusBadRequest},
		{bytes.ErrTooLarge, http.StatusInternalServerError},
	}

	for _, testCase := range cases {
		router := newRootsRouter(&serviceStub{err: testCase.err})

		rec := doRootsJSON(router, http.MethodPost, "/storage-roots", CreateStorageRootDto{Path: "/x"})
		if rec.Code != testCase.code {
			t.Fatalf("error %v: expected %d, got %d", testCase.err, testCase.code, rec.Code)
		}

		rec = doRootsJSON(router, http.MethodGet, "/storage-roots", nil)
		if rec.Code != testCase.code && rec.Code != http.StatusInternalServerError {
			t.Fatalf("GET with %v: unexpected status %d", testCase.err, rec.Code)
		}

		rec = doRootsJSON(router, http.MethodPut, "/storage-roots/1", UpdateStorageRootDto{})
		if rec.Code != testCase.code {
			t.Fatalf("PUT with %v: expected %d, got %d", testCase.err, testCase.code, rec.Code)
		}

		rec = doRootsJSON(router, http.MethodDelete, "/storage-roots/1", nil)
		if rec.Code != testCase.code {
			t.Fatalf("DELETE with %v: expected %d, got %d", testCase.err, testCase.code, rec.Code)
		}
	}
}
