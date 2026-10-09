package documenttext

import "nas-go/api/pkg/utils"

type RepositoryInterface interface {
	ListPendingIndexing(afterFileID int, limit int) ([]PendingDocument, error)
	UpsertDocumentText(documentText DocumentTextModel) error
	SearchDocuments(termPatterns utils.SearchTermPatterns, limit int, offset int) ([]DocumentMatchModel, error)
}

type ServiceInterface interface {
	SearchDocuments(query string, page int, pageSize int) (utils.PaginationResponse[DocumentSearchResultDto], error)
	SearchTopDocuments(query string, limit int) ([]DocumentSearchResultDto, error)
}
