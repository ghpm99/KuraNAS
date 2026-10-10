package video

import (
	"fmt"
	"path/filepath"
	"strings"

	"nas-go/api/internal/roots"
	"nas-go/api/pkg/utils"
)

func (s *Service) ListLibraryFolders(request LibraryFolderRequest) (utils.PaginationResponse[LibraryFolderDto], error) {
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
		folderModels, err = s.Repository.ListLibraryFolders(query)
		if err != nil {
			return utils.PaginationResponse[LibraryFolderDto]{}, err
		}
	}

	response := utils.PaginationResponse[LibraryFolderDto]{
		Items:      toLibraryFolderDtos(folderModels),
		Pagination: utils.Pagination{Page: request.Page, PageSize: request.PageSize},
	}
	response.UpdatePagination()
	return response, nil
}

func (s *Service) ListLibraryFolderVideos(request LibraryFolderVideosRequest) (utils.PaginationResponse[VideoFileDto], error) {
	videoModels, err := s.Repository.ListLibraryFolderVideos(
		roots.ToAbsolutePath(request.FolderPath),
		request.PageSize+1,
		utils.CalculateOffset(request.Page, request.PageSize),
	)
	if err != nil {
		return utils.PaginationResponse[VideoFileDto]{}, fmt.Errorf("ListLibraryFolderVideos: %w", err)
	}

	videoDtos := make([]VideoFileDto, 0, len(videoModels))
	for _, videoModel := range videoModels {
		videoDtos = append(videoDtos, videoModel.ToDto())
	}

	response := utils.PaginationResponse[VideoFileDto]{
		Items:      videoDtos,
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
			VideoCount:  folderModel.VideoCount,
			CoverFileID: folderModel.CoverFileID,
		})
	}
	return folderDtos
}
