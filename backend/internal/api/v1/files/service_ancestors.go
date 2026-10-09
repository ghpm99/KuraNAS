package files

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"nas-go/api/internal/roots"
)

func (s *Service) GetFileAncestors(id int) ([]FileAncestorDto, error) {
	file, err := s.GetFileById(id)
	if err != nil {
		return nil, err
	}
	if file.DeletedAt.HasValue {
		return nil, sql.ErrNoRows
	}

	ancestorPaths := ancestorPathsFromRoot(file.Path)
	if len(ancestorPaths) == 0 {
		return []FileAncestorDto{}, nil
	}

	ancestorModels, err := s.Repository.GetActiveFilesByPaths(ancestorPaths)
	if err != nil {
		return nil, fmt.Errorf("error fetching file ancestors: %w", err)
	}
	return orderAncestorsByPath(ancestorModels, ancestorPaths), nil
}

func ancestorPathsFromRoot(filePath string) []string {
	cleanPath := filepath.Clean(filePath)
	rootPath, hasOwner := owningRootPath(cleanPath)

	reversedPaths := []string{}
	currentPath := cleanPath
	for {
		parentPath := filepath.Dir(currentPath)
		isTopOfFilesystem := parentPath == currentPath
		isAboveRoot := hasOwner && currentPath == rootPath
		if isTopOfFilesystem || isAboveRoot {
			break
		}
		reversedPaths = append(reversedPaths, parentPath)
		currentPath = parentPath
	}

	ancestorPaths := make([]string, 0, len(reversedPaths))
	for index := len(reversedPaths) - 1; index >= 0; index-- {
		ancestorPaths = append(ancestorPaths, reversedPaths[index])
	}
	return ancestorPaths
}

func owningRootPath(absolutePath string) (string, bool) {
	owner, isOwned := roots.OwnerOf(absolutePath)
	return owner.Path, isOwned
}

func orderAncestorsByPath(ancestorModels []FileModel, orderedPaths []string) []FileAncestorDto {
	modelsByPath := make(map[string]FileModel, len(ancestorModels))
	for _, model := range ancestorModels {
		modelsByPath[model.Path] = model
	}

	ancestors := make([]FileAncestorDto, 0, len(orderedPaths))
	for _, ancestorPath := range orderedPaths {
		model, isIndexed := modelsByPath[ancestorPath]
		if !isIndexed {
			continue
		}
		ancestors = append(ancestors, toAncestorDto(model))
	}
	return ancestors
}

func toAncestorDto(model FileModel) FileAncestorDto {
	name := model.Name
	if owner, isOwned := roots.OwnerOf(model.Path); isOwned && owner.Path == filepath.Clean(model.Path) && owner.Label != "" {
		name = owner.Label
	}
	return FileAncestorDto{
		ID:   model.ID,
		Name: name,
		Path: roots.ToRelativePath(model.Path),
		Type: model.Type,
	}
}
