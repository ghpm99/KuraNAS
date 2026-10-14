package video

import (
	"errors"
	"testing"
)

func TestSearchPlaylistsEscapesWildcardsAndPagesWithOneExtraRow(t *testing.T) {
	var receivedPattern string
	var receivedLimit, receivedOffset int
	repository := &videoRepoMock{searchVideoPlaylistsFn: func(searchPattern string, limit int, offset int) ([]VideoPlaylistModel, error) {
		receivedPattern, receivedLimit, receivedOffset = searchPattern, limit, offset
		return []VideoPlaylistModel{{ID: 1}, {ID: 2}, {ID: 3}}, nil
	}}

	searchPage, err := newVideoServiceForTest(t, repository).SearchPlaylists(PlaylistSearchRequest{Query: "  100%  ", Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedPattern != `%100\%%` || receivedLimit != 3 || receivedOffset != 2 {
		t.Fatalf("unexpected window pattern=%q limit=%d offset=%d", receivedPattern, receivedLimit, receivedOffset)
	}
	if len(searchPage.Items) != 2 || !searchPage.Pagination.HasNext || !searchPage.Pagination.HasPrev {
		t.Fatalf("unexpected page %+v", searchPage)
	}
}

func TestSearchPlaylistsWithBlankQueryUsesEmptyPattern(t *testing.T) {
	var receivedPattern = "unset"
	repository := &videoRepoMock{searchVideoPlaylistsFn: func(searchPattern string, limit int, offset int) ([]VideoPlaylistModel, error) {
		receivedPattern = searchPattern
		return nil, nil
	}}

	if _, err := newVideoServiceForTest(t, repository).SearchPlaylists(PlaylistSearchRequest{Query: "   ", Page: 1, PageSize: 5}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedPattern != "" {
		t.Fatalf("pattern = %q, want empty", receivedPattern)
	}
}

func TestSearchPlaylistsPropagatesRepositoryError(t *testing.T) {
	repository := &videoRepoMock{searchVideoPlaylistsFn: func(string, int, int) ([]VideoPlaylistModel, error) {
		return nil, errors.New("boom")
	}}
	if _, err := newVideoServiceForTest(t, repository).SearchPlaylists(PlaylistSearchRequest{Page: 1, PageSize: 5}); err == nil {
		t.Fatal("expected error")
	}
}
