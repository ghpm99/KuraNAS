package tiering

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	tieringengine "nas-go/api/internal/worker/tiering"
)

var (
	ErrFileNotFound         = errors.New("tiering: file not found")
	ErrFileAlreadyHot       = errors.New("tiering: file is already on the hot disk")
	ErrInsufficientHotSpace = errors.New("tiering: not enough free space on the hot disk")
	ErrColdCopyUnavailable  = errors.New("tiering: cold copy is unavailable")
)

func (s *Service) PromoteFile(fileID int) (FileLocationDto, error) {
	file, found, err := s.repository.GetFileById(fileID)
	if err != nil {
		return FileLocationDto{}, err
	}
	if !found {
		return FileLocationDto{}, ErrFileNotFound
	}
	if !file.isCold() {
		return FileLocationDto{}, ErrFileAlreadyHot
	}
	if _, statErr := os.Stat(file.PhysicalPath); statErr != nil {
		return FileLocationDto{}, fmt.Errorf("%w: %w", ErrColdCopyUnavailable, statErr)
	}
	if err := s.ensureHotSpaceFor(file); err != nil {
		return FileLocationDto{}, err
	}

	promotion := tieringengine.Promotion{
		FileID:   file.FileID,
		HotPath:  file.LogicalPath,
		ColdPath: file.PhysicalPath,
	}
	if err := s.promote(promotion, s.repository.SetPhysicalPath); err != nil {
		return FileLocationDto{}, fmt.Errorf("promoting file %d: %w", fileID, err)
	}

	return FileLocationDto{FileID: file.FileID, Tier: tierHot, DiskPath: file.LogicalPath}, nil
}

func (s *Service) ensureHotSpaceFor(file TieredFileModel) error {
	availableBytes, err := s.availableBytes(nearestExistingDirectory(filepath.Dir(file.LogicalPath)))
	if err != nil {
		return fmt.Errorf("reading hot disk free space: %w", err)
	}
	if availableBytes < file.Size {
		return fmt.Errorf("%w: need %d bytes, %d available", ErrInsufficientHotSpace, file.Size, availableBytes)
	}
	return nil
}

func nearestExistingDirectory(directory string) string {
	for {
		if _, err := os.Stat(directory); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return directory
		}
		directory = parent
	}
}
