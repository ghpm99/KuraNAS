package video

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestGetHiddenPlaylistsReturnsOnlyHiddenOrderedByNameThenID_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)

	seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "/visible"},
		{sourcePath: "/zeta", isHidden: true},
		{sourcePath: "/alfa", isHidden: true},
		{sourcePath: "/beta", isHidden: true},
	})

	hiddenPage, err := service.GetHiddenPlaylists(HiddenPlaylistsRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("hidden playlists: %v", err)
	}
	orderedNames := []string{}
	for _, playlist := range hiddenPage.Items {
		if !playlist.IsHidden {
			t.Fatalf("playlist %q is not hidden", playlist.Name)
		}
		orderedNames = append(orderedNames, playlist.Name)
	}
	if want := []string{"/alfa", "/beta", "/zeta"}; !reflect.DeepEqual(orderedNames, want) {
		t.Fatalf("names = %v, want %v", orderedNames, want)
	}
}

func TestGetHiddenPlaylistsPaginatesWithoutOverlap_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)

	seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "/h1", isHidden: true}, {sourcePath: "/h2", isHidden: true}, {sourcePath: "/h3", isHidden: true},
		{sourcePath: "/h4", isHidden: true}, {sourcePath: "/h5", isHidden: true},
	})

	collectedSourcePaths := []string{}
	for page := 1; page <= 3; page++ {
		hiddenPage, err := service.GetHiddenPlaylists(HiddenPlaylistsRequest{Page: page, PageSize: 2})
		if err != nil {
			t.Fatalf("page %d: %v", page, err)
		}
		for _, playlist := range hiddenPage.Items {
			collectedSourcePaths = append(collectedSourcePaths, playlist.SourcePath)
		}
		if wantHasNext := page < 3; hiddenPage.Pagination.HasNext != wantHasNext {
			t.Fatalf("page %d HasNext = %t, want %t", page, hiddenPage.Pagination.HasNext, wantHasNext)
		}
	}
	if want := []string{"/h1", "/h2", "/h3", "/h4", "/h5"}; !reflect.DeepEqual(collectedSourcePaths, want) {
		t.Fatalf("collected = %v, want %v", collectedSourcePaths, want)
	}
}

func TestSetPlaylistHiddenMovesPlaylistBetweenListings_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)
	playlistID := seedPlaylists(t, repository, []seededPlaylist{{sourcePath: "/movable", classification: "series"}})["/movable"]

	if err := service.SetPlaylistHidden(playlistID, true); err != nil {
		t.Fatalf("hide: %v", err)
	}
	hiddenPage, _ := service.GetHiddenPlaylists(HiddenPlaylistsRequest{Page: 1, PageSize: 10})
	sectionPage, _ := service.GetPlaylistsBySection(PlaylistSectionRequest{Section: PlaylistSectionSeries, Page: 1, PageSize: 10})
	if len(hiddenPage.Items) != 1 || len(sectionPage.Items) != 0 {
		t.Fatalf("after hide: hidden=%d section=%d", len(hiddenPage.Items), len(sectionPage.Items))
	}

	if err := service.SetPlaylistHidden(playlistID, false); err != nil {
		t.Fatalf("show: %v", err)
	}
	hiddenPage, _ = service.GetHiddenPlaylists(HiddenPlaylistsRequest{Page: 1, PageSize: 10})
	sectionPage, _ = service.GetPlaylistsBySection(PlaylistSectionRequest{Section: PlaylistSectionSeries, Page: 1, PageSize: 10})
	if len(hiddenPage.Items) != 0 || len(sectionPage.Items) != 1 {
		t.Fatalf("after show: hidden=%d section=%d", len(hiddenPage.Items), len(sectionPage.Items))
	}
}

func TestSetPlaylistHiddenOfMissingPlaylistReturnsNoRows_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	service := NewService(NewRepository(dbContext))

	if err := service.SetPlaylistHidden(999999, true); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestRebuildSmartPlaylistsPreservesHiddenAutomaticPlaylists_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)
	seedVideoFiles(t, repository, []seededVideoFile{
		{name: "a.mkv", folder: "/data/Shows/A"},
		{name: "b.mkv", folder: "/data/Shows/A"},
		{name: "c.mkv", folder: "/data/Shows/B"},
	})
	if _, err := dbContext.GetDatabase().Exec(`TRUNCATE video_playlist RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate playlists: %v", err)
	}

	if err := service.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("first rebuild: %v", err)
	}
	automaticTotal := countRows(t, repository, `SELECT COUNT(*) FROM video_playlist WHERE is_auto = TRUE`)
	if automaticTotal == 0 {
		t.Fatal("rebuild produced no automatic playlists")
	}
	if _, err := dbContext.GetDatabase().Exec(`UPDATE video_playlist SET is_hidden = TRUE WHERE is_auto = TRUE`); err != nil {
		t.Fatalf("hide automatic playlists: %v", err)
	}

	if err := service.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("second rebuild: %v", err)
	}
	if hiddenTotal := countRows(t, repository, `SELECT COUNT(*) FROM video_playlist WHERE is_auto = TRUE AND is_hidden = TRUE`); hiddenTotal != automaticTotal {
		t.Fatalf("hidden automatic playlists after rebuild = %d, want %d", hiddenTotal, automaticTotal)
	}
	if automaticAfter := countRows(t, repository, `SELECT COUNT(*) FROM video_playlist WHERE is_auto = TRUE`); automaticAfter != automaticTotal {
		t.Fatalf("automatic playlists after rebuild = %d, want %d", automaticAfter, automaticTotal)
	}
}
