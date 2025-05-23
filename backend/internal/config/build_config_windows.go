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
		return fmt.Sprintf("%s/kuranas/icons/", os.Getenv("ProgramFiles"))
	case "TranslationsPath":
		return fmt.Sprintf("%s/kuranas/translations/", os.Getenv("ProgramFiles"))
	default:
		return ""
	}
}
