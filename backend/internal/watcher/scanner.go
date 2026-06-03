package watcher

import (
	"io/fs"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/watchfolders"
	"nas-go/api/pkg/utils"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ScannedFile struct {
	SourcePath string
	Category   libraries.LibraryCategory
}

func ScanWatchFolder(watchFolder watchfolders.WatchFolderModel) ([]ScannedFile, error) {
	threshold := time.Time{}
	if watchFolder.LastScanAt != nil {
		threshold = *watchFolder.LastScanAt
	}

	files := make([]ScannedFile, 0)
	err := filepath.WalkDir(watchFolder.Path, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		if !threshold.IsZero() && !info.ModTime().After(threshold) {
			return nil
		}

		category, ok := classifyWatchFile(path)
		if !ok {
			return nil
		}

		files = append(files, ScannedFile{
			SourcePath: path,
			Category:   category,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i int, j int) bool {
		return files[i].SourcePath < files[j].SourcePath
	})

	return files, nil
}

func classifyWatchFile(path string) (libraries.LibraryCategory, bool) {
	formatType := utils.GetFormatTypeByExtension(strings.ToLower(filepath.Ext(path)))
	switch formatType.Type {
	case utils.FormatTypeImage:
		return libraries.LibraryCategoryImages, true
	case utils.FormatTypeAudio:
		return libraries.LibraryCategoryMusic, true
	case utils.FormatTypeVideo:
		return libraries.LibraryCategoryVideos, true
	case utils.FormatTypeDocument:
		return libraries.LibraryCategoryDocuments, true
	default:
		return "", false
	}
}
