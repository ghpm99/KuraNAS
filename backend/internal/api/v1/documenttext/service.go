package documenttext

import (
	"errors"

	"nas-go/api/internal/roots"
	"nas-go/api/pkg/doctext"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
)

var (
	ErrDocumentNotFound     = errors.New("document not found")
	errReindexerUnavailable = errors.New("document reindexer unavailable")
)

var failureMessageKeyByReason = map[string]string{
	doctext.ErrorCodeBinary:        "ERROR_DOCUMENT_INDEX_BINARY",
	doctext.ErrorCodeTimeout:       "ERROR_DOCUMENT_INDEX_TIMEOUT",
	doctext.ErrorCodeUnreadable:    "ERROR_DOCUMENT_INDEX_UNREADABLE",
	doctext.ErrorCodeExtractFailed: "ERROR_DOCUMENT_INDEX_EXTRACT_FAILED",
}

type Service struct {
	Repository RepositoryInterface
	Reindexer  ReindexLauncher
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func NewServiceWithReindexer(repository RepositoryInterface, reindexer ReindexLauncher) ServiceInterface {
	return &Service{Repository: repository, Reindexer: reindexer}
}

// GetIndexState returns the text indexing state of a document, or
// ErrDocumentNotFound when the file is not an active document.
func (service *Service) GetIndexState(fileID int) (IndexStateDto, error) {
	indexState, isFound, err := service.Repository.FindIndexState(fileID)
	if err != nil {
		return IndexStateDto{}, err
	}
	if !isFound {
		return IndexStateDto{}, ErrDocumentNotFound
	}

	indexStateDto := IndexStateDto{FileID: fileID, State: indexState.State, IndexedBytes: indexState.IndexedBytes}
	if indexState.State == StateNameOnly || indexState.State == StateFailed {
		indexStateDto.Reason = indexState.FailureReason
	}
	if indexState.State == StateFailed {
		indexStateDto.Error = i18n.GetMessage(failureMessageKeyFor(indexState.FailureReason))
	}
	return indexStateDto, nil
}

func failureMessageKeyFor(failureReason string) string {
	messageKey, isKnownReason := failureMessageKeyByReason[failureReason]
	if !isKnownReason {
		return failureMessageKeyByReason[doctext.ErrorCodeExtractFailed]
	}
	return messageKey
}

// ReindexDocument queues the text indexing of one document and returns the job
// id, or ErrDocumentNotFound when the file is not an active document.
func (service *Service) ReindexDocument(fileID int) (int, error) {
	_, isFound, err := service.Repository.FindDocumentForIndexing(fileID)
	if err != nil {
		return 0, err
	}
	if !isFound {
		return 0, ErrDocumentNotFound
	}
	if service.Reindexer == nil {
		return 0, errReindexerUnavailable
	}
	return service.Reindexer.LaunchDocumentReindex(fileID)
}

// SearchDocuments returns one page of documents matching the query in name or
// content, or ErrFolderNotFound when params.FolderID is not an active folder.
func (service *Service) SearchDocuments(params DocumentSearchParams) (utils.PaginationResponse[DocumentSearchResultDto], error) {
	emptyPage := utils.PaginationResponse[DocumentSearchResultDto]{
		Items:      []DocumentSearchResultDto{},
		Pagination: utils.Pagination{Page: params.Page, PageSize: params.PageSize, HasPrev: params.Page > 1},
	}

	termPatterns, hasTerms := utils.BuildSearchTermPatterns(params.Query)
	if !hasTerms {
		return emptyPage, nil
	}
	folderPrefix, err := service.resolveFolderPrefix(params.FolderID)
	if err != nil {
		return emptyPage, err
	}

	terms := utils.SplitSearchTerms(params.Query)
	matches, err := service.Repository.SearchDocuments(DocumentSearchQuery{
		TermPatterns: termPatterns,
		Terms:        terms,
		Filter:       params.Filter,
		FolderPath:   folderPrefix,
		Limit:        params.PageSize + 1,
		Offset:       utils.CalculateOffset(params.Page, params.PageSize),
	})
	if err != nil {
		return emptyPage, err
	}

	hasNext := len(matches) > params.PageSize
	if hasNext {
		matches = matches[:params.PageSize]
	}

	emptyPage.Items = mapMatches(matches, terms)
	emptyPage.Pagination.HasNext = hasNext
	return emptyPage, nil
}

func (service *Service) SearchTopDocuments(query string, limit int) ([]DocumentSearchResultDto, error) {
	termPatterns, hasTerms := utils.BuildSearchTermPatterns(query)
	if !hasTerms || limit <= 0 {
		return []DocumentSearchResultDto{}, nil
	}

	terms := utils.SplitSearchTerms(query)
	matches, err := service.Repository.SearchDocuments(DocumentSearchQuery{TermPatterns: termPatterns, Terms: terms, Limit: limit})
	if err != nil {
		return nil, err
	}
	return mapMatches(matches, terms), nil
}

func mapMatches(matches []DocumentMatchModel, terms []string) []DocumentSearchResultDto {
	results := make([]DocumentSearchResultDto, 0, len(matches))
	for _, match := range matches {
		snippet := ""
		if match.HasContentMatch {
			snippet = BuildSnippet(match.ContentWindow, terms)
		}
		results = append(results, DocumentSearchResultDto{
			FileID:     match.FileID,
			Name:       match.Name,
			Path:       roots.ToRelativePath(match.Path),
			ParentPath: roots.ToRelativePath(match.ParentPath),
			Format:     match.Format,
			Size:       match.Size,
			UpdatedAt:  match.UpdatedAt,
			Snippet:    snippet,
			Page:       match.Page,
		})
	}
	return results
}
