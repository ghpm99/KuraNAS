package ingest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	notifications "nas-go/api/internal/api/v1/notifications"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
)

const (
	ytDlpChecksumsAsset = "SHA2-256SUMS"
	ytDlpUpdateGroupKey = "ytdlp-update"
)

var (
	// ytDlpReleaseURL is a var (not const) so tests can point it at a stub.
	ytDlpReleaseURL     = "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest"
	ytDlpAPIClient      = &http.Client{Timeout: 30 * time.Second}
	ytDlpDownloadClient = &http.Client{Timeout: 10 * time.Minute}
)

// ghAsset / ghRelease are the slices of the GitHub releases API we consume.
type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type ghRelease struct {
	TagName     string    `json:"tag_name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt string    `json:"published_at"`
	Assets      []ghAsset `json:"assets"`
}

// Notifier is the slice of the notifications service the checker needs to raise
// an "update available" notice. Declared locally so this stays decoupled and
// mockable.
type Notifier interface {
	GroupOrCreate(dto notifications.CreateNotificationDto) (notifications.NotificationDto, error)
}

// YtDlpService manages the lifecycle of the yt-dlp binary: report its version
// against GitHub, and apply a verified manual update. Collaborators are struct
// fields so tests can swap the network/exec/filesystem out.
type YtDlpService struct {
	assetName    string
	installPath  func() string
	execPath     func() string
	versionOf    func(binary string) (string, error)
	fetchRelease func() (ghRelease, error)
	download     func(url, dest string) error
	fetchText    func(url string) (string, error)
}

func NewYtDlpService() *YtDlpService {
	return &YtDlpService{
		assetName:    ytDlpAssetName(),
		installPath:  resolveYtDlpInstallPath,
		execPath:     resolveYtDlpBinary,
		versionOf:    ytDlpVersion,
		fetchRelease: fetchYtDlpRelease,
		download:     httpDownloadFile,
		fetchText:    httpGetText,
	}
}

// Status never errors: an unreachable GitHub or a missing binary degrades to a
// partial status, exactly like the Ollama status probe. The --version failure
// is still recorded forensically — a binary in the wrong place, an ephemeral
// install path and a genuinely absent binary all collapse to CurrentVersion ==
// "" for the client, so the real cause only survives in the file log.
func (s *YtDlpService) Status() YtDlpStatusDto {
	execPath := s.execPath()
	current, err := s.versionOf(execPath)
	if err != nil {
		if _, statErr := os.Stat(execPath); statErr == nil {
			applog.Error("ytdlp: binary present but --version failed", "path", execPath, "error", err.Error())
		} else {
			applog.Warn("ytdlp: binary not found", "path", execPath, "error", err.Error())
		}
	}
	status := YtDlpStatusDto{Installed: current != "", CurrentVersion: current}

	release, err := s.fetchRelease()
	if err != nil {
		return status
	}
	status.LatestVersion = release.TagName
	status.ReleaseURL = release.HTMLURL
	status.ReleaseDate = release.PublishedAt
	if current == "" {
		status.UpdateAvailable = true // nothing installed yet — an install is offered
	} else {
		status.UpdateAvailable = compareCalVer(current, release.TagName) < 0
	}
	return status
}

// Update downloads the official release asset, verifies it against the
// published SHA2-256SUMS, and only then atomically swaps it into place (keeping
// the previous binary as .bak). A checksum mismatch aborts before any swap.
func (s *YtDlpService) Update() error {
	release, err := s.fetchRelease()
	if err != nil {
		return fmt.Errorf("ytdlp update: fetch release: %w", err)
	}
	binURL := assetURL(release, s.assetName)
	if binURL == "" {
		return fmt.Errorf("ytdlp update: release has no asset %q", s.assetName)
	}
	sumsURL := assetURL(release, ytDlpChecksumsAsset)
	if sumsURL == "" {
		return fmt.Errorf("ytdlp update: release has no %s", ytDlpChecksumsAsset)
	}

	tmpDir, err := os.MkdirTemp("", "kuranas-ytdlp-")
	if err != nil {
		return fmt.Errorf("ytdlp update: temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpBin := filepath.Join(tmpDir, s.assetName)
	if err := s.download(binURL, tmpBin); err != nil {
		return fmt.Errorf("ytdlp update: download binary: %w", err)
	}

	sums, err := s.fetchText(sumsURL)
	if err != nil {
		return fmt.Errorf("ytdlp update: download checksums: %w", err)
	}
	expected, ok := parseSha256Sums(sums, s.assetName)
	if !ok {
		return fmt.Errorf("ytdlp update: no checksum for %q", s.assetName)
	}
	actual, err := sha256File(tmpBin)
	if err != nil {
		return fmt.Errorf("ytdlp update: hash downloaded file: %w", err)
	}
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("ytdlp update: checksum mismatch (expected %s, got %s)", expected, actual)
	}

	return installVerifiedBinary(tmpBin, s.installPath())
}

// CheckAndNotify probes the status and, when an update is available, raises a
// single grouped notification (deduped per version). It never applies anything.
func (s *YtDlpService) CheckAndNotify(notifier Notifier) (bool, error) {
	if notifier == nil {
		return false, nil
	}
	status := s.Status()
	if !status.UpdateAvailable || status.LatestVersion == "" {
		return false, nil
	}

	message := i18n.Translate("YTDLP_INSTALL_NOTIF_MESSAGE", status.LatestVersion)
	if status.Installed {
		message = i18n.Translate("YTDLP_UPDATE_NOTIF_MESSAGE", status.LatestVersion, status.CurrentVersion)
	}

	_, err := notifier.GroupOrCreate(notifications.CreateNotificationDto{
		Type:     string(notifications.NotificationTypeWarning),
		Title:    i18n.GetMessage("YTDLP_UPDATE_NOTIF_TITLE"),
		Message:  message,
		GroupKey: ytDlpUpdateGroupKey + "-" + status.LatestVersion,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func assetURL(release ghRelease, name string) string {
	for _, asset := range release.Assets {
		if asset.Name == name {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

// parseSha256Sums reads the hex hash for the given filename out of a
// "<hash>  <name>" SHA2-256SUMS file.
func parseSha256Sums(text, name string) (string, bool) {
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) != 2 {
			continue
		}
		if fields[1] == name {
			return fields[0], true
		}
	}
	return "", false
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// installVerifiedBinary places src at dst (0755), backing up any existing file
// as dst.bak and restoring it if the copy fails.
func installVerifiedBinary(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("ytdlp update: create install dir: %w", err)
	}
	backup := dst + ".bak"
	if _, err := os.Stat(dst); err == nil {
		_ = os.Remove(backup)
		if err := os.Rename(dst, backup); err != nil {
			return fmt.Errorf("ytdlp update: backup existing binary: %w", err)
		}
	}
	if err := copyFileMode(src, dst, 0o755); err != nil {
		if _, statErr := os.Stat(backup); statErr == nil {
			_ = os.Rename(backup, dst)
		}
		return fmt.Errorf("ytdlp update: install binary: %w", err)
	}
	return nil
}

func copyFileMode(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Chmod(dst, mode)
}

// compareCalVer compares two dot-separated numeric versions (yt-dlp uses CalVer
// like 2024.08.06). A non-numeric segment collapses the version to the oldest.
// Returns -1 if a < b, 0 if equal, 1 if a > b.
func compareCalVer(a, b string) int {
	aParts := numericSegments(a)
	bParts := numericSegments(b)
	for i := 0; i < len(aParts) || i < len(bParts); i++ {
		var av, bv int
		if i < len(aParts) {
			av = aParts[i]
		}
		if i < len(bParts) {
			bv = bParts[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func numericSegments(version string) []int {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	segments := strings.Split(version, ".")
	parts := make([]int, len(segments))
	for i, seg := range segments {
		n, err := strconv.Atoi(seg)
		if err != nil {
			return []int{0}
		}
		parts[i] = n
	}
	return parts
}

// ytDlpVersion runs `yt-dlp --version`. A missing/broken binary yields "" so
// callers treat it as "not installed".
func ytDlpVersion(binary string) (string, error) {
	out, err := exec.Command(binary, "--version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func fetchYtDlpRelease() (ghRelease, error) {
	req, err := http.NewRequest(http.MethodGet, ytDlpReleaseURL, nil)
	if err != nil {
		return ghRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "KuraNAS")

	resp, err := ytDlpAPIClient.Do(req)
	if err != nil {
		return ghRelease{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ghRelease{}, fmt.Errorf("github returned status %d", resp.StatusCode)
	}

	var release ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ghRelease{}, err
	}
	return release, nil
}

func httpDownloadFile(url, dest string) error {
	resp, err := ytDlpDownloadClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func httpGetText(url string) (string, error) {
	resp, err := ytDlpAPIClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
