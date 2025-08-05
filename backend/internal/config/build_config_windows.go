//go:build windows && !dev
// +build windows,!dev

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetBuildConfig(key string) string {
	switch key {
	case "BuildVersion":
		return "windows"
	case "DbPath":
		return fmt.Sprintf(os.TempDir(), "Kuranas", "db.sqlite3")
	case "IconPath":
		return fmt.Sprintf(os.Getenv("ProgramFiles"), "Kuranas", "icons") + string(os.PathSeparator)
	case "TranslationsPath":
		return fmt.Sprintf(os.Getenv("ProgramFiles"), "Kuranas", "translations") + string(os.PathSeparator)
	case "EnvFilePath":
		return filepath.Join(os.Getenv("ProgramFiles"), "Kuranas", ".env")
	case "PythonScript":
		return filepath.Join(os.Getenv("ProgramFiles"), "Kuranas", "scripts", ".venv", "Scripts", "python.exe")
	case "ScriptPath":
		return filepath.Join(os.Getenv("ProgramFiles"), "Kuranas", "scripts") + string(os.PathSeparator)
	default:
		return ""
	}
}
