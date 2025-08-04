//go:build windows && !dev
// +build windows,!dev

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getSystemDriveRoot() string {
	drive := os.Getenv("SystemDrive")
	if !strings.HasSuffix(drive, string(os.PathSeparator)) {
		drive += string(os.PathSeparator)
	}
	return drive
}

func GetBuildConfig(key string) string {
	switch key {
	case "BuildVersion":
		return "windows"
	case "DbPath":
		return fmt.Sprintf("%s\\Kuranas\\db.sqlite3", os.TempDir())
	case "IconPath":
		return fmt.Sprintf("%s\\Kuranas\\icons\\", os.Getenv("ProgramFiles"))
	case "TranslationsPath":
		return fmt.Sprintf("%s\\Kuranas\\translations\\", os.Getenv("ProgramFiles"))
	case "EnvFilePath":
		return fmt.Sprintf("%s\\Kuranas\\.env", os.Getenv("ProgramFiles"))
	case "PythonScript":
		return filepath.Join(getSystemDriveRoot(), "Kuranas", "scripts", ".venv", "Scripts", "python.exe")
	case "ScriptPath":
		return filepath.Join(getSystemDriveRoot(), "Kuranas", "scripts") + string(os.PathSeparator)
	default:
		return ""
	}
}
