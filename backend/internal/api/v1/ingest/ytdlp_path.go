package ingest

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"nas-go/api/internal/config"
)

// ytDlpAssetName is the yt-dlp GitHub release asset for the running platform:
// the standalone binary that needs no Python at runtime.
func ytDlpAssetName() string {
	if runtime.GOOS == "windows" {
		return "yt-dlp.exe"
	}
	return "yt-dlp_linux"
}

// managedYtDlpPath is where the app installs and owns a yt-dlp binary when no
// explicit YTDLP_PATH is configured: a writable dir next to the executable.
func managedYtDlpPath() string {
	exe, err := os.Executable()
	if err != nil {
		return filepath.Join("bin", "yt-dlp")
	}
	return filepath.Join(filepath.Dir(exe), "bin", "yt-dlp")
}

// resolveYtDlpInstallPath is where the updater writes the binary: the configured
// path if set, otherwise the app-managed path. It never falls back to a bare
// command name — you cannot install onto "yt-dlp" on PATH.
func resolveYtDlpInstallPath() string {
	if p := strings.TrimSpace(config.AppConfig.YtDlpPath); p != "" {
		return p
	}
	return managedYtDlpPath()
}

// resolveYtDlpBinary is what the fetch step executes: the configured path, else
// the app-managed path when it already exists, else the bare "yt-dlp" command
// resolved through PATH.
func resolveYtDlpBinary() string {
	if p := strings.TrimSpace(config.AppConfig.YtDlpPath); p != "" {
		return p
	}
	managed := managedYtDlpPath()
	if _, err := os.Stat(managed); err == nil {
		return managed
	}
	return "yt-dlp"
}
