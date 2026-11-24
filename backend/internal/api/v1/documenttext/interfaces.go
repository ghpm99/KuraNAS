package documenttext

import "nas-go/api/pkg/utils"

type RepositoryInterface interface {
	ListPendingIndexing(afterFileID int, limit int) ([]PendingDocument, error)
	UpsertDocumentText(documentText DocumentTextModel) error
	SearchDocuments(termPatterns utils.SearchTermPatterns, limit int, offset int) ([]DocumentMatchModel, error)
	FindDocumentForIndexing(fileID int) (PendingDocument, bool, error)
	FindIndexState(fileID int) (IndexStateModel, bool, error)
	ListDocuments(query DocumentListQuery) (utils.PaginationResponse[DocumentListItemModel], error)
	CountDocuments(filter DocumentFilter, folderPathPrefix string) (int, error)
	FindActiveFolderPath(folderID int) (path string, isFound bool, err error)
}

type ReindexLauncher interface {
	LaunchDocumentReindex(fileID int) (jobID int, err error)
}

type ServiceInterface interface {
	SearchDocuments(query string, page int, pageSize int) (utils.PaginationResponse[DocumentSearchResultDto], error)
	SearchTopDocuments(query string, limit int) ([]DocumentSearchResultDto, error)
	GetIndexState(fileID int) (IndexStateDto, error)
	ReindexDocument(fileID int) (int, error)
	ListDocuments(params DocumentListParams) (utils.PaginationResponse[DocumentItemDto], error)
	CountDocuments(filter DocumentFilter, folderID int) (DocumentCountDto, error)
}
