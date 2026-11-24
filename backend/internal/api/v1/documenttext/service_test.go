package documenttext

import (
	"errors"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/utils"
)

type recordingRepository struct {
	document     PendingDocument
	isFound      bool
	indexState   IndexStateModel
	findErr      error
	matches      []DocumentMatchModel
	err          error
	gotSearch    DocumentSearchQuery
	searchCalled bool

	folderPath    string
	isFolderFound bool
	listedPage    utils.PaginationResponse[DocumentListItemModel]
	gotListQuery  DocumentListQuery
	gotCountPath  string
	total         int
}

func (repository *recordingRepository) ListDocuments(query DocumentListQuery) (utils.PaginationResponse[DocumentListItemModel], error) {
	repository.gotListQuery = query
	return repository.listedPage, repository.err
}

func (repository *recordingRepository) CountDocuments(_ DocumentFilter, folderPathPrefix string) (int, error) {
	repository.gotCountPath = folderPathPrefix
	return repository.total, repository.err
}

func (repository *recordingRepository) FindActiveFolderPath(int) (string, bool, error) {
	return repository.folderPath, repository.isFolderFound, repository.findErr
}

func (repository *recordingRepository) ListPendingIndexing(int, int) ([]PendingDocument, error) {
	return nil, nil
}

func (repository *recordingRepository) FindDocumentForIndexing(int) (PendingDocument, bool, error) {
	return repository.document, repository.isFound, repository.findErr
}

func (repository *recordingRepository) FindDocumentText(int) (DocumentTextContentModel, bool, error) {
	return DocumentTextContentModel{}, false, nil
}

func (repository *recordingRepository) FindIndexState(int) (IndexStateModel, bool, error) {
	return repository.indexState, repository.isFound, repository.findErr
}

func (repository *recordingRepository) UpsertDocumentText(DocumentTextModel) error { return nil }

func (repository *recordingRepository) SearchDocuments(query DocumentSearchQuery) ([]DocumentMatchModel, error) {
	repository.searchCalled = true
	repository.gotSearch = query
	return repository.matches, repository.err
}

