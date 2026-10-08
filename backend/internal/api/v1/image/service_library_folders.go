package image

import (
	"fmt"
	"path/filepath"
	"strings"

	"nas-go/api/internal/roots"
	"nas-go/api/pkg/utils"
)

func (s *LibraryService) ListLibraryFolders(request LibraryFolderRequest) (utils.PaginationResponse[LibraryFolderDto], error) {
	separator := string(filepath.Separator)
	query := LibraryFolderQuery{
		Scopes:    buildLibraryFolderScopes(request.ParentPath, separator),
		Separator: separator,
		Limit:     request.PageSize + 1,
		Offset:    utils.CalculateOffset(request.Page, request.PageSize),
	}

	folderModels := []LibraryFolderModel{}
	if len(query.Scopes) > 0 {
		var err error
		folderModels, err = s.repository.ListLibraryFolders(query)
		if err != nil {
			return utils.PaginationResponse[LibraryFolderDto]{}, fmt.Errorf("ListLibraryFolders: %w", err)
		}
	}

	response := utils.PaginationResponse[LibraryFolderDto]{
		Items:      toLibraryFolderDtos(folderModels),
		Pagination: utils.Pagination{Page: request.Page, PageSize: request.PageSize},
	}
	response.UpdatePagination()
	return response, nil
}

func buildLibraryFolderScopes(parentPath string, separator string) []LibraryFolderScope {
	if parentPath != "" {
		return []LibraryFolderScope{{Prefix: withTrailingSeparator(roots.ToAbsolutePath(parentPath), separator)}}
	}

	enabledRoots := roots.Enabled()
	scopes := make([]LibraryFolderScope, 0, len(enabledRoots))
	for position, root := range enabledRoots {
		isPrimary := position == 0
		scopes = append(scopes, LibraryFolderScope{
			Prefix:      withTrailingSeparator(root.Path, separator),
			IsWholeRoot: !isPrimary,
			Label:       root.Label,
		})
	}
	return scopes
}

func withTrailingSeparator(path string, separator string) string {
	return strings.TrimSuffix(path, separator) + separator
}

func toLibraryFolderDtos(folderModels []LibraryFolderModel) []LibraryFolderDto {
	folderDtos := make([]LibraryFolderDto, 0, len(folderModels))
	for _, folderModel := range folderModels {
		folderDtos = append(folderDtos, LibraryFolderDto{
			Path:        roots.ToRelativePath(folderModel.Path),
			Name:        folderModel.Name,
			ImageCount:  folderModel.ImageCount,
			CoverFileID: folderModel.CoverFileID,
		})
	}
	return folderDtos
}
