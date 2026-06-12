package dav

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nas-go/api/internal/config"
	"nas-go/api/internal/roots"
)

func setupDAV(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	previousEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = previousEntryPoint
		roots.Reset()
	})
	config.AppConfig.EntryPoint = ""

	dataRoot := t.TempDir()
	mediaRoot := t.TempDir()
	roots.Set([]roots.Root{
		{ID: 1, Path: dataRoot, Label: "Dados", Enabled: true},
		{ID: 2, Path: mediaRoot, Label: "Midia", Enabled: true},
	})
	return NewHandler(), dataRoot, mediaRoot
}

func doDAV(handler http.Handler, method string, target string, body string, headers map[string]string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, target, reader)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestPropfindLevelZeroListsRootLabels(t *testing.T) {
	handler, _, _ := setupDAV(t)

	response := doDAV(handler, "PROPFIND", "/dav/", "", map[string]string{"Depth": "1"})
	if response.Code != http.StatusMultiStatus {
		t.Fatalf("expected 207, got %d (%s)", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, label := range []string{"Dados", "Midia"} {
		if !strings.Contains(body, label) {
			t.Fatalf("expected root %q in PROPFIND response, got %s", label, body)
		}
	}
}

func TestPutAndGetRoundTrip(t *testing.T) {
	handler, dataRoot, _ := setupDAV(t)

	put := doDAV(handler, http.MethodPut, "/dav/Dados/nota.txt", "conteudo", nil)
	if put.Code != http.StatusCreated {
		t.Fatalf("PUT: expected 201, got %d (%s)", put.Code, put.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dataRoot, "nota.txt")); err != nil {
		t.Fatalf("PUT must write through to disk: %v", err)
	}

	get := doDAV(handler, http.MethodGet, "/dav/Dados/nota.txt", "", nil)
	if get.Code != http.StatusOK || get.Body.String() != "conteudo" {
		t.Fatalf("GET: expected the stored bytes, got %d %q", get.Code, get.Body.String())
	}
}

func TestMkcolDeleteAndMoveWithinRoot(t *testing.T) {
	handler, dataRoot, _ := setupDAV(t)

	if response := doDAV(handler, "MKCOL", "/dav/Dados/docs", "", nil); response.Code != http.StatusCreated {
		t.Fatalf("MKCOL: expected 201, got %d", response.Code)
	}
	if put := doDAV(handler, http.MethodPut, "/dav/Dados/docs/a.txt", "x", nil); put.Code != http.StatusCreated {
		t.Fatalf("PUT in new dir: expected 201, got %d", put.Code)
	}

	move := doDAV(handler, "MOVE", "/dav/Dados/docs/a.txt", "", map[string]string{
		"Destination": "/dav/Dados/docs/b.txt",
	})
	if move.Code != http.StatusCreated && move.Code != http.StatusNoContent {
		t.Fatalf("MOVE: expected success, got %d (%s)", move.Code, move.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dataRoot, "docs", "b.txt")); err != nil {
		t.Fatalf("MOVE must relocate on disk: %v", err)
	}

	if response := doDAV(handler, http.MethodDelete, "/dav/Dados/docs", "", nil); response.Code != http.StatusNoContent {
		t.Fatalf("DELETE: expected 204, got %d", response.Code)
	}
	if _, err := os.Stat(filepath.Join(dataRoot, "docs")); !os.IsNotExist(err) {
		t.Fatalf("DELETE must remove from disk, stat err=%v", err)
	}
}

func TestMoveAcrossRootsIsRefused(t *testing.T) {
	handler, dataRoot, mediaRoot := setupDAV(t)

	if err := os.WriteFile(filepath.Join(dataRoot, "video.mp4"), []byte("v"), 0644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	move := doDAV(handler, "MOVE", "/dav/Dados/video.mp4", "", map[string]string{
		"Destination": "/dav/Midia/video.mp4",
	})
	if move.Code < http.StatusBadRequest {
		t.Fatalf("MOVE across roots must fail, got %d", move.Code)
	}
	if _, err := os.Stat(filepath.Join(dataRoot, "video.mp4")); err != nil {
		t.Fatalf("source must stay after refused move: %v", err)
	}
	if _, err := os.Stat(filepath.Join(mediaRoot, "video.mp4")); !os.IsNotExist(err) {
		t.Fatalf("destination must not exist, stat err=%v", err)
	}
}

func TestTrashDirIsHiddenAndUnreachable(t *testing.T) {
	handler, dataRoot, _ := setupDAV(t)

	trashDir := filepath.Join(dataRoot, trashDirName)
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		t.Fatalf("mkdir trash: %v", err)
	}
	if err := os.WriteFile(filepath.Join(trashDir, "secreto.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("seed trash file: %v", err)
	}

	listing := doDAV(handler, "PROPFIND", "/dav/Dados/", "", map[string]string{"Depth": "1"})
	if listing.Code != http.StatusMultiStatus {
		t.Fatalf("PROPFIND: expected 207, got %d", listing.Code)
	}
	if strings.Contains(listing.Body.String(), trashDirName) {
		t.Fatalf("trash dir must not appear in listings: %s", listing.Body.String())
	}

	direct := doDAV(handler, http.MethodGet, "/dav/Dados/"+trashDirName+"/secreto.txt", "", nil)
	if direct.Code != http.StatusNotFound {
		t.Fatalf("direct trash access must 404, got %d", direct.Code)
	}
}

func TestLevelZeroAndRootEntriesAreReadOnly(t *testing.T) {
	handler, dataRoot, _ := setupDAV(t)

	if put := doDAV(handler, http.MethodPut, "/dav/topo.txt", "x", nil); put.Code < http.StatusBadRequest {
		t.Fatalf("PUT at level zero must fail, got %d", put.Code)
	}
	if mkcol := doDAV(handler, "MKCOL", "/dav/NovaRaiz", "", nil); mkcol.Code < http.StatusBadRequest {
		t.Fatalf("MKCOL at level zero must fail, got %d", mkcol.Code)
	}
	if del := doDAV(handler, http.MethodDelete, "/dav/Dados", "", nil); del.Code < http.StatusBadRequest {
		t.Fatalf("DELETE of a root entry must fail, got %d", del.Code)
	}
	if _, err := os.Stat(dataRoot); err != nil {
		t.Fatalf("root dir must survive: %v", err)
	}

	unknown := doDAV(handler, http.MethodGet, "/dav/Inexistente/arquivo.txt", "", nil)
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown root label must 404, got %d", unknown.Code)
	}
}

func TestStatOfRootEntryUsesLabel(t *testing.T) {
	handler, _, _ := setupDAV(t)

	response := doDAV(handler, "PROPFIND", "/dav/Dados", "", map[string]string{"Depth": "0"})
	if response.Code != http.StatusMultiStatus {
		t.Fatalf("PROPFIND root entry: expected 207, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "Dados") {
		t.Fatalf("expected label in PROPFIND, got %s", response.Body.String())
	}
}

func TestRootsFSUnitEdges(t *testing.T) {
	handler, dataRoot, _ := setupDAV(t)
	_ = handler
	fsys := rootsFS{}
	ctx := t.Context()

	// Rename involving the virtual level zero or a root entry is refused.
	if err := fsys.Rename(ctx, "/", "/Dados/x"); !os.IsPermission(err) {
		t.Fatalf("rename of level zero: expected permission error, got %v", err)
	}
	if err := fsys.Rename(ctx, "/Dados/x", "/"); !os.IsPermission(err) {
		t.Fatalf("rename to level zero: expected permission error, got %v", err)
	}
	if err := fsys.Rename(ctx, "/Dados", "/Dados/renomeada"); !os.IsPermission(err) {
		t.Fatalf("rename of root entry: expected permission error, got %v", err)
	}
	if err := fsys.Rename(ctx, "/Inexistente/a", "/Dados/a"); !os.IsNotExist(err) {
		t.Fatalf("rename from unknown root: expected not-exist, got %v", err)
	}
	if err := fsys.Rename(ctx, "/Dados/a", "/Inexistente/a"); !os.IsNotExist(err) {
		t.Fatalf("rename to unknown root: expected not-exist, got %v", err)
	}

	// RemoveAll and Mkdir edge cases.
	if err := fsys.RemoveAll(ctx, "/"); !os.IsPermission(err) {
		t.Fatalf("remove level zero: expected permission error, got %v", err)
	}
	if err := fsys.RemoveAll(ctx, "/Inexistente"); !os.IsNotExist(err) {
		t.Fatalf("remove unknown root: expected not-exist, got %v", err)
	}
	if err := fsys.Mkdir(ctx, "/Dados/"+trashDirName+"/x", 0755); !os.IsNotExist(err) {
		t.Fatalf("mkdir inside trash: expected not-exist, got %v", err)
	}

	// Stat of level zero is the virtual directory.
	info, err := fsys.Stat(ctx, "/")
	if err != nil || !info.IsDir() || info.Name() != "/" {
		t.Fatalf("stat level zero: info=%+v err=%v", info, err)
	}
	if info.Size() != 0 || info.Mode()&os.ModeDir == 0 || info.Sys() != nil {
		t.Fatalf("virtual dir info inconsistent: %+v", info)
	}

	if err := os.WriteFile(filepath.Join(dataRoot, "f.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := fsys.Stat(ctx, "/Dados/"+trashDirName); !os.IsNotExist(err) {
		t.Fatalf("stat of trash: expected not-exist, got %v", err)
	}
}

func TestVirtualRootDirBehavior(t *testing.T) {
	setupDAV(t)

	dir := newVirtualRootDir()
	if err := dir.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := dir.Read(nil); err != os.ErrInvalid {
		t.Fatalf("Read: expected ErrInvalid, got %v", err)
	}
	if _, err := dir.Write(nil); err != os.ErrPermission {
		t.Fatalf("Write: expected ErrPermission, got %v", err)
	}
	if _, err := dir.Seek(0, io.SeekStart); err != os.ErrInvalid {
		t.Fatalf("Seek: expected ErrInvalid, got %v", err)
	}
	info, err := dir.Stat()
	if err != nil || !info.IsDir() {
		t.Fatalf("Stat: info=%+v err=%v", info, err)
	}

	// Paged listing: one entry at a time, then EOF.
	first, err := dir.Readdir(1)
	if err != nil || len(first) != 1 || first[0].Name() != "Dados" {
		t.Fatalf("first page: %+v err=%v", first, err)
	}
	second, err := dir.Readdir(1)
	if err != nil || len(second) != 1 || second[0].Name() != "Midia" {
		t.Fatalf("second page: %+v err=%v", second, err)
	}
	if _, err := dir.Readdir(1); err != io.EOF {
		t.Fatalf("exhausted paging: expected EOF, got %v", err)
	}

	// Readdir(-1) drains everything that remains.
	drain := newVirtualRootDir()
	all, err := drain.Readdir(-1)
	if err != nil || len(all) != 2 {
		t.Fatalf("drain: %+v err=%v", all, err)
	}
}
