package updater

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"nas-go/api/api"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const githubReleaseURL = "https://api.github.com/repos/ghpm99/KuraNAS/releases/latest"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CheckForUpdate() (UpdateStatusDto, error) {
	release, err := fetchLatestRelease()
	if err != nil {
		return UpdateStatusDto{}, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(api.Version, "v")

	updateAvailable := compareVersions(currentVersion, latestVersion) < 0

	assetName := getAssetName()
	var assetSize int64
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			assetSize = asset.Size
			break
		}
	}

	return UpdateStatusDto{
		CurrentVersion:  api.Version,
		LatestVersion:   release.TagName,
		UpdateAvailable: updateAvailable,
		ReleaseURL:      release.HTMLURL,
		ReleaseDate:     release.PublishedAt,
		ReleaseNotes:    release.Body,
		AssetName:       assetName,
		AssetSize:       assetSize,
	}, nil
}

func (s *Service) DownloadAndApply() error {
	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	assetName := getAssetName()
	var downloadURL string
	var expectedSize int64
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			expectedSize = asset.Size
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no matching asset found for %s", assetName)
	}

	tmpDir := filepath.Join(os.TempDir(), "kuranas-update")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(downloadURL, zipPath); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	info, err := os.Stat(zipPath)
	if err != nil {
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}
	if info.Size() != expectedSize {
		return fmt.Errorf("downloaded file size mismatch: expected %d, got %d", expectedSize, info.Size())
	}

	binaryPath, err := extractBinary(zipPath, tmpDir)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	if err := applyUpdate(binaryPath); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	go restartProcess()

	return nil
}

func fetchLatestRelease() (GitHubRelease, error) {
	req, err := http.NewRequest("GET", githubReleaseURL, nil)
	if err != nil {
		return GitHubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "KuraNAS-Updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GitHubRelease{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GitHubRelease{}, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return GitHubRelease{}, err
	}

	return release, nil
}

func getAssetName() string {
	if runtime.GOOS == "windows" {
		return "kuranas-windows.zip"
	}
	return "kuranas-linux.zip"
}

// compareVersions compares two SemVer version strings (without "v" prefix).
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Non-numeric versions (e.g. "dev") are treated as 0.0.0.
func compareVersions(a, b string) int {
	aParts := parseSemVer(a)
	bParts := parseSemVer(b)

	for i := 0; i < 3; i++ {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	return 0
}

func parseSemVer(version string) [3]int {
	var parts [3]int
	segments := strings.SplitN(version, ".", 3)
	for i, seg := range segments {
		if i >= 3 {
			break
		}
		n, err := strconv.Atoi(seg)
		if err != nil {
			return [3]int{0, 0, 0}
		}
		parts[i] = n
	}
	return parts
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
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

func extractBinary(zipPath, destDir string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	binaryName := "kuranas"
	if runtime.GOOS == "windows" {
		binaryName = "kuranas.exe"
	}

	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if baseName == binaryName {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			extractedPath := filepath.Join(destDir, binaryName)
			out, err := os.Create(extractedPath)
			if err != nil {
				return "", err
			}
			defer out.Close()

			if _, err := io.Copy(out, rc); err != nil {
				return "", err
			}

			return extractedPath, nil
		}
	}

	return "", fmt.Errorf("binary %s not found in archive", binaryName)
}

func applyUpdate(newBinaryPath string) error {
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	oldPath := currentPath + ".old"
	if err := os.Rename(currentPath, oldPath); err != nil {
		return fmt.Errorf("failed to rename current binary: %w", err)
	}

	src, err := os.Open(newBinaryPath)
	if err != nil {
		os.Rename(oldPath, currentPath)
		return fmt.Errorf("failed to open new binary: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(currentPath)
	if err != nil {
		os.Rename(oldPath, currentPath)
		return fmt.Errorf("failed to create new binary: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(currentPath)
		os.Rename(oldPath, currentPath)
		return fmt.Errorf("failed to copy new binary: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(currentPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	return nil
}

func restartProcess() {
	execPath, err := os.Executable()
	if err != nil {
		return
	}

	if runtime.GOOS == "windows" {
		proc, err := os.StartProcess(execPath, os.Args, &os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		})
		if err != nil {
			return
		}
		proc.Release()
		os.Exit(0)
	} else {
		syscall.Exec(execPath, os.Args, os.Environ())
	}
}
