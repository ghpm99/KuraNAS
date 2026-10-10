package video

import (
	"reflect"
	"sort"
	"testing"

	"nas-go/api/internal/testutil"
)

func sectionSourcePaths(playlists []VideoPlaylistDto) []string {
	sourcePaths := make([]string, 0, len(playlists))
	for _, playlist := range playlists {
		sourcePaths = append(sourcePaths, playlist.SourcePath)
	}
	sort.Strings(sourcePaths)
	return sourcePaths
}

func TestGetPlaylistsBySectionFiltersEachSection_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)

	seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "/series", classification: "series"},
		{sourcePath: "/anime", classification: "anime"},
		{sourcePath: "/movie", classification: "movie"},
		{sourcePath: "/personal", classification: "personal"},
		{sourcePath: "/clip", classification: "clip"},
		{sourcePath: "/program", classification: "program"},
		{sourcePath: "/folder", classification: "other", playlistType: "folder"},
		{sourcePath: "/hidden-series", classification: "series", isHidden: true},
	})

	expectedBySection := map[PlaylistSection][]string{
		PlaylistSectionSeries:   {"/anime", "/series"},
		PlaylistSectionMovies:   {"/movie"},
		PlaylistSectionPersonal: {"/personal"},
		PlaylistSectionClips:    {"/clip", "/program"},
		PlaylistSectionFolders:  {"/folder"},
	}

	for section, expectedSourcePaths := range expectedBySection {
		sectionPage, err := service.GetPlaylistsBySection(PlaylistSectionRequest{Section: section, Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("section %s: %v", section, err)
		}
		if got := sectionSourcePaths(sectionPage.Items); !reflect.DeepEqual(got, expectedSourcePaths) {
			t.Fatalf("section %s = %v, want %v", section, got, expectedSourcePaths)
		}
		if sectionPage.Pagination.HasNext || sectionPage.Pagination.HasPrev {
			t.Fatalf("section %s unexpected pagination %+v", section, sectionPage.Pagination)
		}
	}
}

func TestGetPlaylistsBySectionPaginatesWithoutOverlap_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)

	seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "/s1", classification: "series"},
		{sourcePath: "/s2", classification: "series"},
		{sourcePath: "/s3", classification: "series"},
		{sourcePath: "/s4", classification: "anime"},
		{sourcePath: "/s5", classification: "series"},
	})

	collectedSourcePaths := []string{}
	for page := 1; page <= 3; page++ {
		sectionPage, err := service.GetPlaylistsBySection(PlaylistSectionRequest{Section: PlaylistSectionSeries, Page: page, PageSize: 2})
		if err != nil {
			t.Fatalf("page %d: %v", page, err)
		}
		isLastPage := page == 3
		if sectionPage.Pagination.HasNext == isLastPage {
			t.Fatalf("page %d unexpected has_next %+v", page, sectionPage.Pagination)
		}
		if sectionPage.Pagination.HasPrev == (page == 1) {
			t.Fatalf("page %d unexpected has_prev %+v", page, sectionPage.Pagination)
		}
		collectedSourcePaths = append(collectedSourcePaths, sectionSourcePaths(sectionPage.Items)...)
	}

	sort.Strings(collectedSourcePaths)
	if want := []string{"/s1", "/s2", "/s3", "/s4", "/s5"}; !reflect.DeepEqual(collectedSourcePaths, want) {
		t.Fatalf("paged playlists = %v, want %v", collectedSourcePaths, want)
	}
}

func TestGetPlaylistsBySectionHidesAutoPlaylistsWithoutActiveItems_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)

	seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "/auto-empty", isAuto: true, classification: "series"},
		{sourcePath: "/custom-empty", classification: "series"},
	})

	sectionPage, err := service.GetPlaylistsBySection(PlaylistSectionRequest{Section: PlaylistSectionSeries, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("section: %v", err)
	}
	if got := sectionSourcePaths(sectionPage.Items); !reflect.DeepEqual(got, []string{"/custom-empty"}) {
		t.Fatalf("section = %v", got)
	}
}
