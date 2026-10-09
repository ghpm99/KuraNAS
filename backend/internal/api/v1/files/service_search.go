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
	namePattern := utils.BuildContainsLikePattern(params.Query)

	if params.ParentID == 0 {
		models, err := s.Repository.SearchActiveFilesByName(namePattern, params.Page, params.PageSize)
		return s.toSearchDtoPage(models, err)
	}

	parentFolder, err := s.findActiveFolder(params.ParentID)
	if err != nil {
		return utils.PaginationResponse[FileDto]{}, err
	}

	if params.IsRecursive {
		models, searchErr := s.Repository.SearchActiveFilesByNameUnderPath(buildDescendantsPathPrefix(parentFolder.Path), namePattern, params.Page, params.PageSize)
		return s.toSearchDtoPage(models, searchErr)
	}

	models, err := s.Repository.SearchActiveChildrenByName(parentFolder.Path, namePattern, params.Page, params.PageSize)
	return s.toSearchDtoPage(models, err)
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
