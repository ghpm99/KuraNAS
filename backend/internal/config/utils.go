package config

import (
	"os"
	"path/filepath"
)

func FilepathJoin(isPath bool, element ...string) string {
	filepathString := filepath.Join(element...)
	if isPath {
		filepathString = filepathString + string(os.PathSeparator)
	}
	return filepathString
}
