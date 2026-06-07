package distribution

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

// writeManifest writes a manifest.json into dir from the given artifacts.
func writeManifest(t *testing.T, dir string, artifacts []Artifact) {
	t.Helper()
	data, err := json.Marshal(manifest{Artifacts: artifacts})
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifestFileName), data, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

// writeFile drops a file with content into dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", name, err)
	}
}

func TestRepositoryListArtifactsReturnsOnlyPresentFiles(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, []Artifact{
		{ID: "android", Platform: "android", File: "app.apk", Version: "1.0.0"},
		{ID: "plugin", Platform: "browser", File: "plugin.zip", Version: "2.0.0"},
		{ID: "ghost", Platform: "android", File: "missing.apk", Version: "9.9.9"},
	})
	writeFile(t, dir, "app.apk", "apk-bytes")
	writeFile(t, dir, "plugin.zip", "zip")

	repo := NewRepository(dir)
	artifacts, err := repo.ListArtifacts()
	if err != nil {
		t.Fatalf("ListArtifacts: %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("expected 2 present artifacts, got %d", len(artifacts))
	}

	byID := map[string]Artifact{}
	for _, a := range artifacts {
		byID[a.ID] = a
	}
	if byID["android"].SizeBytes != int64(len("apk-bytes")) {
		t.Errorf("android size = %d, want %d", byID["android"].SizeBytes, len("apk-bytes"))
	}
	if byID["android"].AbsPath == "" {
		t.Error("android AbsPath not filled")
	}
	if _, ok := byID["ghost"]; ok {
		t.Error("ghost artifact with missing file should be skipped")
	}
}

func TestRepositorySkipsPathTraversalAndDirs(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "downloads")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// a real secret file outside the downloads dir
	writeFile(t, base, "secret.txt", "top-secret")
	// a directory inside downloads (must not be served as a file)
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}
	writeManifest(t, dir, []Artifact{
		{ID: "escape", File: "../secret.txt"},
		{ID: "dir", File: "subdir"},
		{ID: "blank", File: "   "},
	})

	repo := NewRepository(dir)
	artifacts, err := repo.ListArtifacts()
	if err != nil {
		t.Fatalf("ListArtifacts: %v", err)
	}
	if len(artifacts) != 0 {
		t.Fatalf("expected nothing servable, got %d: %+v", len(artifacts), artifacts)
	}
}

func TestRepositoryMissingManifestIsEmptyCatalog(t *testing.T) {
	repo := NewRepository(t.TempDir())
	artifacts, err := repo.ListArtifacts()
	if err != nil {
		t.Fatalf("expected nil error for missing manifest, got %v", err)
	}
	if len(artifacts) != 0 {
		t.Fatalf("expected empty catalog, got %d", len(artifacts))
	}
}

func TestRepositoryInvalidManifestErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, manifestFileName), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	repo := NewRepository(dir)
	if _, err := repo.ListArtifacts(); err == nil {
		t.Fatal("expected error for invalid manifest json")
	}
	if _, err := repo.GetArtifact("any"); err == nil {
		t.Fatal("expected error for invalid manifest json on GetArtifact")
	}
}

func TestRepositoryGetArtifact(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, []Artifact{
		{ID: "android", File: "app.apk"},
		{ID: "ghost", File: "missing.apk"},
	})
	writeFile(t, dir, "app.apk", "bytes")
	repo := NewRepository(dir)

	got, err := repo.GetArtifact("android")
	if err != nil {
		t.Fatalf("GetArtifact: %v", err)
	}
	if got.AbsPath == "" {
		t.Error("AbsPath not filled")
	}

	if _, err := repo.GetArtifact("ghost"); !errors.Is(err, ErrArtifactNotFound) {
		t.Errorf("ghost: expected ErrArtifactNotFound, got %v", err)
	}
	if _, err := repo.GetArtifact("unknown"); !errors.Is(err, ErrArtifactNotFound) {
		t.Errorf("unknown: expected ErrArtifactNotFound, got %v", err)
	}
}

