package video

import (
	"database/sql"
	"reflect"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestGetPlaylistItemsPageOrdersPaginatesAndReportsStatus_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)

	episodeNames := []string{"e1.mkv", "e2.mkv", "e3.mkv", "e4.mkv", "e5.mkv"}
	seeds := make([]seededVideoFile, len(episodeNames))
	for position, name := range episodeNames {
		seeds[position] = seededVideoFile{name: name, folder: "/data/show"}
	}
	videoIDsByName := seedVideoFiles(t, repository, seeds)
	playlistID := seedPlaylists(t, repository, []seededPlaylist{{sourcePath: "/show", isAuto: false}})["/show"]

	orderedVideoIDs := make([]int, len(episodeNames))
	for position, name := range episodeNames {
		orderedVideoIDs[position] = videoIDsByName[name]
	}
	setupErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if err := repository.InsertPlaylistItemsWithSource(tx, playlistID, orderedVideoIDs, "manual"); err != nil {
			return err
		}
		_, err := repository.UpsertVideoWatchProgress(tx, VideoWatchProgressModel{
			ClientID: "tv", VideoID: orderedVideoIDs[1], PositionSeconds: 100, DurationSeconds: 100, Completed: true,
		})
		return err
	})
	if setupErr != nil {
		t.Fatalf("setup: %v", setupErr)
	}

	firstPage, err := service.GetPlaylistItemsPage("tv", playlistID, 1, 2)
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	if got := itemVideoIDs(firstPage.Items); !reflect.DeepEqual(got, orderedVideoIDs[:2]) {
		t.Fatalf("first page ids = %v, want %v", got, orderedVideoIDs[:2])
	}
	if !firstPage.Pagination.HasNext || firstPage.Pagination.HasPrev {
		t.Fatalf("unexpected first page pagination: %+v", firstPage.Pagination)
	}
	if firstPage.Items[0].Status != "not_started" || firstPage.Items[1].Status != "completed" {
		t.Fatalf("unexpected statuses: %s, %s", firstPage.Items[0].Status, firstPage.Items[1].Status)
	}

	lastPage, err := service.GetPlaylistItemsPage("tv", playlistID, 3, 2)
	if err != nil {
		t.Fatalf("last page: %v", err)
	}
	if got := itemVideoIDs(lastPage.Items); !reflect.DeepEqual(got, orderedVideoIDs[4:]) {
		t.Fatalf("last page ids = %v, want %v", got, orderedVideoIDs[4:])
	}
	if lastPage.Pagination.HasNext || !lastPage.Pagination.HasPrev {
		t.Fatalf("unexpected last page pagination: %+v", lastPage.Pagination)
	}

	if _, err := service.GetPlaylistItemsPage("tv", playlistID+999, 1, 2); err == nil {
		t.Fatalf("expected error for unknown playlist")
	}
}

func itemVideoIDs(items []VideoPlaylistItemDto) []int {
	videoIDs := make([]int, 0, len(items))
	for _, item := range items {
		videoIDs = append(videoIDs, item.Video.ID)
	}
	return videoIDs
}
