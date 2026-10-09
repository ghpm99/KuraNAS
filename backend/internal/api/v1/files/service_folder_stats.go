package files

import (
	"database/sql"
	"fmt"
	"path/filepath"
)

func (s *Service) GetFolderStats(id int) (FolderStatsDto, error) {
	folder, err := s.GetFileById(id)
	if err != nil {
		return FolderStatsDto{}, err
	}
	if folder.DeletedAt.HasValue || folder.Type != Directory {
		return FolderStatsDto{}, sql.ErrNoRows
	}

	stats, err := s.Repository.GetFolderStats(folder.Path + string(filepath.Separator))
	if err != nil {
		return FolderStatsDto{}, fmt.Errorf("error fetching folder stats: %w", err)
	}
	return stats, nil
}
