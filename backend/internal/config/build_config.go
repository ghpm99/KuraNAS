//go:build dev
// +build dev

package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func FindProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func GetBuildConfig(key string) string {
	switch key {
	case "BuildVersion":
		return "dev"
	case "DbPath":
		return "db.sqlite3"
	case "IconPath":
		currentDir := FindProjectRoot()
		return FilepathJoin(true, currentDir, "icons")
	case "TranslationsPath":
		currentDir := FindProjectRoot()
		return FilepathJoin(true, currentDir, "translations")
	case "EnvFilePath":
		currentDir := FindProjectRoot()
		return FilepathJoin(false, currentDir, ".env")
	case "PythonScript":
		currentDir := FindProjectRoot()
		if runtime.GOOS == "windows" {
			return FilepathJoin(false, currentDir, "scripts", ".venv", "Scripts", "python.exe")
		}
		return FilepathJoin(false, currentDir, "scripts", ".venv", "bin", "python")
	case "ScriptPath":
		currentDir := FindProjectRoot()
		return FilepathJoin(true, currentDir, "scripts")
	case "ThumbnailPath":
		currentDir := FindProjectRoot()
		return FilepathJoin(true, currentDir, "thumbnails")
	default:
		return ""
	}
}
