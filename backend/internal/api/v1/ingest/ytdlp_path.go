package ingest

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"nas-go/api/internal/config"
)

// ytDlpExecExt is the executable extension for the running platform. It is the
// single source of truth for ".exe" so the downloaded asset name and the
// installed binary name can never disagree on it again (the Windows bug where
// the asset arrived as yt-dlp.exe but was installed/executed as bare yt-dlp).
func ytDlpExecExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// ytDlpAssetName is the yt-dlp GitHub release asset for the running platform:
// the standalone binary that needs no Python at runtime.
func ytDlpAssetName() string {
	if runtime.GOOS == "windows" {
		return "yt-dlp" + ytDlpExecExt()
	}
	return "yt-dlp_linux"
}

// ytDlpBinaryName is the on-disk name of the managed binary: bare "yt-dlp",
// carrying the platform extension so Windows can actually execute it.
func ytDlpBinaryName() string {
	return "yt-dlp" + ytDlpExecExt()
}

// managedYtDlpDir is the directory the app installs its own yt-dlp into. In a
// production build it sits next to the executable; in a dev build (go run) the
// executable lives in a throwaway temp dir, so build config pins it to a stable
// project-relative directory instead — otherwise every recompile/restart loses
// the downloaded binary and Status() reports "not installed".
//
// It is a var so tests can pin it to a temp dir without depending on the
// developer's real bin/.
var managedYtDlpDir = func() string {
	if dir := strings.TrimSpace(config.GetBuildConfig("YtDlpDir")); dir != "" {
		return dir
	}
	exe, err := os.Executable()
	if err != nil {
		return "bin"
	}
	return filepath.Join(filepath.Dir(exe), "bin")
}

// managedYtDlpPath is where the app installs and owns a yt-dlp binary when no
// explicit YTDLP_PATH is configured.
func managedYtDlpPath() string {
	return filepath.Join(managedYtDlpDir(), ytDlpBinaryName())
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
