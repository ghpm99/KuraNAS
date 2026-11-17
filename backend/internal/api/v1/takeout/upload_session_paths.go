package takeout

import (
	"nas-go/api/internal/config"
	"path/filepath"
)

const uploadSessionsDirName = ".takeout_uploads"

// UploadSessionsRootDir returns the directory that holds one subdirectory per
// Takeout upload session.
func UploadSessionsRootDir() string {
	return filepath.Join(config.AppConfig.EntryPoint, uploadSessionsDirName)
}

// UploadSessionDirOfZip returns the session directory that directly contains
// zipPath. It reports false when zipPath is empty or does not live exactly one
// level inside the upload sessions area, so callers never delete anything else.
func UploadSessionDirOfZip(zipPath string) (string, bool) {
	if zipPath == "" || config.AppConfig.EntryPoint == "" {
		return "", false
	}
	sessionDir := filepath.Dir(filepath.Clean(zipPath))
	if filepath.Dir(sessionDir) != filepath.Clean(UploadSessionsRootDir()) {
		return "", false
	}
	return sessionDir, true
}
