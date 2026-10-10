package video

import (
	"database/sql"
	"reflect"
	"testing"

	"nas-go/api/internal/testutil"
)

func playlistVideoIDsInOrder(t *testing.T, repository *Repository, playlistID int) []int {
	t.Helper()
	rows, err := repository.DbContext.GetDatabase().Query(
		`SELECT video_id, order_index FROM video_playlist_item WHERE playlist_id = $1 ORDER BY order_index`, playlistID)
	if err != nil {
		t.Fatalf("query items: %v", err)
	}
	defer rows.Close()
	videoIDs := []int{}
	expectedOrderIndex := 0
	for rows.Next() {
		var videoID, orderIndex int
		if err := rows.Scan(&videoID, &orderIndex); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if orderIndex != expectedOrderIndex {
			t.Fatalf("order_index not contiguous: got %d, want %d", orderIndex, expectedOrderIndex)
		}
		expectedOrderIndex++
		videoIDs = append(videoIDs, videoID)
	}
	return videoIDs
}

func TestRebuildAutoItemsKeepsManualItemsAndRenumbersContiguously_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	videoIDs := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "a.mkv", folder: "/data/A"},
		{name: "b.mkv", folder: "/data/A"},
		{name: "c.mkv", folder: "/data/A"},
		{name: "d.mkv", folder: "/data/A"},
	})
	playlistIDs := seedPlaylists(t, repository, []seededPlaylist{{sourcePath: "/rebuild", isAuto: true}})
	playlistID := playlistIDs["/rebuild"]

	setupErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if err := repository.InsertPlaylistItemsWithSource(tx, playlistID, []int{videoIDs["a.mkv"], videoIDs["b.mkv"]}, "auto"); err != nil {
			return err
		}
		return repository.AddPlaylistVideoManual(tx, playlistID, videoIDs["c.mkv"])
	})
	if setupErr != nil {
		t.Fatalf("setup: %v", setupErr)
	}

	rebuildErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if err := repository.DeleteAutoPlaylistItems(tx, playlistID); err != nil {
			return err
		}
		if err := repository.InsertPlaylistItemsWithSource(tx, playlistID, []int{videoIDs["b.mkv"], videoIDs["a.mkv"], videoIDs["d.mkv"]}, "auto"); err != nil {
			return err
		}
		return repository.RenumberPlaylistItems(tx, playlistID)
	})
	if rebuildErr != nil {
		t.Fatalf("rebuild transaction failed: %v", rebuildErr)
	}

	got := playlistVideoIDsInOrder(t, repository, playlistID)
	want := []int{videoIDs["b.mkv"], videoIDs["a.mkv"], videoIDs["d.mkv"], videoIDs["c.mkv"]}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}
