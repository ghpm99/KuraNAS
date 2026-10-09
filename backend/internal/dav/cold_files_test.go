package dav

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/internal/roots"
)

type fakeColdFileCatalog struct {
	coldFilesByParent map[string][]ColdFile
	listError         error
}

func (catalog *fakeColdFileCatalog) ListColdFilesByParentPath(parentPath string) ([]ColdFile, error) {
	if catalog.listError != nil {
		return nil, catalog.listError
	}
	return catalog.coldFilesByParent[parentPath], nil
}

type coldFixture struct {
	handler       http.Handler
	hotRoot       string
	coldFilePath  string
	logicalPath   string
	coldFileBytes string
	catalog       *fakeColdFileCatalog
}

func setupColdDAV(t *testing.T) coldFixture {
	t.Helper()
	previousEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = previousEntryPoint
		roots.Reset()
	})
	config.AppConfig.EntryPoint = ""

	hotRoot := t.TempDir()
	coldDir := t.TempDir()
	roots.Set([]roots.Root{{ID: 1, Path: hotRoot, Label: "Dados", Enabled: true}})

	coldFilePath := filepath.Join(coldDir, "Dados", "relatorio.pdf")
	if err := os.MkdirAll(filepath.Dir(coldFilePath), 0o755); err != nil {
		t.Fatalf("failed to create cold dir: %v", err)
	}
	coldFileBytes := "bytes-frios"
	if err := os.WriteFile(coldFilePath, []byte(coldFileBytes), 0o644); err != nil {
		t.Fatalf("failed to write cold file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hotRoot, "quente.txt"), []byte("quente"), 0o644); err != nil {
		t.Fatalf("failed to write hot file: %v", err)
	}

	logicalPath := filepath.Join(hotRoot, "relatorio.pdf")
	catalog := &fakeColdFileCatalog{coldFilesByParent: map[string][]ColdFile{
		hotRoot: {{
			Name:         "relatorio.pdf",
			LogicalPath:  logicalPath,
			PhysicalPath: coldFilePath,
			Size:         int64(len(coldFileBytes)),
			ModTime:      time.Now(),
		}},
	}}
	return coldFixture{
		handler:       NewHandler(catalog),
		hotRoot:       hotRoot,
		coldFilePath:  coldFilePath,
		logicalPath:   logicalPath,
		coldFileBytes: coldFileBytes,
		catalog:       catalog,
	}
}

