package music

import (
	"os"
	"path/filepath"
	"strings"
)

var folderCoverFileNames = []string{"cover", "folder", "front", "albumart"}
var folderCoverExtensions = []string{".jpg", ".jpeg", ".png"}

func findFolderCover(audioPath string) ([]byte, bool) {
	directoryEntries, err := os.ReadDir(filepath.Dir(audioPath))
	if err != nil {
		return nil, false
	}
	coverEntries := make(map[string]string, len(directoryEntries))
	for _, directoryEntry := range directoryEntries {
		if directoryEntry.IsDir() {
			continue
		}
		coverEntries[strings.ToLower(directoryEntry.Name())] = directoryEntry.Name()
	}
	for _, baseName := range folderCoverFileNames {
		for _, extension := range folderCoverExtensions {
			actualName, isPresent := coverEntries[baseName+extension]
			if !isPresent {
				continue
			}
			coverBytes, err := os.ReadFile(filepath.Join(filepath.Dir(audioPath), actualName))
			if err == nil && len(coverBytes) > 0 {
				return coverBytes, true
			}
		}
	}
	return nil, false
}
