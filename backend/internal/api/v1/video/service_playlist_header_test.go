package video

import "testing"

func TestGetPlaylistHeaderByIDReturnsCountWithoutLoadingItems(t *testing.T) {
	repo := &videoRepoMock{
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: id, Name: "Show"}, nil
		},
		countVideoPlaylistItemsFn: func(playlistID int) (int, error) { return 3, nil },
		getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
			t.Fatalf("detailed items must not be loaded for the header")
			return nil, nil
		},
		getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
			t.Fatalf("progress must not be computed for the header")
			return VideoPlaybackStateModel{}, nil
		},
	}

	header, err := newVideoServiceForTest(t, repo).GetPlaylistHeaderByID(7)
	if err != nil {
		t.Fatalf("expected header success, err=%v", err)
	}
	if header.ID != 7 || header.ItemCount != 3 {
		t.Fatalf("unexpected header id=%d item_count=%d", header.ID, header.ItemCount)
	}
	if header.Items == nil || len(header.Items) != 0 {
		t.Fatalf("expected empty non-nil items, got %#v", header.Items)
	}
}
