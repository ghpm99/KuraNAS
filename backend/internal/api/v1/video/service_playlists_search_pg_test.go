package video

import (
	"reflect"
	"sort"
	"testing"

	"nas-go/api/internal/testutil"
)

func sortedCopy(values []string) []string {
	sorted := append([]string{}, values...)
	sort.Strings(sorted)
	return sorted
}

func TestSearchPlaylistsFoldsAccentsAndExcludesHidden_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)

	seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "Ação Clássicos"},
		{sourcePath: "acao moderna"},
		{sourcePath: "Comédia"},
		{sourcePath: "Ação oculta", isHidden: true},
		{sourcePath: "100% real"},
		{sourcePath: "1000 reais"},
	})

	cases := map[string][]string{
		"acao":    {"Ação Clássicos", "acao moderna"},
		"AÇÃO":    {"Ação Clássicos", "acao moderna"},
		"comedia": {"Comédia"},
		"100%":    {"100% real"},
		"nada":    {},
	}
	for query, expectedNames := range cases {
		searchPage, err := service.SearchPlaylists(PlaylistSearchRequest{Query: query, Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("query %q: %v", query, err)
		}
		if got := sectionSourcePaths(searchPage.Items); !reflect.DeepEqual(got, sortedCopy(expectedNames)) {
			t.Fatalf("query %q = %v, want %v", query, got, expectedNames)
		}
	}
}

func TestSearchPlaylistsPaginatesWithoutOverlap_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)

	seedPlaylists(t, repository, []seededPlaylist{
		{sourcePath: "/s1"}, {sourcePath: "/s2"}, {sourcePath: "/s3"}, {sourcePath: "/s4"}, {sourcePath: "/s5"},
	})

	collectedSourcePaths := []string{}
	for page := 1; page <= 3; page++ {
		searchPage, err := service.SearchPlaylists(PlaylistSearchRequest{Page: page, PageSize: 2})
		if err != nil {
			t.Fatalf("page %d: %v", page, err)
		}
		if searchPage.Pagination.HasNext == (page == 3) {
			t.Fatalf("page %d unexpected has_next %+v", page, searchPage.Pagination)
		}
		collectedSourcePaths = append(collectedSourcePaths, sectionSourcePaths(searchPage.Items)...)
	}
	if want := []string{"/s1", "/s2", "/s3", "/s4", "/s5"}; !reflect.DeepEqual(sortedCopy(collectedSourcePaths), want) {
		t.Fatalf("paged playlists = %v, want %v", collectedSourcePaths, want)
	}
}
