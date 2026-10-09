package documenttext

import (
	"errors"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/utils"
)

type recordingRepository struct {
	matches      []DocumentMatchModel
	err          error
	gotPatterns  utils.SearchTermPatterns
	gotLimit     int
	gotOffset    int
	searchCalled bool
}

func (repository *recordingRepository) ListPendingIndexing(int, int) ([]PendingDocument, error) {
	return nil, nil
}

func (repository *recordingRepository) UpsertDocumentText(DocumentTextModel) error { return nil }

func (repository *recordingRepository) SearchDocuments(termPatterns utils.SearchTermPatterns, limit int, offset int) ([]DocumentMatchModel, error) {
	repository.searchCalled = true
	repository.gotPatterns = termPatterns
	repository.gotLimit = limit
	repository.gotOffset = offset
	return repository.matches, repository.err
}

func TestSearchDocumentsPaginatesWithLookAheadRow(t *testing.T) {
	repository := &recordingRepository{matches: []DocumentMatchModel{
		{FileID: 1, Name: "a.txt", Path: "/a.txt", ParentPath: "/", Format: ".txt", Size: 5, UpdatedAt: time.Unix(1, 0), ExtractedText: "um contrato aqui"},
		{FileID: 2, Name: "b.txt", Path: "/b.txt", ParentPath: "/", Format: ".txt", ExtractedText: "outro contrato"},
		{FileID: 3, Name: "c.txt", Path: "/c.txt", ParentPath: "/", Format: ".txt", ExtractedText: "terceiro contrato"},
	}}
	service := NewService(repository)

	page, err := service.SearchDocuments("Contrato", 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repository.gotLimit != 3 || repository.gotOffset != 2 {
		t.Fatalf("expected limit 3 offset 2, got %d/%d", repository.gotLimit, repository.gotOffset)
	}
	if len(page.Items) != 2 || !page.Pagination.HasNext || !page.Pagination.HasPrev || page.Pagination.Page != 2 || page.Pagination.PageSize != 2 {
		t.Fatalf("unexpected page %+v", page)
	}
	if page.Items[0].FileID != 1 || page.Items[0].Snippet != "um contrato aqui" || page.Items[0].Size != 5 {
		t.Fatalf("unexpected first item %+v", page.Items[0])
	}
}

func TestSearchDocumentsLastPageHasNoNext(t *testing.T) {
	repository := &recordingRepository{matches: []DocumentMatchModel{{FileID: 1, ExtractedText: "x termo"}}}

	page, err := NewService(repository).SearchDocuments("termo", 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Pagination.HasNext || page.Pagination.HasPrev || len(page.Items) != 1 {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestSearchDocumentsBlankQueryDoesNotHitRepository(t *testing.T) {
	repository := &recordingRepository{}

	page, err := NewService(repository).SearchDocuments("   ", 1, 20)
	if err != nil || repository.searchCalled || page.Items == nil || len(page.Items) != 0 {
		t.Fatalf("unexpected result page=%+v err=%v called=%v", page, err, repository.searchCalled)
	}
}

func TestSearchDocumentsBuildsMultiWordPatterns(t *testing.T) {
	repository := &recordingRepository{}

	if _, err := NewService(repository).SearchDocuments("Contrato locação", 1, 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.gotPatterns.AllPatterns) != 2 || !strings.Contains(repository.gotPatterns.DrivingPattern, "contrato") {
		t.Fatalf("unexpected patterns %+v", repository.gotPatterns)
	}
}

func TestSearchDocumentsPropagatesRepositoryError(t *testing.T) {
	repository := &recordingRepository{err: errors.New("boom")}

	if _, err := NewService(repository).SearchDocuments("termo", 1, 20); err == nil {
		t.Fatal("expected repository error")
	}
	if _, err := NewService(repository).SearchTopDocuments("termo", 5); err == nil {
		t.Fatal("expected repository error from top search")
	}
}

func TestSearchTopDocumentsUsesLimitWithoutLookAhead(t *testing.T) {
	repository := &recordingRepository{matches: []DocumentMatchModel{{FileID: 7, ExtractedText: "um termo"}}}
	service := NewService(repository)

	results, err := service.SearchTopDocuments("termo", 6)
	if err != nil || len(results) != 1 || results[0].FileID != 7 {
		t.Fatalf("unexpected results %+v err=%v", results, err)
	}
	if repository.gotLimit != 6 || repository.gotOffset != 0 {
		t.Fatalf("expected limit 6 offset 0, got %d/%d", repository.gotLimit, repository.gotOffset)
	}

	empty, err := service.SearchTopDocuments("", 6)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v err=%v", empty, err)
	}
}