func TestColdFileAppearsInDirectoryListing(t *testing.T) {
	fixture := setupColdDAV(t)

	response := doDAV(fixture.handler, "PROPFIND", "/dav/Dados/", "", map[string]string{"Depth": "1"})

	if response.Code != http.StatusMultiStatus {
		t.Fatalf("expected 207, got %d (%s)", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, expectedName := range []string{"relatorio.pdf", "quente.txt"} {
		if !strings.Contains(body, expectedName) {
			t.Fatalf("expected %q in listing, got %s", expectedName, body)
		}
	}
}

func TestColdFileListingSkipsEntryWhenColdVolumeIsUnmounted(t *testing.T) {
	fixture := setupColdDAV(t)
	if err := os.Remove(fixture.coldFilePath); err != nil {
		t.Fatalf("failed to simulate unmounted cold volume: %v", err)
	}

	response := doDAV(fixture.handler, "PROPFIND", "/dav/Dados/", "", map[string]string{"Depth": "1"})

	if response.Code != http.StatusMultiStatus {
		t.Fatalf("listing must survive an unreachable cold file, got %d", response.Code)
	}
	if strings.Contains(response.Body.String(), "relatorio.pdf") {
		t.Fatalf("unreachable cold file must be omitted, got %s", response.Body.String())
	}
}

func TestColdFileListingDoesNotDuplicateStrayHotCopy(t *testing.T) {
	fixture := setupColdDAV(t)
	if err := os.WriteFile(fixture.logicalPath, []byte("copia-quente"), 0o644); err != nil {
		t.Fatalf("failed to write stray hot copy: %v", err)
	}

	response := doDAV(fixture.handler, "PROPFIND", "/dav/Dados/", "", map[string]string{"Depth": "1"})

	if occurrences := strings.Count(response.Body.String(), "<D:href>/dav/Dados/relatorio.pdf</D:href>"); occurrences != 1 {
		t.Fatalf("expected a single entry for the file, got %d in %s", occurrences, response.Body.String())
	}
}

func TestColdFileDownloadServesPhysicalBytes(t *testing.T) {
	fixture := setupColdDAV(t)

	response := doDAV(fixture.handler, http.MethodGet, "/dav/Dados/relatorio.pdf", "", nil)

	if response.Code != http.StatusOK || response.Body.String() != fixture.coldFileBytes {
		t.Fatalf("expected cold bytes, got %d %q", response.Code, response.Body.String())
	}
}

func TestColdFilePropfindReportsLogicalName(t *testing.T) {
	fixture := setupColdDAV(t)

	response := doDAV(fixture.handler, "PROPFIND", "/dav/Dados/relatorio.pdf", "", map[string]string{"Depth": "0"})

	if response.Code != http.StatusMultiStatus {
		t.Fatalf("expected 207, got %d (%s)", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "relatorio.pdf") {
		t.Fatalf("expected logical name, got %s", response.Body.String())
	}
}

func TestMissingFileWithoutColdEntryStaysNotFound(t *testing.T) {
	fixture := setupColdDAV(t)

	response := doDAV(fixture.handler, http.MethodGet, "/dav/Dados/inexistente.pdf", "", nil)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", response.Code)
	}
}

func TestCatalogFailureDegradesToHotOnlyListing(t *testing.T) {
	fixture := setupColdDAV(t)
	fixture.catalog.listError = errors.New("database down")

	listing := doDAV(fixture.handler, "PROPFIND", "/dav/Dados/", "", map[string]string{"Depth": "1"})
	download := doDAV(fixture.handler, http.MethodGet, "/dav/Dados/relatorio.pdf", "", nil)

	if listing.Code != http.StatusMultiStatus || !strings.Contains(listing.Body.String(), "quente.txt") {
		t.Fatalf("hot listing must survive catalog failure, got %d %s", listing.Code, listing.Body.String())
	}
	if download.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cold file when catalog fails, got %d", download.Code)
	}
}

func TestWritesToColdFileAreRejectedAndLeaveNoHotCopy(t *testing.T) {
	fixture := setupColdDAV(t)

	put := doDAV(fixture.handler, http.MethodPut, "/dav/Dados/relatorio.pdf", "novo", nil)

	if put.Code < 400 {
		t.Fatalf("expected PUT on a cold file to be rejected, got %d", put.Code)
	}
	if _, err := os.Stat(fixture.logicalPath); !os.IsNotExist(err) {
		t.Fatalf("PUT must not create a hot copy of a cold file, stat err: %v", err)
	}
	coldBytes, _ := os.ReadFile(fixture.coldFilePath)
	if string(coldBytes) != fixture.coldFileBytes {
		t.Fatalf("cold bytes must stay untouched, got %q", coldBytes)
	}
}

func TestDeleteAndMoveOfColdFileAreRejected(t *testing.T) {
	fixture := setupColdDAV(t)

	deleteResponse := doDAV(fixture.handler, http.MethodDelete, "/dav/Dados/relatorio.pdf", "", nil)
	moveResponse := doDAV(fixture.handler, "MOVE", "/dav/Dados/relatorio.pdf", "", map[string]string{
		"Destination": "/dav/Dados/outro.pdf",
	})

	if deleteResponse.Code < 400 {
		t.Fatalf("expected DELETE of a cold file to be rejected, got %d", deleteResponse.Code)
	}
	if moveResponse.Code < 400 || moveResponse.Code == http.StatusBadGateway {
		t.Fatalf("expected MOVE of a cold file to be rejected, got %d", moveResponse.Code)
	}
	if _, err := os.Stat(fixture.coldFilePath); err != nil {
		t.Fatalf("cold bytes must survive rejected operations: %v", err)
	}
}

func TestMkcolOverColdFileNameIsRejected(t *testing.T) {
	fixture := setupColdDAV(t)

	response := doDAV(fixture.handler, "MKCOL", "/dav/Dados/relatorio.pdf", "", nil)

	if response.Code < 400 {
		t.Fatalf("expected MKCOL over a cold file name to be rejected, got %d", response.Code)
	}
	if _, err := os.Stat(fixture.logicalPath); !os.IsNotExist(err) {
		t.Fatalf("MKCOL must not create a hot entry, stat err: %v", err)
	}
}

func TestCopyOfColdFileReadsPhysicalBytes(t *testing.T) {
	fixture := setupColdDAV(t)

	response := doDAV(fixture.handler, "COPY", "/dav/Dados/relatorio.pdf", "", map[string]string{
		"Destination": "/dav/Dados/copia.pdf",
	})

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", response.Code, response.Body.String())
	}
	copied, err := os.ReadFile(filepath.Join(fixture.hotRoot, "copia.pdf"))
	if err != nil || string(copied) != fixture.coldFileBytes {
		t.Fatalf("expected the copy to carry cold bytes, got %q err=%v", copied, err)
	}
}

func TestPagedReaddirIncludesColdEntries(t *testing.T) {
	fixture := setupColdDAV(t)
	filesystem := &rootsFS{coldFiles: fixture.catalog}

	directory, err := filesystem.OpenFile(t.Context(), "/Dados", os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("open directory failed: %v", err)
	}
	defer directory.Close()

	var names []string
	for {
		page, readErr := directory.Readdir(1)
		for _, entry := range page {
			names = append(names, entry.Name())
		}
		if readErr != nil {
			break
		}
	}

	if len(names) != 2 {
		t.Fatalf("expected hot and cold entries across pages, got %v", names)
	}
}
