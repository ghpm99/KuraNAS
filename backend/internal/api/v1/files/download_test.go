package files

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type downloadServiceStub struct {
	ServiceInterface
	filesById map[int]FileDto
}

func (stub *downloadServiceStub) GetFileById(id int) (FileDto, error) {
	file, exists := stub.filesById[id]
	if !exists {
		return FileDto{}, sql.ErrNoRows
	}
	return file, nil
}

func (stub *downloadServiceStub) GetFilesByPathPrefix(prefix string, page int, pageSize int) (utils.PaginationResponse[FileDto], error) {
	descendants := make([]FileDto, 0)
	for _, file := range stub.filesById {
		if strings.HasPrefix(file.Path, prefix) {
			descendants = append(descendants, file)
		}
	}
	sort.Slice(descendants, func(left int, right int) bool { return descendants[left].Path < descendants[right].Path })
	return utils.PaginationResponse[FileDto]{Items: descendants}, nil
}

func newDownloadRouter(stub *downloadServiceStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(stub, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.GET("/files/blob/:id", handler.GetBlobFileHandler)
	router.GET("/files/download/:id", handler.DownloadFileHandler)
	router.GET("/files/download-zip", handler.DownloadZipHandler)
	return router
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func performGet(router *gin.Engine, url string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, url, nil)
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func readZipEntries(t *testing.T, body []byte) map[string]string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("invalid zip: %v", err)
	}
	entries := map[string]string{}
	for _, entry := range reader.File {
		opened, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, _ := io.ReadAll(opened)
		opened.Close()
		entries[entry.Name] = string(content)
	}
	return entries
}

func TestBlobServesRangeRequest(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "song.mp3")
	writeTestFile(t, filePath, "0123456789")
	router := newDownloadRouter(&downloadServiceStub{filesById: map[int]FileDto{
		1: {ID: 1, Name: "song.mp3", Path: filePath, Format: ".mp3", Type: File},
	}})

	full := performGet(router, "/files/blob/1", nil)
	if full.Code != http.StatusOK || full.Body.String() != "0123456789" {
		t.Fatalf("expected full body, got %d %q", full.Code, full.Body.String())
	}
	if full.Header().Get("Last-Modified") == "" || full.Header().Get("Accept-Ranges") != "bytes" {
		t.Fatalf("expected Last-Modified and Accept-Ranges, got %v", full.Header())
	}

	partial := performGet(router, "/files/blob/1", map[string]string{"Range": "bytes=2-5"})
	if partial.Code != http.StatusPartialContent || partial.Body.String() != "2345" {
		t.Fatalf("expected 206 with 2345, got %d %q", partial.Code, partial.Body.String())
	}
	if partial.Header().Get("Content-Range") != "bytes 2-5/10" {
		t.Fatalf("unexpected Content-Range %q", partial.Header().Get("Content-Range"))
	}
	if strings.Contains(partial.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("blob must not be an attachment")
	}
}

func TestBlobMissingFileOnDiskReturnsNotFound(t *testing.T) {
	router := newDownloadRouter(&downloadServiceStub{filesById: map[int]FileDto{
		1: {ID: 1, Name: "gone.txt", Path: filepath.Join(t.TempDir(), "gone.txt"), Format: ".txt", Type: File},
	}})

	if recorder := performGet(router, "/files/blob/1", nil); recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
	if recorder := performGet(router, "/files/blob/2", nil); recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown id, got %d", recorder.Code)
	}
}

