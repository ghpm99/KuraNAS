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
	"time"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withMockHTTPClient(t *testing.T, fn roundTripFunc) {
	t.Helper()

	original := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: fn}
	t.Cleanup(func() {
		http.DefaultClient = original
	})
}

func resetServiceFns() {
	fetchLatestReleaseFunc = fetchLatestRelease
	getAssetNameFunc = getAssetName
	downloadFileFunc = downloadFile
	extractBinaryFunc = extractBinary
	applyUpdateFunc = applyUpdate
	restartProcessFunc = restartProcess
}

func TestFetchLatestRelease_Success(t *testing.T) {
	withMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
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
	withMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
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
	withMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
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
	withMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
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
	withMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
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

func TestExtractBinary_Success(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "release.zip")

	binName := "kuranas"
	if runtime.GOOS == "windows" {
		binName = "kuranas.exe"
	}

	createZipWithFiles(t, zipPath, map[string]string{
		"build/" + binName: "binary-data",
		"other/file.txt":   "ignore",
	})

	extracted, err := extractBinary(zipPath, tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filepath.Base(extracted) != binName {
		t.Fatalf("expected extracted binary %s, got %s", binName, extracted)
	}
	if _, err := os.Stat(extracted); err != nil {
		t.Fatalf("expected extracted file to exist: %v", err)
	}
}

func TestExtractBinary_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "release.zip")
	createZipWithFiles(t, zipPath, map[string]string{
		"readme.txt": "no binary here",
	})

	_, err := extractBinary(zipPath, tmpDir)
	if err == nil {
		t.Fatalf("expected not found error")
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
	service := NewService()
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
		expectRestart   bool
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
			expectErrSubstr: "failed to extract binary",
		},
		{
			name:            "apply update error",
			release:         testRelease,
			applyErr:        errors.New("apply failed"),
			expectErrSubstr: "failed to apply update",
		},
		{
			name:          "success",
			release:       testRelease,
			expectRestart: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetServiceFns()
			t.Cleanup(resetServiceFns)

			restartCalled := make(chan struct{}, 1)

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
			extractBinaryFunc = func(zipPath, destDir string) (string, error) {
				if tc.extractErr != nil {
					return "", tc.extractErr
				}
				bin := filepath.Join(destDir, "kuranas")
				if runtime.GOOS == "windows" {
					bin = filepath.Join(destDir, "kuranas.exe")
				}
				if err := os.WriteFile(bin, []byte("bin"), 0755); err != nil {
					return "", err
				}
				return bin, nil
			}
			applyUpdateFunc = func(newBinaryPath string) error {
				return tc.applyErr
			}
			restartProcessFunc = func() {
				select {
				case restartCalled <- struct{}{}:
				default:
				}
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

			if tc.expectRestart {
				select {
				case <-restartCalled:
				case <-time.After(100 * time.Millisecond):
					t.Fatalf("expected restartProcess to be called")
				}
			}
		})
	}
}
