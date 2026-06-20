package ingest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	notifications "nas-go/api/internal/api/v1/notifications"
	"nas-go/api/pkg/applog"
)

func TestCompareCalVer(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"2024.08.06", "2024.08.06", 0},
		{"2024.08.06", "2024.09.01", -1},
		{"2024.09.01", "2024.08.06", 1},
		{"2023.12.30", "2024.01.01", -1},
		{"2024.08", "2024.08.06", -1}, // fewer segments, older
		{"dev", "2024.08.06", -1},     // non-numeric is oldest
	}
	for _, tc := range cases {
		if got := compareCalVer(tc.a, tc.b); got != tc.want {
			t.Errorf("compareCalVer(%q,%q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestParseSha256Sums(t *testing.T) {
	text := "aaa111  yt-dlp\nbbb222  yt-dlp_linux\nmalformed line\n"
	if hash, ok := parseSha256Sums(text, "yt-dlp_linux"); !ok || hash != "bbb222" {
		t.Fatalf("expected bbb222, got %q ok=%v", hash, ok)
	}
	if _, ok := parseSha256Sums(text, "missing"); ok {
		t.Fatal("expected not found for missing asset")
	}
}

func TestAssetURL(t *testing.T) {
	release := ghRelease{Assets: []ghAsset{{Name: "yt-dlp_linux", BrowserDownloadURL: "http://x/bin"}}}
	if got := assetURL(release, "yt-dlp_linux"); got != "http://x/bin" {
		t.Fatalf("got %q", got)
	}
	if got := assetURL(release, "nope"); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestSha256File(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f")
	content := []byte("hello yt-dlp")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	want := hex.EncodeToString(sum[:])
	got, err := sha256File(path)
	if err != nil || got != want {
		t.Fatalf("sha256File = %q, %v; want %q", got, err, want)
	}
}

func TestInstallVerifiedBinaryBacksUpExisting(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "yt-dlp")
	if err := os.WriteFile(dst, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), "new")
	if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := installVerifiedBinary(src, dst); err != nil {
		t.Fatalf("installVerifiedBinary: %v", err)
	}
	if data, _ := os.ReadFile(dst); string(data) != "new" {
		t.Fatalf("dst not replaced: %q", data)
	}
	if data, _ := os.ReadFile(dst + ".bak"); string(data) != "old" {
		t.Fatalf("backup not kept: %q", data)
	}
}

func newTestYtDlp(version string, versionErr error, release ghRelease, releaseErr error) *YtDlpService {
	return &YtDlpService{
		assetName:    "yt-dlp_linux",
		installPath:  func() string { return "" },
		execPath:     func() string { return "yt-dlp" },
		versionOf:    func(string) (string, error) { return version, versionErr },
		fetchRelease: func() (ghRelease, error) { return release, releaseErr },
		download:     func(string, string) error { return nil },
		fetchText:    func(string) (string, error) { return "", nil },
	}
}

func TestStatus(t *testing.T) {
	release := ghRelease{TagName: "2024.09.01", HTMLURL: "http://x", PublishedAt: "2024-09-01"}

	t.Run("update available", func(t *testing.T) {
		st := newTestYtDlp("2024.08.06", nil, release, nil).Status()
		if !st.Installed || !st.UpdateAvailable || st.LatestVersion != "2024.09.01" {
			t.Fatalf("unexpected: %+v", st)
		}
	})
	t.Run("up to date", func(t *testing.T) {
		st := newTestYtDlp("2024.09.01", nil, release, nil).Status()
		if st.UpdateAvailable {
			t.Fatalf("expected up to date: %+v", st)
		}
	})
	t.Run("not installed offers install", func(t *testing.T) {
		st := newTestYtDlp("", errors.New("missing"), release, nil).Status()
		if st.Installed || !st.UpdateAvailable {
			t.Fatalf("expected install offered: %+v", st)
		}
	})
	t.Run("github unreachable", func(t *testing.T) {
		st := newTestYtDlp("2024.08.06", nil, ghRelease{}, errors.New("offline")).Status()
		if st.UpdateAvailable || st.LatestVersion != "" {
			t.Fatalf("expected no update info: %+v", st)
		}
	})
}

func TestStatusLogsVersionFailureForensically(t *testing.T) {
	var buf bytes.Buffer
	applog.Setup(applog.Options{Writer: &buf, Level: slog.LevelInfo})

	t.Run("binary present but --version fails -> error", func(t *testing.T) {
		buf.Reset()
		existing := filepath.Join(t.TempDir(), "yt-dlp")
		if err := os.WriteFile(existing, []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
		svc := newTestYtDlp("", errors.New("boom"), ghRelease{}, errors.New("offline"))
		svc.execPath = func() string { return existing }
		svc.Status()
		if !strings.Contains(buf.String(), "binary present but --version failed") {
			t.Fatalf("expected present-but-failed log, got: %q", buf.String())
		}
	})

	t.Run("binary absent -> warn", func(t *testing.T) {
		buf.Reset()
		svc := newTestYtDlp("", errors.New("missing"), ghRelease{}, errors.New("offline"))
		svc.execPath = func() string { return filepath.Join(t.TempDir(), "absent") }
		svc.Status()
		if !strings.Contains(buf.String(), "binary not found") {
			t.Fatalf("expected not-found log, got: %q", buf.String())
		}
	})
}

func TestUpdateVerifiesChecksumAndInstalls(t *testing.T) {
	content := []byte("the-new-ytdlp-binary")
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])
	installTo := filepath.Join(t.TempDir(), "yt-dlp")

	release := ghRelease{
		TagName: "2024.09.01",
		Assets: []ghAsset{
			{Name: "yt-dlp_linux", BrowserDownloadURL: "http://x/bin"},
			{Name: "SHA2-256SUMS", BrowserDownloadURL: "http://x/sums"},
		},
	}
	svc := &YtDlpService{
		assetName:    "yt-dlp_linux",
		installPath:  func() string { return installTo },
		execPath:     func() string { return "yt-dlp" },
		versionOf:    func(string) (string, error) { return "2024.08.06", nil },
		fetchRelease: func() (ghRelease, error) { return release, nil },
		download:     func(_, dest string) error { return os.WriteFile(dest, content, 0o644) },
		fetchText:    func(string) (string, error) { return fmt.Sprintf("%s  yt-dlp_linux\n", hash), nil },
	}

	if err := svc.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if data, _ := os.ReadFile(installTo); string(data) != string(content) {
		t.Fatalf("installed content wrong: %q", data)
	}
}

