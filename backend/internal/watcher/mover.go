package watcher

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func MoveToLibrary(file ScannedFile, libraryPath string) (string, error) {
	if err := os.MkdirAll(libraryPath, 0755); err != nil {
		return "", fmt.Errorf("create library path: %w", err)
	}

	baseName := filepath.Base(file.SourcePath)
	targetPath, err := resolveFileConflict(filepath.Join(libraryPath, baseName))
	if err != nil {
		return "", err
	}

	if err := moveFile(file.SourcePath, targetPath); err != nil {
		return "", err
	}

	return targetPath, nil
}

func resolveFileConflict(targetPath string) (string, error) {
	dir := filepath.Dir(targetPath)
	name := filepath.Base(targetPath)
	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]

	candidate := targetPath
	for idx := 1; ; idx++ {
		_, err := os.Stat(candidate)
		if errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
		if err != nil {
			return "", fmt.Errorf("resolve conflict stat: %w", err)
		}
		candidate = filepath.Join(dir, base+"_"+strconv.Itoa(idx)+ext)
	}
}

func moveFile(sourcePath string, targetPath string) error {
	if err := os.Rename(sourcePath, targetPath); err == nil {
		return nil
	}

	in, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	out, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create target file: %w", err)
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("copy file: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close target file: %w", err)
	}

	if err := os.Remove(sourcePath); err != nil {
		return fmt.Errorf("remove source file: %w", err)
	}

	return nil
}
