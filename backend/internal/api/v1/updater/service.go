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
	"time"
)

const githubReleaseURL = "https://api.github.com/repos/ghpm99/KuraNAS/releases/latest"

var (
	fetchLatestReleaseFunc = fetchLatestRelease
	getAssetNameFunc       = getAssetName
	downloadFileFunc       = downloadFile
	extractAllFunc         = extractAll
	applyFullUpdateFunc    = applyFullUpdate
	osExecutableFunc       = os.Executable
	evalSymlinksFunc       = filepath.EvalSymlinks
	osRenameFunc           = os.Rename
	osRemoveAllFunc        = os.RemoveAll
	osMkdirAllFunc         = os.MkdirAll
	runtimeGOOS            = runtime.GOOS

	apiHTTPClient      = &http.Client{Timeout: 30 * time.Second}
	downloadHTTPClient = &http.Client{Timeout: 10 * time.Minute}
)

type Service struct {
	shutdownFn func()
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SetShutdownFn(fn func()) {
	s.shutdownFn = fn
}

func (s *Service) CheckForUpdate() (UpdateStatusDto, error) {
	release, err := fetchLatestReleaseFunc()
	if err != nil {
		return UpdateStatusDto{}, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(api.Version, "v")

	updateAvailable := compareVersions(currentVersion, latestVersion) < 0

	assetName := getAssetNameFunc()
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
	release, err := fetchLatestReleaseFunc()
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	assetName := getAssetNameFunc()
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

	zipPath := filepath.Join(tmpDir, assetName)
	if err := downloadFileFunc(downloadURL, zipPath); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to download update: %w", err)
	}

	info, err := os.Stat(zipPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}
	if info.Size() != expectedSize {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("downloaded file size mismatch: expected %d, got %d", expectedSize, info.Size())
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := extractAllFunc(zipPath, extractDir); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to extract update: %w", err)
	}

	if err := applyFullUpdateFunc(extractDir); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to apply update: %w", err)
	}

	os.RemoveAll(tmpDir)

	if s.shutdownFn != nil {
		time.AfterFunc(2*time.Second, s.shutdownFn)
	}

	return nil
}

func fetchLatestRelease() (GitHubRelease, error) {
	req, err := http.NewRequest("GET", githubReleaseURL, nil)
	if err != nil {
		return GitHubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "KuraNAS-Updater")

	resp, err := apiHTTPClient.Do(req)
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
	if runtimeGOOS == "windows" {
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
	resp, err := downloadHTTPClient.Get(url)
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

// extractAll extracts all files from the zip archive into destDir.
// It handles both Linux zips (with build/ prefix) and Windows zips (no prefix).
// It includes zip-slip protection.
func extractAll(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	prefix := detectZipPrefix(r.File)

	for _, f := range r.File {
		name := f.Name

		if prefix != "" {
			name = strings.TrimPrefix(name, prefix)
			if name == "" {
				continue
			}
		}

		targetPath := filepath.Join(destDir, filepath.FromSlash(name))

		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("zip-slip detected: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := osMkdirAllFunc(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			continue
		}

		if err := osMkdirAllFunc(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
		}

		if err := extractFile(f, targetPath); err != nil {
			return fmt.Errorf("failed to extract %s: %w", f.Name, err)
		}
	}

	return nil
}

// detectZipPrefix checks if all files in the zip share a common top-level directory prefix
// (e.g. "build/"). Returns the prefix to strip, or empty string if no common prefix.
func detectZipPrefix(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}

	var firstDir string
	for _, f := range files {
		parts := strings.SplitN(filepath.ToSlash(f.Name), "/", 2)
		if len(parts) < 2 {
			return ""
		}
		dir := parts[0] + "/"
		if firstDir == "" {
			firstDir = dir
		} else if dir != firstDir {
			return ""
		}
	}

	return firstDir
}

func extractFile(f *zip.File, targetPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// applyFullUpdate copies all extracted files to the installation directory.
// For the binary, it uses the rename-to-.old trick (works on Windows even while running).
// For other directories (dist, translations, scripts, icons), it replaces them in place.
// The scripts/.venv directory is preserved if it exists.
func applyFullUpdate(extractedDir string) error {
	installDir, err := getInstallDir()
	if err != nil {
		return err
	}

	if err := applyBinaryUpdate(extractedDir, installDir); err != nil {
		return err
	}

	assetDirs := []string{"dist", "icons", "translations", "scripts"}
	for _, dir := range assetDirs {
		srcDir := filepath.Join(extractedDir, dir)
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			continue
		}

		dstDir := filepath.Join(installDir, dir)

		if dir == "scripts" {
			if err := updateScriptsDir(srcDir, dstDir); err != nil {
				return fmt.Errorf("failed to update %s: %w", dir, err)
			}
			continue
		}

		if err := osRemoveAllFunc(dstDir); err != nil {
			return fmt.Errorf("failed to remove old %s: %w", dir, err)
		}
		if err := copyDir(srcDir, dstDir); err != nil {
			return fmt.Errorf("failed to copy new %s: %w", dir, err)
		}
	}

	return nil
}

func getInstallDir() (string, error) {
	exePath, err := osExecutableFunc()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	exePath, err = evalSymlinksFunc(exePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	return filepath.Dir(exePath), nil
}

func applyBinaryUpdate(extractedDir, installDir string) error {
	binaryName := "kuranas"
	if runtimeGOOS == "windows" {
		binaryName = "kuranas.exe"
	}

	newBinaryPath := filepath.Join(extractedDir, binaryName)
	if _, err := os.Stat(newBinaryPath); os.IsNotExist(err) {
		return fmt.Errorf("new binary not found in extracted files: %s", binaryName)
	}

	currentPath := filepath.Join(installDir, binaryName)
	oldPath := currentPath + ".old"

	// Remove previous .old if it exists
	os.Remove(oldPath)

	if err := osRenameFunc(currentPath, oldPath); err != nil {
		return fmt.Errorf("failed to rename current binary: %w", err)
	}

	if err := copyFile(newBinaryPath, currentPath); err != nil {
		osRenameFunc(oldPath, currentPath)
		return fmt.Errorf("failed to copy new binary: %w", err)
	}

	if runtimeGOOS != "windows" {
		if err := os.Chmod(currentPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	return nil
}

// updateScriptsDir replaces the scripts directory while preserving .venv
func updateScriptsDir(srcDir, dstDir string) error {
	venvDir := filepath.Join(dstDir, ".venv")
	venvBackup := filepath.Join(filepath.Dir(dstDir), ".venv-backup")

	hasVenv := false
	if _, err := os.Stat(venvDir); err == nil {
		hasVenv = true
		if err := osRenameFunc(venvDir, venvBackup); err != nil {
			return fmt.Errorf("failed to backup .venv: %w", err)
		}
	}

	if err := osRemoveAllFunc(dstDir); err != nil {
		if hasVenv {
			osRenameFunc(venvBackup, venvDir)
		}
		return fmt.Errorf("failed to remove old scripts: %w", err)
	}

	if err := copyDir(srcDir, dstDir); err != nil {
		if hasVenv {
			osMkdirAllFunc(dstDir, 0755)
			osRenameFunc(venvBackup, venvDir)
		}
		return fmt.Errorf("failed to copy new scripts: %w", err)
	}

	if hasVenv {
		if err := osRenameFunc(venvBackup, venvDir); err != nil {
			return fmt.Errorf("failed to restore .venv: %w", err)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return osMkdirAllFunc(targetPath, 0755)
		}

		return copyFile(path, targetPath)
	})
}
