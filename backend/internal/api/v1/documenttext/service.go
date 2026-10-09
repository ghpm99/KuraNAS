package documenttext

import (
	"nas-go/api/internal/roots"
	"nas-go/api/pkg/utils"
)

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func (service *Service) SearchDocuments(query string, page int, pageSize int) (utils.PaginationResponse[DocumentSearchResultDto], error) {
	emptyPage := utils.PaginationResponse[DocumentSearchResultDto]{
		Items:      []DocumentSearchResultDto{},
		Pagination: utils.Pagination{Page: page, PageSize: pageSize, HasPrev: page > 1},
	}

	termPatterns, hasTerms := utils.BuildSearchTermPatterns(query)
	if !hasTerms {
		return emptyPage, nil
	}

	matches, err := service.Repository.SearchDocuments(termPatterns, pageSize+1, (page-1)*pageSize)
	if err != nil {
		return emptyPage, err
	}

	hasNext := len(matches) > pageSize
	if hasNext {
		matches = matches[:pageSize]
	}

	emptyPage.Items = mapMatches(matches, utils.SplitSearchTerms(query))
	emptyPage.Pagination.HasNext = hasNext
	return emptyPage, nil
}

func (service *Service) SearchTopDocuments(query string, limit int) ([]DocumentSearchResultDto, error) {
	termPatterns, hasTerms := utils.BuildSearchTermPatterns(query)
	if !hasTerms || limit <= 0 {
		return []DocumentSearchResultDto{}, nil
	}

	matches, err := service.Repository.SearchDocuments(termPatterns, limit, 0)
	if err != nil {
		return nil, err
	}
	return mapMatches(matches, utils.SplitSearchTerms(query)), nil
}

func mapMatches(matches []DocumentMatchModel, terms []string) []DocumentSearchResultDto {
	results := make([]DocumentSearchResultDto, 0, len(matches))
	for _, match := range matches {
		results = append(results, DocumentSearchResultDto{
			FileID:     match.FileID,
			Name:       match.Name,
			Path:       roots.ToRelativePath(match.Path),
			ParentPath: roots.ToRelativePath(match.ParentPath),
			Format:     match.Format,
			Size:       match.Size,
			UpdatedAt:  match.UpdatedAt,
			Snippet:    BuildSnippet(match.ExtractedText, terms),
		})
	}
	return results
}
