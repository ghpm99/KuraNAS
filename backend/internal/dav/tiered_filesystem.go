package dav

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"nas-go/api/pkg/applog"

	"golang.org/x/net/webdav"
)

const writeIntentFlags = os.O_WRONLY | os.O_RDWR | os.O_CREATE | os.O_TRUNC | os.O_APPEND

func logicalDiskPath(dir webdav.Dir, rest string) string {
	return filepath.Join(string(dir), filepath.FromSlash(rest))
}

func (filesystem rootsFS) findColdFile(dir webdav.Dir, rest string) (ColdFile, bool) {
	if filesystem.coldFiles == nil || rest == "/" {
		return ColdFile{}, false
	}
	logicalPath := logicalDiskPath(dir, rest)
	coldFiles, err := filesystem.coldFiles.ListColdFilesByParentPath(filepath.Dir(logicalPath))
	if err != nil {
		applog.ErrorWithStack("dav: listing cold files failed", err, "logical_path", logicalPath)
		return ColdFile{}, false
	}
	for _, coldFile := range coldFiles {
		if filepath.Clean(coldFile.LogicalPath) == logicalPath {
			return coldFile, true
		}
	}
	return ColdFile{}, false
}

func (filesystem rootsFS) isColdOnly(ctx context.Context, dir webdav.Dir, rest string) bool {
	if _, statErr := dir.Stat(ctx, rest); !errors.Is(statErr, os.ErrNotExist) {
		return false
	}
	_, isCold := filesystem.findColdFile(dir, rest)
	return isCold
}

func (filesystem rootsFS) openWithTierFallback(ctx context.Context, dir webdav.Dir, rest string, flag int, perm os.FileMode) (webdav.File, error) {
	if flag&writeIntentFlags != 0 {
		if filesystem.isColdOnly(ctx, dir, rest) {
			return nil, os.ErrPermission
		}
		return dir.OpenFile(ctx, rest, flag, perm)
	}

	file, openErr := dir.OpenFile(ctx, rest, flag, perm)
	if !errors.Is(openErr, os.ErrNotExist) {
		return file, openErr
	}
	coldFile, isCold := filesystem.findColdFile(dir, rest)
	if !isCold {
		return nil, openErr
	}
	return openColdContent(coldFile)
}

func statColdFile(coldFile ColdFile) (os.FileInfo, error) {
	info, err := os.Stat(coldFile.PhysicalPath)
	if err != nil {
		return nil, err
	}
	return renamedFileInfo{FileInfo: info, name: coldFile.Name}, nil
}

func openColdContent(coldFile ColdFile) (webdav.File, error) {
	content, err := os.Open(coldFile.PhysicalPath)
	if err != nil {
		return nil, err
	}
	return &coldContentFile{File: content, logicalName: coldFile.Name}, nil
}

type coldContentFile struct {
	*os.File
	logicalName string
}

func (file *coldContentFile) Stat() (fs.FileInfo, error) {
	info, err := file.File.Stat()
	if err != nil {
		return nil, err
	}
	return renamedFileInfo{FileInfo: info, name: file.logicalName}, nil
}

func (file *coldContentFile) Write([]byte) (int, error) {
	return 0, os.ErrPermission
}

type tieredDirFile struct {
	webdav.File
	logicalDir    string
	coldFiles     ColdFileCatalog
	hotNames      map[string]bool
	pendingCold   []fs.FileInfo
	hasLoadedCold bool
}

func (file *tieredDirFile) Readdir(count int) ([]fs.FileInfo, error) {
	hotEntries, hotErr := file.File.Readdir(count)
	if hotErr != nil && !errors.Is(hotErr, io.EOF) {
		return hotEntries, hotErr
	}
	visibleHot := file.recordVisibleHotEntries(hotEntries)

	if count <= 0 {
		return append(visibleHot, file.takeColdEntries(-1)...), nil
	}
	if len(visibleHot) >= count {
		return visibleHot, nil
	}
	pageEntries := append(visibleHot, file.takeColdEntries(count-len(visibleHot))...)
	if len(pageEntries) == 0 {
		return nil, io.EOF
	}
	return pageEntries, nil
}

func (file *tieredDirFile) recordVisibleHotEntries(entries []fs.FileInfo) []fs.FileInfo {
	visible := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() == trashDirName {
			continue
		}
		file.hotNames[entry.Name()] = true
		visible = append(visible, entry)
	}
	return visible
}

func (file *tieredDirFile) takeColdEntries(limit int) []fs.FileInfo {
	file.loadColdEntries()
	if limit < 0 || limit > len(file.pendingCold) {
		limit = len(file.pendingCold)
	}
	taken := file.pendingCold[:limit]
	file.pendingCold = file.pendingCold[limit:]
	return taken
}

func (file *tieredDirFile) loadColdEntries() {
	if file.hasLoadedCold {
		return
	}
	file.hasLoadedCold = true
	if file.coldFiles == nil {
		return
	}
	coldFiles, err := file.coldFiles.ListColdFilesByParentPath(file.logicalDir)
	if err != nil {
		applog.ErrorWithStack("dav: listing cold files failed", err, "logical_dir", file.logicalDir)
		return
	}
	for _, coldFile := range coldFiles {
		if file.hotNames[coldFile.Name] {
			continue
		}
		info, statErr := statColdFile(coldFile)
		if statErr != nil {
			applog.Warn("dav: cold file unreachable, omitted from listing", "physical_path", coldFile.PhysicalPath, "error", statErr)
			continue
		}
		file.pendingCold = append(file.pendingCold, info)
	}
}
