//go:build linux && !dev
// +build linux,!dev

package config

import (
	"fmt"
	"os"
)

func GetBuildConfig(key string) string {
	switch key {
	case "BuildVersion":
		return "linux"
	case "DbPath":
		return fmt.Sprintf("%s/kuranas/db.sqlite3", os.TempDir())
	case "IconPath":
		return "/etc/kuranas/icons/"
	case "TranslationsPath":
		return "/etc/kuranas/translations/"
	default:
		return ""
	}
}
