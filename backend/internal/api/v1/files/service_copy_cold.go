package files

import (
	"os"
	"path/filepath"
	"strings"
)

const coldDescendantsPageSize = 500

func (s *Service) copyColdDescendants(sourceDirectoryPath string, destinationDirectoryPath string) error {
	directoryPrefix := sourceDirectoryPath + string(filepath.Separator)

	for page := 1; ; page++ {
		descendants, err := s.Repository.GetFilesByPathPrefix(directoryPrefix, page, coldDescendantsPageSize)
		if err != nil {
			return err
		}

		for _, descendant := range descendants.Items {
			if !isActiveColdFile(descendant) || !strings.HasPrefix(descendant.Path, directoryPrefix) {
				continue
			}
			relativePath := strings.TrimPrefix(descendant.Path, directoryPrefix)
			if err := copyColdFile(descendant.PhysicalPath.String, filepath.Join(destinationDirectoryPath, relativePath)); err != nil {
				return err
			}
		}

		if !descendants.Pagination.HasNext {
			return nil
		}
	}
}

func isActiveColdFile(file FileModel) bool {
	return file.Type == File && file.PhysicalPath.Valid && file.PhysicalPath.String != "" && !file.DeletedAt.Valid
}

func copyColdFile(coldPath string, destinationPath string) error {
	coldInfo, err := os.Stat(coldPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return err
	}
	return copyRegularFile(coldPath, destinationPath, coldInfo.Mode())
}
