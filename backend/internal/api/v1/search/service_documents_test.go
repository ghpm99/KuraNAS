package search

import (
	"errors"
	"testing"

	"nas-go/api/internal/api/v1/documenttext"
)

type documentSearcherStub struct {
	results  []documenttext.DocumentSearchResultDto
	err      error
	gotLimit int
}

func (stub *documentSearcherStub) SearchTopDocuments(query string, limit int) ([]documenttext.DocumentSearchResultDto, error) {
	stub.gotLimit = limit
	return stub.results, stub.err
}

func emptyBucketsRepository() *searchRepositoryMock {
	return &searchRepositoryMock{
		searchFilesFn:          func(string, int) ([]FileResultModel, error) { return nil, nil },
		searchFoldersFn:        func(string, int) ([]FolderResultModel, error) { return nil, nil },
		searchArtistsFn:        func(string, int) ([]ArtistResultModel, error) { return nil, nil },
		searchAlbumsFn:         func(string, int) ([]AlbumResultModel, error) { return nil, nil },
		searchMusicPlaylistsFn: func(string, int) ([]MusicPlaylistResultModel, error) { return nil, nil },
		searchVideoPlaylistsFn: func(string, int) ([]VideoPlaylistResultModel, error) { return nil, nil },
		searchVideosFn:         func(string, int) ([]VideoResultModel, error) { return nil, nil },
		searchImagesFn:         func(string, int) ([]ImageResultModel, error) { return nil, nil },
		searchTracksFn:         func(string, int) ([]TrackResultModel, error) { return nil, nil },
	}
}

func TestSearchGlobalIncludesDocumentsGroup(t *testing.T) {
	searcher := &documentSearcherStub{results: []documenttext.DocumentSearchResultDto{{FileID: 3, Name: "a.pdf", Snippet: "trecho"}}}
	service := NewServiceWithDocuments(emptyBucketsRepository(), nil, searcher)

	response, err := service.SearchGlobal("termo", 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(response.Documents) != 1 || response.Documents[0].Snippet != "trecho" || searcher.gotLimit != 4 {
		t.Fatalf("unexpected documents %+v limit=%d", response.Documents, searcher.gotLimit)
	}
}

func TestSearchGlobalDocumentsFailureDegradesToEmptyGroup(t *testing.T) {
	service := NewServiceWithDocuments(emptyBucketsRepository(), nil, &documentSearcherStub{err: errors.New("boom")})

	response, err := service.SearchGlobal("termo", 4)
	if err != nil || response.Documents == nil || len(response.Documents) != 0 {
		t.Fatalf("expected empty documents group, got %+v err=%v", response.Documents, err)
	}
}

func TestSearchGlobalWithoutDocumentSearcherReturnsEmptyGroup(t *testing.T) {
	response, err := NewService(emptyBucketsRepository(), nil).SearchGlobal("termo", 4)
	if err != nil || response.Documents == nil || len(response.Documents) != 0 {
		t.Fatalf("expected empty documents group, got %+v err=%v", response.Documents, err)
	}
}
