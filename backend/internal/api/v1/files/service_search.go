package files

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"

	"nas-go/api/pkg/utils"
)

func buildDescendantsPathPrefix(folderPath string) string {
	separator := string(filepath.Separator)
	if strings.HasSuffix(folderPath, separator) {
		return folderPath
	}
	return folderPath + separator
}

func (s *Service) SearchFilesByName(params FileSearchParams) (utils.PaginationResponse[FileDto], error) {
	query := FileSearchQuery{
		Query:    params.Query,
		Filter:   params.Filter,
		Page:     params.Page,
		PageSize: params.PageSize,
	}

	if params.ParentID != 0 {
		parentFolder, err := s.findActiveFolder(params.ParentID)
		if err != nil {
			return utils.PaginationResponse[FileDto]{}, err
		}
		query.Scope, query.ScopePath = searchScopeOf(parentFolder, params.IsRecursive)
	}

	models, err := s.Repository.SearchActiveFiles(query)
	return s.toSearchDtoPage(models, err)
}

func searchScopeOf(parentFolder FileDto, isRecursive bool) (FileSearchScope, string) {
	if isRecursive {
		return SearchScopeDescendants, buildDescendantsPathPrefix(parentFolder.Path)
	}
	return SearchScopeChildren, parentFolder.Path
}

func (s *Service) findActiveFolder(folderID int) (FileDto, error) {
	folder, err := s.GetFileById(folderID)
	if errors.Is(err, sql.ErrNoRows) {
		return FileDto{}, ErrFileNotFound
	}
	if err != nil {
		return FileDto{}, err
	}
	if folder.DeletedAt.HasValue || folder.Type != Directory {
		return FileDto{}, ErrFileNotFound
	}
	return folder, nil
}

func (s *Service) toSearchDtoPage(models utils.PaginationResponse[FileModel], searchErr error) (utils.PaginationResponse[FileDto], error) {
	if searchErr != nil {
		return utils.PaginationResponse[FileDto]{}, searchErr
	}
	return s.toDtoPageWithCounts(models)
}
