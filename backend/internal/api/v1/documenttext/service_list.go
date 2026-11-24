package documenttext

import (
	"errors"
	"path/filepath"
	"strings"

	"nas-go/api/internal/roots"
	"nas-go/api/pkg/utils"
)

var ErrFolderNotFound = errors.New("folder not found")

func buildFolderPathPrefix(folderPath string) string {
	separator := string(filepath.Separator)
	if strings.HasSuffix(folderPath, separator) {
		return folderPath
	}
	return folderPath + separator
}

func (service *Service) resolveFolderPrefix(folderID int) (string, error) {
	if folderID == 0 {
		return "", nil
	}
	folderPath, isFound, err := service.Repository.FindActiveFolderPath(folderID)
	if err != nil {
		return "", err
	}
	if !isFound {
		return "", ErrFolderNotFound
	}
	return buildFolderPathPrefix(folderPath), nil
}

// ListDocuments returns one page of the document library, or
// ErrFolderNotFound when params.FolderID is not an active folder.
func (service *Service) ListDocuments(params DocumentListParams) (utils.PaginationResponse[DocumentItemDto], error) {
	emptyPage := utils.PaginationResponse[DocumentItemDto]{
		Items:      []DocumentItemDto{},
		Pagination: utils.Pagination{Page: params.Page, PageSize: params.PageSize, HasPrev: params.Page > 1},
	}
	folderPrefix, err := service.resolveFolderPrefix(params.FolderID)
	if err != nil {
		return emptyPage, err
	}

	page, err := service.Repository.ListDocuments(DocumentListQuery{
		Filter:     params.Filter,
		FolderPath: folderPrefix,
		Sort:       params.Sort,
		Order:      params.Order,
		Page:       params.Page,
		PageSize:   params.PageSize,
	})
	if err != nil {
		return emptyPage, err
	}

	emptyPage.Items = mapDocumentItems(page.Items)
	emptyPage.Pagination = page.Pagination
	return emptyPage, nil
}

// CountDocuments returns how many documents match the filter, or
// ErrFolderNotFound when folderID is not an active folder.
func (service *Service) CountDocuments(filter DocumentFilter, folderID int) (DocumentCountDto, error) {
	folderPrefix, err := service.resolveFolderPrefix(folderID)
	if err != nil {
		return DocumentCountDto{}, err
	}
	total, err := service.Repository.CountDocuments(filter, folderPrefix)
	if err != nil {
		return DocumentCountDto{}, err
	}
	return DocumentCountDto{Total: total}, nil
}

func mapDocumentItems(models []DocumentListItemModel) []DocumentItemDto {
	items := make([]DocumentItemDto, 0, len(models))
	for _, model := range models {
		items = append(items, DocumentItemDto{
			ID:         model.ID,
			Name:       model.Name,
			Format:     model.Format,
			ParentPath: roots.ToRelativePath(model.ParentPath),
			Size:       model.Size,
			ModifiedAt: model.ModifiedAt,
			CreatedAt:  model.CreatedAt,
			Starred:    model.Starred,
			IndexState: model.IndexState,
		})
	}
	return items
}
