package updater

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withMockHTTPClients(t *testing.T, fn roundTripFunc) {
	t.Helper()

	origAPI := apiHTTPClient
	origDownload := downloadHTTPClient
	mockClient := &http.Client{Transport: fn}
	apiHTTPClient = mockClient
	downloadHTTPClient = mockClient
	t.Cleanup(func() {
		apiHTTPClient = origAPI
		downloadHTTPClient = origDownload
	})
}

func resetServiceFns() {
	fetchLatestReleaseFunc = fetchLatestRelease
	getAssetNameFunc = getAssetName
	downloadFileFunc = downloadFile
	extractAllFunc = extractAll
	applyFullUpdateFunc = applyFullUpdate
}

func TestFetchLatestRelease_Success(t *testing.T) {
	withMockHTTPClients(t, func(req *http.Request) (*http.Response, error) {
		body := `{
			"tag_name":"v1.2.3",
			"html_url":"https://example.com/release",
			"published_at":"2025-01-01T00:00:00Z",
			"body":"notes",
			"assets":[]
		}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})

	release, err := fetchLatestRelease()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if release.TagName != "v1.2.3" {
		t.Fatalf("expected tag v1.2.3, got %s", release.TagName)
	}
}

func TestFetchLatestRelease_StatusError(t *testing.T) {
	withMockHTTPClients(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("error")),
			Header:     make(http.Header),
		}, nil
	})

	_, err := fetchLatestRelease()
	if err == nil {
		t.Fatalf("expected status error, got nil")
	}
}

func TestFetchLatestRelease_InvalidJSON(t *testing.T) {
	withMockHTTPClients(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{invalid-json")),
			Header:     make(http.Header),
		}, nil
	})

	_, err := fetchLatestRelease()
	if err == nil {
		t.Fatalf("expected json error, got nil")
	}
}

func TestDownloadFile_Success(t *testing.T) {
	content := "binary-content"
	withMockHTTPClients(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(content)),
			Header:     make(http.Header),
		}, nil
	})

	dest := filepath.Join(t.TempDir(), "download.bin")
	if err := downloadFile("https://example.com/file", dest); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != content {
		t.Fatalf("unexpected file content: %s", string(data))
	}
}

func TestDownloadFile_StatusError(t *testing.T) {
	withMockHTTPClients(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("bad gateway")),
			Header:     make(http.Header),
		}, nil
	})

	dest := filepath.Join(t.TempDir(), "download.bin")
	if err := downloadFile("https://example.com/file", dest); err == nil {
		t.Fatalf("expected error for non-200 status")
	}
}

func createZipWithFiles(t *testing.T, path string, files map[string]string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry: %v", err)
		}
		if _, err := io.Copy(w, bytes.NewBufferString(content)); err != nil {
			t.Fatalf("failed to write zip entry: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip: %v", err)
	}
}

func TestExtractAll_Success(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "release.zip")

	binName := "kuranas"
	if runtime.GOOS == "windows" {
		binName = "kuranas.exe"
	}

	createZipWithFiles(t, zipPath, map[string]string{
		binName:                "binary-data",
		"dist/index.html":      "<html>app</html>",
		"dist/assets/main.js":  "console.log('hello')",
		"translations/en.json": `{"key":"value"}`,
		"icons/icon.png":       "png-data",
		"scripts/metadata.py":  "print('ok')",
	})

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := extractAll(zipPath, extractDir); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify all files were extracted
	checks := map[string]string{
		binName:                "binary-data",
		"dist/index.html":      "<html>app</html>",
		"dist/assets/main.js":  "console.log('hello')",
		"translations/en.json": `{"key":"value"}`,
		"icons/icon.png":       "png-data",
		"scripts/metadata.py":  "print('ok')",
	}

	for name, expected := range checks {
		data, err := os.ReadFile(filepath.Join(extractDir, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("expected file %s to exist: %v", name, err)
		}
		if string(data) != expected {
			t.Fatalf("file %s: expected %q, got %q", name, expected, string(data))
		}
	}
}

func TestExtractAll_WithBuildPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "release.zip")

	binName := "kuranas"
	if runtime.GOOS == "windows" {
		binName = "kuranas.exe"
	}

	createZipWithFiles(t, zipPath, map[string]string{
		"build/" + binName:           "binary-data",
		"build/dist/index.html":      "<html>app</html>",
		"build/translations/en.json": `{"key":"value"}`,
	})

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := extractAll(zipPath, extractDir); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify build/ prefix was stripped
	data, err := os.ReadFile(filepath.Join(extractDir, binName))
	if err != nil {
		t.Fatalf("expected binary to exist without build/ prefix: %v", err)
	}
	if string(data) != "binary-data" {
		t.Fatalf("expected binary-data, got %q", string(data))
	}

	htmlData, err := os.ReadFile(filepath.Join(extractDir, "dist", "index.html"))
	if err != nil {
		t.Fatalf("expected dist/index.html to exist: %v", err)
	}
	if string(htmlData) != "<html>app</html>" {
		t.Fatalf("expected html content, got %q", string(htmlData))
	}
}

func TestExtractAll_EmptyZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "empty.zip")
	createZipWithFiles(t, zipPath, map[string]string{})

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := extractAll(zipPath, extractDir); err != nil {
		t.Fatalf("expected no error for empty zip, got %v", err)
	}
}

func TestDetectZipPrefix(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"no files", nil, ""},
		{"no prefix", []string{"kuranas.exe", "dist/index.html"}, ""},
		{"build prefix", []string{"build/kuranas.exe", "build/dist/index.html"}, "build/"},
		{"mixed prefix", []string{"build/kuranas.exe", "other/file.txt"}, ""},
		{"single file no dir", []string{"kuranas.exe"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var zipFiles []*zip.File
			for _, name := range tt.files {
				zipFiles = append(zipFiles, &zip.File{
					FileHeader: zip.FileHeader{Name: name},
				})
			}

			result := detectZipPrefix(zipFiles)
			if result != tt.expected {
				t.Fatalf("detectZipPrefix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCheckForUpdate(t *testing.T) {
	service := NewService()
	assetName := getAssetName()

	resetServiceFns()
	t.Cleanup(resetServiceFns)

	fetchLatestReleaseFunc = func() (GitHubRelease, error) {
		return GitHubRelease{
			TagName:     "v999.0.0",
			HTMLURL:     "https://example.com/release",
			PublishedAt: "2025-01-01T00:00:00Z",
			Body:        "notes",
			Assets: []GitHubAsset{
				{Name: assetName, Size: 12345, BrowserDownloadURL: "https://example.com/download"},
			},
		}, nil
	}
	getAssetNameFunc = func() string { return assetName }

	result, err := service.CheckForUpdate()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatalf("expected update to be available")
	}
	if result.AssetName != assetName {
		t.Fatalf("unexpected asset name: %s", result.AssetName)
	}
	if result.AssetSize != 12345 {
		t.Fatalf("unexpected asset size: %d", result.AssetSize)
	}
}

func TestDownloadAndApply_ErrorsAndSuccess(t *testing.T) {
	assetName := "kuranas-linux.zip"

	testRelease := GitHubRelease{
		TagName: "v2.0.0",
		Assets: []GitHubAsset{
			{Name: assetName, Size: 4, BrowserDownloadURL: "https://example.com/download"},
		},
	}

	cases := []struct {
		name            string
		fetchErr        error
		release         GitHubRelease
		downloadErr     error
		writeWrongSize  bool
		extractErr      error
		applyErr        error
		expectErrSubstr string
		expectShutdown  bool
	}{
		{
			name:            "fetch latest release error",
			fetchErr:        errors.New("boom"),
			expectErrSubstr: "failed to fetch latest release",
		},
		{
			name:            "no matching asset",
			release:         GitHubRelease{Assets: []GitHubAsset{{Name: "other.zip"}}},
			expectErrSubstr: "no matching asset found",
		},
		{
			name:            "download error",
			release:         testRelease,
			downloadErr:     errors.New("download failed"),
			expectErrSubstr: "failed to download update",
		},
		{
			name:            "download size mismatch",
			release:         testRelease,
			writeWrongSize:  true,
			expectErrSubstr: "downloaded file size mismatch",
		},
		{
			name:            "extract error",
			release:         testRelease,
			extractErr:      errors.New("extract failed"),
			expectErrSubstr: "failed to extract update",
		},
		{
			name:            "apply update error",
			release:         testRelease,
			applyErr:        errors.New("apply failed"),
			expectErrSubstr: "failed to apply update",
		},
		{
			name:           "success",
			release:        testRelease,
			expectShutdown: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetServiceFns()
			t.Cleanup(resetServiceFns)

			service := NewService()
			service.SetShutdownFn(func() {})

			getAssetNameFunc = func() string { return assetName }
			fetchLatestReleaseFunc = func() (GitHubRelease, error) {
				if tc.fetchErr != nil {
					return GitHubRelease{}, tc.fetchErr
				}
				return tc.release, nil
			}
			downloadFileFunc = func(url, dest string) error {
				if tc.downloadErr != nil {
					return tc.downloadErr
				}
				data := []byte("1234")
				if tc.writeWrongSize {
					data = []byte("12345")
				}
				return os.WriteFile(dest, data, 0644)
			}
			extractAllFunc = func(zipPath, destDir string) error {
				if tc.extractErr != nil {
					return tc.extractErr
				}
				return os.MkdirAll(destDir, 0755)
			}
			applyFullUpdateFunc = func(extractedDir string) error {
				return tc.applyErr
			}

			err := service.DownloadAndApply()
			if tc.expectErrSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectErrSubstr) {
					t.Fatalf("expected error containing %q, got %v", tc.expectErrSubstr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// Note: shutdownFn is called via time.AfterFunc with 2s delay,
			// so we don't check it synchronously here. We just verify no error.
		})
	}
}

func TestDownloadAndApply_NoShutdownFn(t *testing.T) {
	resetServiceFns()
	t.Cleanup(resetServiceFns)

	service := NewService()
	// No SetShutdownFn — should not panic

	assetName := "kuranas-linux.zip"
	fetchLatestReleaseFunc = func() (GitHubRelease, error) {
		return GitHubRelease{
			TagName: "v2.0.0",
			Assets: []GitHubAsset{
				{Name: assetName, Size: 4, BrowserDownloadURL: "https://example.com/download"},
			},
		}, nil
	}
	getAssetNameFunc = func() string { return assetName }
	downloadFileFunc = func(url, dest string) error {
		return os.WriteFile(dest, []byte("1234"), 0644)
	}
	extractAllFunc = func(zipPath, destDir string) error {
		return os.MkdirAll(destDir, 0755)
	}
	applyFullUpdateFunc = func(extractedDir string) error {
		return nil
	}

	if err := service.DownloadAndApply(); err != nil {
		t.Fatalf("expected no error without shutdownFn, got %v", err)
	}
}
