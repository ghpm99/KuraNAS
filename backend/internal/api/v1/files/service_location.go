package files

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nas-go/api/internal/roots"
)

func (s *Service) GetFileLocation(id int) (FileLocationDto, error) {
	file, err := s.GetFileById(id)
	if err != nil {
		return FileLocationDto{}, err
	}
	if file.DeletedAt.HasValue {
		return FileLocationDto{}, sql.ErrNoRows
	}

	diskPath := file.ResolveContentPath()
	return FileLocationDto{
		FileID:          file.ID,
		Tier:            file.Tier,
		LogicalPath:     roots.ToRelativePath(file.Path),
		DiskPath:        diskPath,
		LogicalDiskPath: file.Path,
		RootLabel:       owningRootLabel(file.Path),
		ExistsOnDisk:    isPresentOnDisk(diskPath),
	}, nil
}

func (s *Service) GetActiveFileByDiskPath(diskPath string) (FileDto, error) {
	normalizedPath := normalizeDiskPath(diskPath)
	model, found, err := s.Repository.GetActiveFileByPathOrPhysicalPath(normalizedPath)
	if err != nil {
		return FileDto{}, fmt.Errorf("error fetching file by disk path: %w", err)
	}
	if !found {
		return FileDto{}, sql.ErrNoRows
	}
	return model.ToDto()
}

func normalizeDiskPath(diskPath string) string {
	trimmedPath := strings.TrimSpace(diskPath)
	if trimmedPath == "" {
		return ""
	}
	return filepath.Clean(trimmedPath)
}

func owningRootLabel(absolutePath string) string {
	owner, isOwned := roots.OwnerOf(absolutePath)
	if !isOwned {
		return ""
	}
	return owner.Label
}

func isPresentOnDisk(diskPath string) bool {
	_, err := os.Stat(diskPath)
	return err == nil
}
