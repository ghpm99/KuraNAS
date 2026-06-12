package trash

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTrashRouter(t *testing.T) (*gin.Engine, *trashRepoMock, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	service, repo, _, root := newTrashServiceForTest(t)
	handler := NewHandler(service, nil)

	router := gin.New()
	group := router.Group("/trash")
	group.GET("", handler.GetTrashItemsHandler)
	group.POST("/:id/restore", handler.RestoreTrashItemHandler)
	group.DELETE("/:id", handler.DeleteTrashItemHandler)
	group.DELETE("", handler.EmptyTrashHandler)
	group.GET("/retention", handler.GetTrashRetentionHandler)
	group.PUT("/retention", handler.UpdateTrashRetentionHandler)
	return router, repo, root
}

func doTrashJSON(router *gin.Engine, method, url string, payload any) *httptest.ResponseRecorder {
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

func TestTrashHandlers_ListRestoreDeleteEmpty(t *testing.T) {
	router, repo, root := newTrashRouter(t)

	// Seed two trashed files through the service.
	service := NewService(repo, &filesIndexMock{})
	for _, name := range []string{"a.txt", "b.txt"} {
		path := filepath.Join(root, name)
		writeFile(t, path, name)
		if err := service.MoveToTrash(path, 1); err != nil {
			t.Fatalf("seed MoveToTrash %s: %v", name, err)
		}
	}

	// List
	rec := doTrashJSON(router, http.MethodGet, "/trash?page=1&page_size=10", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var listResponse struct {
		Items []TrashItemDto `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResponse); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResponse.Items) != 2 {
		t.Fatalf("expected 2 items, got %+v", listResponse.Items)
	}

	// Restore the first
	rec = doTrashJSON(router, http.MethodPost, "/trash/1/restore", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("restore: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err != nil {
		t.Fatalf("restored file missing: %v", err)
	}

	// Restore conflict: occupy b.txt's original path again
	writeFile(t, filepath.Join(root, "b.txt"), "ocupado")
	rec = doTrashJSON(router, http.MethodPost, "/trash/2/restore", nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("restore conflict: expected 409, got %d (%s)", rec.Code, rec.Body.String())
	}
	var conflictBody map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &conflictBody); err != nil || conflictBody["error"] == "" {
		t.Fatalf("conflict must carry an error message, got %s", rec.Body.String())
	}

	// Permanently delete the second
	rec = doTrashJSON(router, http.MethodDelete, "/trash/2", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	// Unknown id → 404
	rec = doTrashJSON(router, http.MethodPost, "/trash/99/restore", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("restore unknown: expected 404, got %d", rec.Code)
	}
	rec = doTrashJSON(router, http.MethodDelete, "/trash/99", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete unknown: expected 404, got %d", rec.Code)
	}

	// Bad id → 400
	rec = doTrashJSON(router, http.MethodPost, "/trash/abc/restore", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("restore bad id: expected 400, got %d", rec.Code)
	}

	// Empty trash
	path := filepath.Join(root, "c.txt")
	writeFile(t, path, "c")
	if err := service.MoveToTrash(path, 1); err != nil {
		t.Fatalf("seed MoveToTrash c.txt: %v", err)
	}
	rec = doTrashJSON(router, http.MethodDelete, "/trash", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("empty: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if items, _ := repo.GetAllItems(); len(items) != 0 {
		t.Fatalf("trash must be empty, got %v", items)
	}
}

func TestTrashHandlers_Retention(t *testing.T) {
	router, _, _ := newTrashRouter(t)

	rec := doTrashJSON(router, http.MethodGet, "/trash/retention", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get retention: expected 200, got %d", rec.Code)
	}
	var retention RetentionDto
	if err := json.Unmarshal(rec.Body.Bytes(), &retention); err != nil {
		t.Fatalf("decode retention: %v", err)
	}
	if retention.Days != DefaultRetentionDays {
		t.Fatalf("expected default %d, got %d", DefaultRetentionDays, retention.Days)
	}

	rec = doTrashJSON(router, http.MethodPut, "/trash/retention", RetentionDto{Days: 7})
	if rec.Code != http.StatusOK {
		t.Fatalf("put retention: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	rec = doTrashJSON(router, http.MethodGet, "/trash/retention", nil)
	_ = json.Unmarshal(rec.Body.Bytes(), &retention)
	if retention.Days != 7 {
		t.Fatalf("expected retention 7, got %d", retention.Days)
	}

	rec = doTrashJSON(router, http.MethodPut, "/trash/retention", RetentionDto{Days: 0})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid retention: expected 400, got %d", rec.Code)
	}
}
