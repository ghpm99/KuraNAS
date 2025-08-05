//go:build windows && !dev
// +build windows,!dev

package config

import (
	"os"
)

func GetBuildConfig(key string) string {
	switch key {
	case "BuildVersion":
		return "windows"
	case "DbPath":
		return FilepathJoin(false, os.TempDir(), "Kuranas", "db.sqlite3")
	case "IconPath":
		return FilepathJoin(true, os.Getenv("ProgramFiles"), "Kuranas", "icons")
	case "TranslationsPath":
		return FilepathJoin(true, os.Getenv("ProgramFiles"), "Kuranas", "translations")
	case "EnvFilePath":
		return FilepathJoin(false, os.Getenv("ProgramFiles"), "Kuranas", ".env")
	case "PythonScript":
		return FilepathJoin(false, os.Getenv("ProgramFiles"), "Kuranas", "scripts", ".venv", "Scripts", "python.exe")
	case "ScriptPath":
		return FilepathJoin(true, os.Getenv("ProgramFiles"), "Kuranas", "scripts")
	default:
		return ""
	}
}
