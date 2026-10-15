package video

import (
	"testing"

	"nas-go/api/internal/testutil"
)

func TestCountVideoPlaylistItemsIgnoresDeletedFiles_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	ids := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "e1.mkv", folder: "/data/Show"},
		{name: "e2.mkv", folder: "/data/Show"},
		{name: "gone.mkv", folder: "/data/Show", deleted: true},
	})
	playlistID := seedPlaylistWithVideos(t, repository, ids, "e1.mkv", "e2.mkv", "gone.mkv")

	itemCount, err := repository.CountVideoPlaylistItems(playlistID)
	if err != nil || itemCount != 2 {
		t.Fatalf("item count = %d err = %v, want 2", itemCount, err)
	}
}
