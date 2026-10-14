package video

import (
	"errors"
	"testing"
)

func TestGetHiddenPlaylistsPagesWithOneExtraRow(t *testing.T) {
	var receivedLimit, receivedOffset int
	repository := &videoRepoMock{getHiddenVideoPlaylistsFn: func(limit int, offset int) ([]VideoPlaylistModel, error) {
		receivedLimit, receivedOffset = limit, offset
		return []VideoPlaylistModel{{ID: 1, IsHidden: true}, {ID: 2, IsHidden: true}, {ID: 3, IsHidden: true}}, nil
	}}

	hiddenPage, err := newVideoServiceForTest(t, repository).GetHiddenPlaylists(HiddenPlaylistsRequest{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedLimit != 3 || receivedOffset != 2 {
		t.Fatalf("unexpected window limit=%d offset=%d", receivedLimit, receivedOffset)
	}
	if len(hiddenPage.Items) != 2 || !hiddenPage.Pagination.HasNext || !hiddenPage.Pagination.HasPrev {
		t.Fatalf("unexpected page %+v", hiddenPage)
	}
}

func TestGetHiddenPlaylistsPropagatesRepositoryError(t *testing.T) {
	repository := &videoRepoMock{getHiddenVideoPlaylistsFn: func(int, int) ([]VideoPlaylistModel, error) {
		return nil, errors.New("boom")
	}}
	if _, err := newVideoServiceForTest(t, repository).GetHiddenPlaylists(HiddenPlaylistsRequest{Page: 1, PageSize: 5}); err == nil {
		t.Fatal("expected error")
	}
}
