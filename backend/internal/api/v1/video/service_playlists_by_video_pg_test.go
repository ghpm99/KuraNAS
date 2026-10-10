package video

import (
	"reflect"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestGetPlaylistsByVideoReturnsOnlyVisiblePlaylistsContainingTheVideo_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)
	videoIDs := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "target.mkv", folder: "/data/A"},
		{name: "other.mkv", folder: "/data/B"},
	})
	targetVideoID := videoIDs["target.mkv"]
	otherVideoID := videoIDs["other.mkv"]

	seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "/with-target", isAuto: true, itemKind: "auto", videoID: targetVideoID, playlistType: "folder"},
		{sourcePath: "/favorites", isAuto: false, itemKind: "manual", videoID: targetVideoID},
		{sourcePath: "/hidden-with-target", isAuto: false, itemKind: "manual", videoID: targetVideoID, isHidden: true},
		{sourcePath: "/with-other", isAuto: true, itemKind: "auto", videoID: otherVideoID},
		{sourcePath: "/empty", isAuto: false},
	})

	playlists, err := service.GetPlaylistsByVideo(targetVideoID)
	if err != nil {
		t.Fatalf("GetPlaylistsByVideo: %v", err)
	}

	gotPlaylists := make([][2]string, 0, len(playlists))
	for _, playlist := range playlists {
		gotPlaylists = append(gotPlaylists, [2]string{playlist.Name, playlist.Type})
	}
	wantPlaylists := [][2]string{{"/with-target", "folder"}, {"/favorites", "custom"}}
	if !reflect.DeepEqual(gotPlaylists, wantPlaylists) {
		t.Fatalf("playlists = %v, want %v", gotPlaylists, wantPlaylists)
	}

	noPlaylists, err := service.GetPlaylistsByVideo(otherVideoID + 1000)
	if err != nil || len(noPlaylists) != 0 {
		t.Fatalf("unknown video playlists = %v err=%v", noPlaylists, err)
	}
}
