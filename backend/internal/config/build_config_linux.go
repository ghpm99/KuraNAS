//go:build linux && !dev
// +build linux,!dev

package config

import (
	"os"
)

func GetBuildConfig(key string) string {
	switch key {
	case "BuildVersion":
		return "linux"
	case "DbPath":
		return FilepathJoin(false, os.TempDir(), "kuranas", "db.sqlite3")
	case "IconPath":
		return FilepathJoin(true, "etc", "kuranas", "icons")
	case "TranslationsPath":
		return FilepathJoin(true, "etc", "kuranas", "translations")
	case "EnvFilePath":
		return FilepathJoin(false, "etc", "kuranas", ".env")
	case "PythonScript":
		return FilepathJoin(false, "etc", "kuranas", "scripts", ".venv", "bin", "python")
	case "ScriptPath":
		return FilepathJoin(true, "etc", "kuranas", "scripts")
	default:
		return ""
	}
}
