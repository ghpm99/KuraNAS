//go:build windows && !dev
// +build windows,!dev

package config

import (
	"fmt"
	"os"
)

func GetBuildConfig(key string) string {
	switch key {
	case "BuildVersion":
		return "windows"
	case "DbPath":
		return fmt.Sprintf("%s/kuranas/db.sqlite3", os.TempDir())
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
	default:
		return ""
	}
}
