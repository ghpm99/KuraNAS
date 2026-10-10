package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

type seededPlaylist struct {
	sourcePath string
	isAuto     bool
	itemKind   string
	videoID    int
}

func seedPlaylists(t *testing.T, repository *Repository, seeds []seededPlaylist) map[string]int {
	t.Helper()
	playlistIDsBySourcePath := map[string]int{}
	seedErr := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_playlist RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		for _, seed := range seeds {
			var playlistID int
			playlistType := "folder"
			if !seed.isAuto {
				playlistType = "custom"
			}
			err := tx.QueryRow(
				`INSERT INTO video_playlist (type, source_path, name, is_auto) VALUES ($1, $2, $2, $3) RETURNING id`,
				playlistType, seed.sourcePath, seed.isAuto,
			).Scan(&playlistID)
			if err != nil {
				return err
			}
			playlistIDsBySourcePath[seed.sourcePath] = playlistID
			if seed.itemKind == "" {
				continue
			}
			if _, err := tx.Exec(
				`INSERT INTO video_playlist_item (playlist_id, video_id, order_index, source_kind) VALUES ($1, $2, 0, $3)`,
				playlistID, seed.videoID, seed.itemKind,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed playlists: %v", seedErr)
	}
	return playlistIDsBySourcePath
}

func listedSourcePaths(t *testing.T, repository *Repository) map[string]bool {
	t.Helper()
	playlists, err := repository.GetVideoPlaylists(true)
	if err != nil {
		t.Fatalf("GetVideoPlaylists: %v", err)
	}
	sourcePaths := map[string]bool{}
	for _, playlist := range playlists {
		sourcePaths[playlist.SourcePath] = true
	}
	return sourcePaths
}

func TestDeleteStaleAutoPlaylistsKeepsProducedAndManuallyPopulatedAndCustom_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	videoIDs := seedVideoFiles(t, repository, []seededVideoFile{{name: "a.mkv", folder: "/data/A"}})
	videoID := videoIDs["a.mkv"]
	playlistIDs := seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "/produced", isAuto: true, itemKind: "auto", videoID: videoID},
		{sourcePath: "/orphan-auto", isAuto: true, itemKind: "auto", videoID: videoID},
		{sourcePath: "/orphan-with-manual", isAuto: true, itemKind: "manual", videoID: videoID},
		{sourcePath: "/custom-empty", isAuto: false},
		{sourcePath: "/context-in-playback", isAuto: true, itemKind: "auto", videoID: videoID},
	})
	playbackErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO video_playback_state (client_id, playlist_id, video_id) VALUES ('stale-test-client', $1, $2)
			 ON CONFLICT (client_id) DO UPDATE SET playlist_id = EXCLUDED.playlist_id, video_id = EXCLUDED.video_id`,
			playlistIDs["/context-in-playback"], videoID,
		)
		return err
	})
	if playbackErr != nil {
		t.Fatalf("seed playback state: %v", playbackErr)
	}

	deleteErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		return repository.DeleteStaleAutoPlaylists(tx, []int{playlistIDs["/produced"]})
	})
	if deleteErr != nil {
		t.Fatalf("DeleteStaleAutoPlaylists: %v", deleteErr)
	}

	remaining := listedSourcePaths(t, repository)
	if remaining["/orphan-auto"] {
		t.Fatalf("orphan auto playlist should be deleted")
	}
	for _, kept := range []string{"/produced", "/orphan-with-manual", "/custom-empty", "/context-in-playback"} {
		if !remaining[kept] {
			t.Fatalf("playlist %s should be kept, remaining=%v", kept, remaining)
		}
	}
}

func TestGetVideoPlaylistsHidesEmptyAutoButListsEmptyCustom_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	videoIDs := seedVideoFiles(t, repository, []seededVideoFile{
		{name: "live.mkv", folder: "/data/A"},
		{name: "gone.mkv", folder: "/data/B", deleted: true},
	})
	seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "/auto-empty", isAuto: true},
		{sourcePath: "/auto-only-deleted", isAuto: true, itemKind: "auto", videoID: videoIDs["gone.mkv"]},
		{sourcePath: "/auto-populated", isAuto: true, itemKind: "auto", videoID: videoIDs["live.mkv"]},
		{sourcePath: "/custom-empty", isAuto: false},
	})

	listed := listedSourcePaths(t, repository)

	if listed["/auto-empty"] || listed["/auto-only-deleted"] {
		t.Fatalf("auto playlists without active items must be hidden, listed=%v", listed)
	}
	if !listed["/auto-populated"] || !listed["/custom-empty"] {
		t.Fatalf("populated auto and empty custom must be listed, listed=%v", listed)
	}
}
