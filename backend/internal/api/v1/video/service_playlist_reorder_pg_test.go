package video

import (
	"database/sql"
	"reflect"
	"testing"

	"nas-go/api/internal/testutil"
)

func reorderScenario(t *testing.T, initialNames []string, reorderedNames []string) {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)

	seeds := make([]seededVideoFile, len(initialNames))
	for position, name := range initialNames {
		seeds[position] = seededVideoFile{name: name, folder: "/data/A"}
	}
	videoIDsByName := seedVideoFiles(t, repository, seeds)
	playlistID := seedPlaylists(t, repository, []seededPlaylist{{sourcePath: "/reorder", isAuto: false}})["/reorder"]

	initialVideoIDs := make([]int, len(initialNames))
	for position, name := range initialNames {
		initialVideoIDs[position] = videoIDsByName[name]
	}
	setupErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		return repository.InsertPlaylistItemsWithSource(tx, playlistID, initialVideoIDs, "manual")
	})
	if setupErr != nil {
		t.Fatalf("setup: %v", setupErr)
	}

	requests := make([]ReorderPlaylistItemRequest, len(reorderedNames))
	wantVideoIDs := make([]int, len(reorderedNames))
	for position, name := range reorderedNames {
		requests[position] = ReorderPlaylistItemRequest{VideoID: videoIDsByName[name], OrderIndex: position}
		wantVideoIDs[position] = videoIDsByName[name]
	}

	if err := service.ReorderPlaylistItems(playlistID, requests); err != nil {
		t.Fatalf("reorder failed: %v", err)
	}

	gotVideoIDs := playlistVideoIDsInOrder(t, repository, playlistID)
	if !reflect.DeepEqual(gotVideoIDs, wantVideoIDs) {
		t.Fatalf("order = %v, want %v", gotVideoIDs, wantVideoIDs)
	}
}

func TestReorderPlaylistItemsMovesFirstItemToLast_Postgres(t *testing.T) {
	reorderScenario(t, []string{"a.mkv", "b.mkv", "c.mkv", "d.mkv"}, []string{"b.mkv", "c.mkv", "d.mkv", "a.mkv"})
}

func TestReorderPlaylistItemsMovesLastItemToFirst_Postgres(t *testing.T) {
	reorderScenario(t, []string{"a.mkv", "b.mkv", "c.mkv", "d.mkv"}, []string{"d.mkv", "a.mkv", "b.mkv", "c.mkv"})
}

func TestReorderPlaylistItemsSwapsMiddleItems_Postgres(t *testing.T) {
	reorderScenario(t, []string{"a.mkv", "b.mkv", "c.mkv", "d.mkv"}, []string{"a.mkv", "c.mkv", "b.mkv", "d.mkv"})
}
