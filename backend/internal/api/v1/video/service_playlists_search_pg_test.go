package video

import (
	"database/sql"
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

func insertKeyedPlaylist(t *testing.T, repository *Repository, sourcePath string, storedName string, nameKey string, videoID int) {
	t.Helper()
	insertErr := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		var playlistID int
		err := tx.QueryRow(
			`INSERT INTO video_playlist (type, source_path, name, name_key, is_auto) VALUES ('folder', $1, $2, NULLIF($3, ''), TRUE) RETURNING id`,
			sourcePath, storedName, nameKey,
		).Scan(&playlistID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			`INSERT INTO video_playlist_item (playlist_id, video_id, order_index, source_kind) VALUES ($1, $2, 0, 'auto')`,
			playlistID, videoID,
		)
		return err
	})
	if insertErr != nil {
		t.Fatalf("insert playlist %s: %v", sourcePath, insertErr)
	}
}

func TestSearchPlaylistsMatchesResolvedNameOfKeyedPlaylist_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)
	videoIDs := seedVideoFiles(t, repository, []seededVideoFile{{name: "a.mp4", folder: "/data"}})
	seedPlaylists(t, repository, nil)
	insertKeyedPlaylist(t, repository, "related:movie", "Filmes", "VIDEO_PLAYLIST_NAME_MOVIES", videoIDs["a.mp4"])
	insertKeyedPlaylist(t, repository, "folder:/media/Filmes", "Filmes", "", videoIDs["a.mp4"])
	insertKeyedPlaylist(t, repository, "folder:/media/Aventura", "Aventura", "", videoIDs["a.mp4"])
	useTranslationLocale(t, "en-US")

	cases := map[string][]string{
		"movies":    {"related:movie"},
		"MOVIES":    {"related:movie"},
		"filmes":    {"folder:/media/Filmes"},
		"aventura":  {"folder:/media/Aventura"},
		"nonexists": {},
	}
	for query, expectedSourcePaths := range cases {
		searchPage, err := service.SearchPlaylists(PlaylistSearchRequest{Query: query, Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("query %q: %v", query, err)
		}
		if got := sectionSourcePaths(searchPage.Items); !reflect.DeepEqual(got, sortedCopy(expectedSourcePaths)) {
			t.Fatalf("query %q = %v, want %v", query, got, expectedSourcePaths)
		}
	}
}

func TestSearchPlaylistsReturnsResolvedNameInCurrentLanguage_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	service := NewService(repository)
	videoIDs := seedVideoFiles(t, repository, []seededVideoFile{{name: "a.mp4", folder: "/data"}})
	seedPlaylists(t, repository, nil)
	insertKeyedPlaylist(t, repository, "related:movie", "Filmes", "VIDEO_PLAYLIST_NAME_MOVIES", videoIDs["a.mp4"])
	useTranslationLocale(t, "en-US")

	searchPage, err := service.SearchPlaylists(PlaylistSearchRequest{Query: "movies", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(searchPage.Items) != 1 || searchPage.Items[0].Name != "Movies" {
		t.Fatalf("items = %+v", searchPage.Items)
	}
}