func TestBlobServesColdFileFromPhysicalPath(t *testing.T) {
	coldPath := filepath.Join(t.TempDir(), "cold.bin")
	writeTestFile(t, coldPath, "cold-bytes")
	router := newDownloadRouter(&downloadServiceStub{filesById: map[int]FileDto{
		1: {ID: 1, Name: "cold.txt", Path: "/logical/does/not/exist/cold.txt", PhysicalPath: coldPath, Format: ".txt", Type: File},
	}})

	recorder := performGet(router, "/files/blob/1", nil)
	if recorder.Code != http.StatusOK || recorder.Body.String() != "cold-bytes" {
		t.Fatalf("expected cold bytes, got %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestDownloadFileSendsAttachmentWithEncodedName(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "relatório.txt")
	writeTestFile(t, filePath, "conteudo")
	router := newDownloadRouter(&downloadServiceStub{filesById: map[int]FileDto{
		1: {ID: 1, Name: "relatório.txt", Path: filePath, Format: ".txt", Type: File},
	}})

	recorder := performGet(router, "/files/download/1", nil)
	if recorder.Code != http.StatusOK || recorder.Body.String() != "conteudo" {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
	disposition, params, err := mime.ParseMediaType(recorder.Header().Get("Content-Disposition"))
	if err != nil || disposition != "attachment" || params["filename"] != "relatório.txt" {
		t.Fatalf("unexpected disposition %q (%v, %v)", recorder.Header().Get("Content-Disposition"), params, err)
	}
	if !strings.Contains(recorder.Header().Get("Content-Disposition"), "filename*=") {
		t.Fatalf("expected RFC 5987 filename*, got %q", recorder.Header().Get("Content-Disposition"))
	}
}

func TestDownloadFolderStreamsZipWithRelativeEntries(t *testing.T) {
	root := t.TempDir()
	folderPath := filepath.Join(root, "docs")
	hotPath := filepath.Join(folderPath, "a.txt")
	nestedPath := filepath.Join(folderPath, "sub", "b.txt")
	coldPath := filepath.Join(root, "cold-store", "c.bin")
	writeTestFile(t, hotPath, "hot")
	writeTestFile(t, nestedPath, "nested")
	writeTestFile(t, coldPath, "cold")
	router := newDownloadRouter(&downloadServiceStub{filesById: map[int]FileDto{
		1: {ID: 1, Name: "docs", Path: folderPath, Type: Directory},
		2: {ID: 2, Name: "a.txt", Path: hotPath, Type: File},
		3: {ID: 3, Name: "b.txt", Path: nestedPath, Type: File},
		4: {ID: 4, Name: "c.txt", Path: filepath.Join(folderPath, "c.txt"), PhysicalPath: coldPath, Type: File},
		5: {ID: 5, Name: "missing.txt", Path: filepath.Join(folderPath, "missing.txt"), Type: File},
		6: {ID: 6, Name: "other.txt", Path: filepath.Join(root, "docs-other", "x.txt"), Type: File},
	}})

	recorder := performGet(router, "/files/download/1", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	_, params, _ := mime.ParseMediaType(recorder.Header().Get("Content-Disposition"))
	if params["filename"] != "docs.zip" {
		t.Fatalf("expected docs.zip, got %q", params["filename"])
	}
	entries := readZipEntries(t, recorder.Body.Bytes())
	expected := map[string]string{"a.txt": "hot", "sub/b.txt": "nested", "c.txt": "cold"}
	if len(entries) != len(expected) {
		t.Fatalf("unexpected entries %v", entries)
	}
	for name, content := range expected {
		if entries[name] != content {
			t.Fatalf("entry %q expected %q got %q", name, content, entries[name])
		}
	}
}

func TestDownloadZipCombinesFilesAndFolders(t *testing.T) {
	root := t.TempDir()
	loosePath := filepath.Join(root, "loose.txt")
	folderPath := filepath.Join(root, "folder")
	insidePath := filepath.Join(folderPath, "inside.txt")
	otherLoosePath := filepath.Join(root, "elsewhere", "loose.txt")
	writeTestFile(t, loosePath, "loose")
	writeTestFile(t, insidePath, "inside")
	writeTestFile(t, otherLoosePath, "other")
	router := newDownloadRouter(&downloadServiceStub{filesById: map[int]FileDto{
		1: {ID: 1, Name: "loose.txt", Path: loosePath, Type: File},
		2: {ID: 2, Name: "folder", Path: folderPath, Type: Directory},
		3: {ID: 3, Name: "inside.txt", Path: insidePath, Type: File},
		4: {ID: 4, Name: "loose.txt", Path: otherLoosePath, Type: File},
	}})

	recorder := performGet(router, "/files/download-zip?ids=1,2&ids=4", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	_, params, _ := mime.ParseMediaType(recorder.Header().Get("Content-Disposition"))
	if !strings.HasPrefix(params["filename"], "kuranas-") || !strings.HasSuffix(params["filename"], ".zip") {
		t.Fatalf("unexpected archive name %q", params["filename"])
	}
	entries := readZipEntries(t, recorder.Body.Bytes())
	expected := map[string]string{"loose.txt": "loose", "folder/inside.txt": "inside", "loose (2).txt": "other"}
	if len(entries) != len(expected) {
		t.Fatalf("unexpected entries %v", entries)
	}
	for name, content := range expected {
		if entries[name] != content {
			t.Fatalf("entry %q expected %q got %q", name, content, entries[name])
		}
	}
}

func TestDownloadZipRejectsInvalidSelections(t *testing.T) {
	router := newDownloadRouter(&downloadServiceStub{filesById: map[int]FileDto{}})

	cases := []struct {
		name string
		url  string
		code int
	}{
		{"no ids", "/files/download-zip", http.StatusBadRequest},
		{"malformed id", "/files/download-zip?ids=abc", http.StatusBadRequest},
		{"unknown id", "/files/download-zip?ids=7", http.StatusNotFound},
		{"over the cap", "/files/download-zip?ids=" + joinSequentialIds(maxZipDownloadIds+1), http.StatusBadRequest},
	}
	for _, testCase := range cases {
		if recorder := performGet(router, testCase.url, nil); recorder.Code != testCase.code {
			t.Fatalf("%s: expected %d got %d", testCase.name, testCase.code, recorder.Code)
		}
	}
}

func joinSequentialIds(count int) string {
	ids := make([]string, 0, count)
	for id := 1; id <= count; id++ {
		ids = append(ids, strconv.Itoa(id))
	}
	return strings.Join(ids, ",")
}