func TestServiceListDownloadsMapsFields(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, []Artifact{
		{ID: "android", Platform: "android", File: "app.apk", Version: "1.0.0", MinOS: "Android 13", SHA256: "abc"},
	})
	writeFile(t, dir, "app.apk", "bytes")

	service := NewService(NewRepository(dir))
	items, err := service.ListDownloads()
	if err != nil {
		t.Fatalf("ListDownloads: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.DownloadURL != "/api/v1/downloads/android" {
		t.Errorf("DownloadURL = %q", item.DownloadURL)
	}
	if item.Platform != "android" || item.Version != "1.0.0" || item.MinOS != "Android 13" || item.SHA256 != "abc" {
		t.Errorf("unexpected mapping: %+v", item)
	}
	if item.SizeBytes != int64(len("bytes")) {
		t.Errorf("SizeBytes = %d", item.SizeBytes)
	}
	// With no NameKey set, the name falls back to the id instead of a raw key.
	if item.Name != "android" {
		t.Errorf("Name fallback = %q, want %q", item.Name, "android")
	}
}

func TestServiceResolveDownload(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, []Artifact{{ID: "android", File: "app.apk"}})
	writeFile(t, dir, "app.apk", "bytes")
	service := NewService(NewRepository(dir))

	path, filename, err := service.ResolveDownload("android")
	if err != nil {
		t.Fatalf("ResolveDownload: %v", err)
	}
	if filename != "app.apk" {
		t.Errorf("filename = %q", filename)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Errorf("resolved path not a real file: %v", statErr)
	}

	if _, _, err := service.ResolveDownload("nope"); !errors.Is(err, ErrArtifactNotFound) {
		t.Errorf("expected ErrArtifactNotFound, got %v", err)
	}
}

func TestResolveKeyFallback(t *testing.T) {
	if got := resolveKey("", "fallback"); got != "fallback" {
		t.Errorf("empty key: got %q", got)
	}
	// An unknown key resolves to itself in i18n, so we must fall back.
	if got := resolveKey("DEFINITELY_MISSING_KEY", "fb"); got != "fb" {
		t.Errorf("missing key: got %q, want fb", got)
	}
}

func TestBaseName(t *testing.T) {
	cases := map[string]string{
		"app.apk":          "app.apk",
		"nested/app.apk":   "app.apk",
		"a\\b\\plugin.zip": "plugin.zip",
		"":                 "",
	}
	for in, want := range cases {
		if got := baseName(in); got != want {
			t.Errorf("baseName(%q) = %q, want %q", in, got, want)
		}
	}
}

// serviceStub lets handler tests drive specific error and success paths.
type serviceStub struct {
	listFn    func() ([]DownloadItemDto, error)
	resolveFn func(id string) (string, string, error)
}

func (s serviceStub) ListDownloads() ([]DownloadItemDto, error) { return s.listFn() }
func (s serviceStub) ResolveDownload(id string) (string, string, error) {
	return s.resolveFn(id)
}

func newRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/downloads", handler.GetDownloadsHandler)
	router.GET("/api/v1/downloads/:id", handler.DownloadFileHandler)
	return router
}

func TestHandlerListSuccess(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, []Artifact{{ID: "android", File: "app.apk", Version: "1.0.0"}})
	writeFile(t, dir, "app.apk", "bytes")
	handler := NewHandler(NewService(NewRepository(dir)))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloads", nil)
	newRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var items []DownloadItemDto
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(items) != 1 || items[0].ID != "android" {
		t.Fatalf("unexpected body: %+v", items)
	}
}

func TestHandlerListError(t *testing.T) {
	handler := NewHandler(serviceStub{
		listFn: func() ([]DownloadItemDto, error) { return nil, errors.New("boom") },
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloads", nil)
	newRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestHandlerDownloadSuccess(t *testing.T) {
	dir := t.TempDir()
	writeManifest(t, dir, []Artifact{{ID: "android", File: "app.apk"}})
	writeFile(t, dir, "app.apk", "apk-content")
	handler := NewHandler(NewService(NewRepository(dir)))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloads/android", nil)
	newRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.String() != "apk-content" {
		t.Errorf("body = %q", rec.Body.String())
	}
	if cd := rec.Header().Get("Content-Disposition"); cd == "" {
		t.Error("missing Content-Disposition header")
	}
}

func TestHandlerDownloadNotFound(t *testing.T) {
	handler := NewHandler(serviceStub{
		resolveFn: func(id string) (string, string, error) { return "", "", ErrArtifactNotFound },
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloads/nope", nil)
	newRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestHandlerDownloadGenericError(t *testing.T) {
	handler := NewHandler(serviceStub{
		resolveFn: func(id string) (string, string, error) { return "", "", errors.New("disk error") },
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/downloads/x", nil)
	newRouter(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}