func TestSearchDocumentsPaginatesWithLookAheadRow(t *testing.T) {
	repository := &recordingRepository{matches: []DocumentMatchModel{
		{FileID: 1, Name: "a.txt", Path: "/a.txt", ParentPath: "/", Format: ".txt", Size: 5, UpdatedAt: time.Unix(1, 0), HasContentMatch: true, ContentWindow: "um contrato aqui"},
		{FileID: 2, Name: "b.txt", Path: "/b.txt", ParentPath: "/", Format: ".txt", HasContentMatch: true, ContentWindow: "outro contrato"},
		{FileID: 3, Name: "c.txt", Path: "/c.txt", ParentPath: "/", Format: ".txt", HasContentMatch: true, ContentWindow: "terceiro contrato"},
	}}
	service := NewService(repository)

	page, err := service.SearchDocuments(DocumentSearchParams{Query: "Contrato", Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repository.gotSearch.Limit != 3 || repository.gotSearch.Offset != 2 {
		t.Fatalf("expected limit 3 offset 2, got %d/%d", repository.gotSearch.Limit, repository.gotSearch.Offset)
	}
	if len(page.Items) != 2 || !page.Pagination.HasNext || !page.Pagination.HasPrev || page.Pagination.Page != 2 || page.Pagination.PageSize != 2 {
		t.Fatalf("unexpected page %+v", page)
	}
	if page.Items[0].FileID != 1 || page.Items[0].Snippet != "um contrato aqui" || page.Items[0].Size != 5 {
		t.Fatalf("unexpected first item %+v", page.Items[0])
	}
}

func TestSearchDocumentsLastPageHasNoNext(t *testing.T) {
	repository := &recordingRepository{matches: []DocumentMatchModel{{FileID: 1, HasContentMatch: true, ContentWindow: "x termo"}}}

	page, err := NewService(repository).SearchDocuments(DocumentSearchParams{Query: "termo", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Pagination.HasNext || page.Pagination.HasPrev || len(page.Items) != 1 {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestSearchDocumentsBlankQueryDoesNotHitRepository(t *testing.T) {
	repository := &recordingRepository{}

	page, err := NewService(repository).SearchDocuments(DocumentSearchParams{Query: "   ", Page: 1, PageSize: 20})
	if err != nil || repository.searchCalled || page.Items == nil || len(page.Items) != 0 {
		t.Fatalf("unexpected result page=%+v err=%v called=%v", page, err, repository.searchCalled)
	}
}

func TestSearchDocumentsBuildsMultiWordPatterns(t *testing.T) {
	repository := &recordingRepository{}

	if _, err := NewService(repository).SearchDocuments(DocumentSearchParams{Query: "Contrato locação", Page: 1, PageSize: 20}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.gotSearch.TermPatterns.AllPatterns) != 2 || !strings.Contains(repository.gotSearch.TermPatterns.DrivingPattern, "contrato") {
		t.Fatalf("unexpected patterns %+v", repository.gotSearch.TermPatterns)
	}
}

func TestSearchDocumentsPropagatesRepositoryError(t *testing.T) {
	repository := &recordingRepository{err: errors.New("boom")}

	if _, err := NewService(repository).SearchDocuments(DocumentSearchParams{Query: "termo", Page: 1, PageSize: 20}); err == nil {
		t.Fatal("expected repository error")
	}
	if _, err := NewService(repository).SearchTopDocuments("termo", 5); err == nil {
		t.Fatal("expected repository error from top search")
	}
}

func TestSearchTopDocumentsUsesLimitWithoutLookAhead(t *testing.T) {
	repository := &recordingRepository{matches: []DocumentMatchModel{{FileID: 7, HasContentMatch: true, ContentWindow: "um termo"}}}
	service := NewService(repository)

	results, err := service.SearchTopDocuments("termo", 6)
	if err != nil || len(results) != 1 || results[0].FileID != 7 {
		t.Fatalf("unexpected results %+v err=%v", results, err)
	}
	if repository.gotSearch.Limit != 6 || repository.gotSearch.Offset != 0 {
		t.Fatalf("expected limit 6 offset 0, got %d/%d", repository.gotSearch.Limit, repository.gotSearch.Offset)
	}

	empty, err := service.SearchTopDocuments("", 6)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v err=%v", empty, err)
	}
}

func TestSearchDocumentsOmitsSnippetWhenMatchWasOnlyInName(t *testing.T) {
	page := 3
	repository := &recordingRepository{matches: []DocumentMatchModel{
		{FileID: 1, Name: "contrato.pdf", HasContentMatch: false},
		{FileID: 2, Name: "b.pdf", HasContentMatch: true, ContentWindow: "texto com contrato no meio", Page: &page},
	}}

	results, err := NewService(repository).SearchTopDocuments("contrato", 5)
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Snippet != "" || results[0].Page != nil {
		t.Fatalf("name-only match must have empty snippet and no page: %+v", results[0])
	}
	if results[1].Snippet != "texto com contrato no meio" || results[1].Page == nil || *results[1].Page != 3 {
		t.Fatalf("unexpected content match %+v", results[1])
	}
}

func TestSearchDocumentsResolvesFolderAndPassesFilter(t *testing.T) {
	repository := &recordingRepository{folderPath: "/data/docs", isFolderFound: true}
	filter := DocumentFilter{Format: GroupPDF, FavoritesOnly: true, IndexState: StateNameOnly}

	_, err := NewService(repository).SearchDocuments(DocumentSearchParams{Query: "termo", Filter: filter, FolderID: 4, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if repository.gotSearch.FolderPath != buildFolderPathPrefix("/data/docs") || repository.gotSearch.Filter != filter {
		t.Fatalf("unexpected search query %+v", repository.gotSearch)
	}
}

func TestSearchDocumentsUnknownFolderIsNotFound(t *testing.T) {
	repository := &recordingRepository{}

	_, err := NewService(repository).SearchDocuments(DocumentSearchParams{Query: "termo", FolderID: 4, Page: 1, PageSize: 10})
	if !errors.Is(err, ErrFolderNotFound) || repository.searchCalled {
		t.Fatalf("expected ErrFolderNotFound without search, got %v", err)
	}
}
