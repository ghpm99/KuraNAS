//go:build dev
// +build dev

package config

import (
	"fmt"
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
		return fmt.Sprintf("%s/icons/", currentDir)
	case "TranslationsPath":
		currentDir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%s/translations/", currentDir)
	case "EnvFilePath":
		currentDir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%s/.env", currentDir)
	case "PythonScript":
		currentDir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%s/scripts/.venv/bin/python", currentDir)
	case "ScriptPath":
		currentDir, err := os.Getwd()
		if err != nil {
			return ""
		}
		return fmt.Sprintf("%s/scripts/", currentDir)
	default:
		return ""
	}
}
