//go:build dev
// +build dev

package config

import (
	"os"
)

func GetBuildConfig(key string) string {
	switch key {
	case "BuildVersion":
		return "dev"
	case "DbPath":
		return "db.sqlite3"
	case "IconPath":
		currentDir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return FilepathJoin(true, currentDir, "icons")
	case "TranslationsPath":
		currentDir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return FilepathJoin(true, currentDir, "translations")
	case "EnvFilePath":
		currentDir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return FilepathJoin(false, currentDir, ".env")
	case "PythonScript":
		currentDir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return FilepathJoin(false, currentDir, "scripts", ".venv", "bin", "python")
	case "ScriptPath":
		currentDir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return FilepathJoin(true, currentDir, "scripts")
	default:
		return ""
	}
}