func TestUpdateChecksumMismatchAborts(t *testing.T) {
	installTo := filepath.Join(t.TempDir(), "yt-dlp")
	release := ghRelease{Assets: []ghAsset{
		{Name: "yt-dlp_linux", BrowserDownloadURL: "http://x/bin"},
		{Name: "SHA2-256SUMS", BrowserDownloadURL: "http://x/sums"},
	}}
	svc := &YtDlpService{
		assetName:    "yt-dlp_linux",
		installPath:  func() string { return installTo },
		execPath:     func() string { return "yt-dlp" },
		versionOf:    func(string) (string, error) { return "2024.08.06", nil },
		fetchRelease: func() (ghRelease, error) { return release, nil },
		download:     func(_, dest string) error { return os.WriteFile(dest, []byte("tampered"), 0o644) },
		fetchText:    func(string) (string, error) { return "deadbeef  yt-dlp_linux\n", nil },
	}

	if err := svc.Update(); err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if _, err := os.Stat(installTo); !os.IsNotExist(err) {
		t.Fatal("nothing should have been installed on mismatch")
	}
}

func TestUpdateMissingAsset(t *testing.T) {
	svc := newTestYtDlp("2024.08.06", nil, ghRelease{TagName: "2024.09.01"}, nil)
	if err := svc.Update(); err == nil {
		t.Fatal("expected error when release has no binary asset")
	}
}

type fakeNotifier struct {
	called   bool
	groupKey string
	err      error
}

func (n *fakeNotifier) GroupOrCreate(dto notifications.CreateNotificationDto) (notifications.NotificationDto, error) {
	n.called = true
	n.groupKey = dto.GroupKey
	return notifications.NotificationDto{}, n.err
}

func TestCheckAndNotify(t *testing.T) {
	release := ghRelease{TagName: "2024.09.01"}

	t.Run("notifies when update available", func(t *testing.T) {
		notifier := &fakeNotifier{}
		ok, err := newTestYtDlp("2024.08.06", nil, release, nil).CheckAndNotify(notifier)
		if err != nil || !ok || !notifier.called {
			t.Fatalf("expected notification: ok=%v err=%v called=%v", ok, err, notifier.called)
		}
		if notifier.groupKey != ytDlpUpdateGroupKey+"-2024.09.01" {
			t.Fatalf("unexpected group key %q", notifier.groupKey)
		}
	})
	t.Run("silent when up to date", func(t *testing.T) {
		notifier := &fakeNotifier{}
		ok, err := newTestYtDlp("2024.09.01", nil, release, nil).CheckAndNotify(notifier)
		if err != nil || ok || notifier.called {
			t.Fatalf("expected no notification: ok=%v called=%v", ok, notifier.called)
		}
	})
	t.Run("nil notifier is a no-op", func(t *testing.T) {
		ok, err := newTestYtDlp("2024.08.06", nil, release, nil).CheckAndNotify(nil)
		if err != nil || ok {
			t.Fatalf("expected no-op: ok=%v err=%v", ok, err)
		}
	})
}
