package ingest

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nas-go/api/internal/config"
)

func TestResolveYtDlpPaths(t *testing.T) {
	original := config.AppConfig.YtDlpPath
	t.Cleanup(func() { config.AppConfig.YtDlpPath = original })

	config.AppConfig.YtDlpPath = "/opt/yt-dlp"
	if got := resolveYtDlpInstallPath(); got != "/opt/yt-dlp" {
		t.Fatalf("install path: got %q", got)
	}
	if got := resolveYtDlpBinary(); got != "/opt/yt-dlp" {
		t.Fatalf("binary: got %q", got)
	}

	config.AppConfig.YtDlpPath = ""
	if got := resolveYtDlpInstallPath(); !strings.HasSuffix(got, filepath.Join("bin", "yt-dlp")) {
		t.Fatalf("managed install path: got %q", got)
	}
	// No managed binary on disk -> falls back to the PATH command.
	if got := resolveYtDlpBinary(); got != "yt-dlp" {
		t.Fatalf("expected PATH fallback, got %q", got)
	}
}

func TestYtDlpAssetNameAndConstructor(t *testing.T) {
	if ytDlpAssetName() == "" {
		t.Fatal("asset name must not be empty")
	}
	svc := NewYtDlpService()
	if svc == nil || svc.fetchRelease == nil || svc.download == nil || svc.versionOf == nil {
		t.Fatal("constructor left collaborators unset")
	}
}

func TestYtDlpVersion(t *testing.T) {
	// `echo --version` exits 0 with output — exercises the success path without
	// depending on a real yt-dlp install (the exact text is OS-dependent).
	if out, err := ytDlpVersion("echo"); err != nil || out == "" {
		t.Fatalf("ytDlpVersion(echo) = %q, %v", out, err)
	}
	if _, err := ytDlpVersion("kuranas-no-such-binary-xyz"); err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestHTTPHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/text":
			w.Write([]byte("hello"))
		case "/bin":
			w.Write([]byte("binarydata"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	text, err := httpGetText(server.URL + "/text")
	if err != nil || text != "hello" {
		t.Fatalf("httpGetText = %q, %v", text, err)
	}
	if _, err := httpGetText(server.URL + "/missing"); err == nil {
		t.Fatal("expected error on 404")
	}

	dest := filepath.Join(t.TempDir(), "out")
	if err := httpDownloadFile(server.URL+"/bin", dest); err != nil {
		t.Fatalf("httpDownloadFile: %v", err)
	}
	if data, _ := os.ReadFile(dest); string(data) != "binarydata" {
		t.Fatalf("downloaded content wrong: %q", data)
	}
	if err := httpDownloadFile(server.URL+"/missing", dest); err == nil {
		t.Fatal("expected error on 404 download")
	}
}

func TestFetchYtDlpRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"2024.09.01","html_url":"http://x","assets":[{"name":"yt-dlp_linux","browser_download_url":"http://x/bin","size":10}]}`))
	}))
	t.Cleanup(server.Close)

	original := ytDlpReleaseURL
	ytDlpReleaseURL = server.URL
	t.Cleanup(func() { ytDlpReleaseURL = original })

	release, err := fetchYtDlpRelease()
	if err != nil || release.TagName != "2024.09.01" || len(release.Assets) != 1 {
		t.Fatalf("fetchYtDlpRelease = %+v, %v", release, err)
	}
}
