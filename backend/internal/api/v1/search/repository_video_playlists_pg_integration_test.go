package search

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
)

func useTranslationLocale(t *testing.T, locale string) {
	t.Helper()
	previousLocale := config.AppConfig.Lang
	config.AppConfig.Lang = locale
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("load translations %s: %v", locale, err)
	}
	t.Cleanup(func() {
		config.AppConfig.Lang = previousLocale
		_ = i18n.LoadTranslations()
	})
}

func newVideoPlaylistPostgresRepository(t *testing.T) (*Repository, *database.DbContext) {
	t.Helper()
	repository, dbContext := newSearchPostgresRepository(t)
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`TRUNCATE video_playlist RESTART IDENTITY CASCADE`)
		return execErr
	})
	if err != nil {
		t.Fatalf("truncate video playlists: %v", err)
	}
	return repository, dbContext
}

func seedVideoPlaylist(t *testing.T, dbContext *database.DbContext, sourcePath string, storedName string, nameKey string) {
	t.Helper()
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(
			`INSERT INTO video_playlist (type, source_path, name, name_key, is_auto) VALUES ('folder', $1, $2, NULLIF($3, ''), TRUE)`,
			sourcePath, storedName, nameKey,
		)
		return execErr
	})
	if err != nil {
		t.Fatalf("seed video playlist %q: %v", sourcePath, err)
	}
}

func TestSearchVideoPlaylistsMatchesResolvedNameAndKeepsDataNames_Postgres(t *testing.T) {
	repository, dbContext := newVideoPlaylistPostgresRepository(t)
	seedVideoPlaylist(t, dbContext, "related:movie", "Filmes", "VIDEO_PLAYLIST_NAME_MOVIES")
	seedVideoPlaylist(t, dbContext, "folder:/media/Filmes", "Filmes", "")
	seedVideoPlaylist(t, dbContext, "folder:/media/Aventura", "Aventura", "")
	useTranslationLocale(t, "en-US")

	foundByEnglishName, err := repository.SearchVideoPlaylists("movies", 10)
	if err != nil {
		t.Fatalf("search movies: %v", err)
	}
	if len(foundByEnglishName) != 1 || foundByEnglishName[0].SourcePath != "related:movie" || !foundByEnglishName[0].NameKey.Valid {
		t.Fatalf("movies = %+v", foundByEnglishName)
	}

	foundByStoredText, err := repository.SearchVideoPlaylists("filmes", 10)
	if err != nil {
		t.Fatalf("search filmes: %v", err)
	}
	if len(foundByStoredText) != 1 || foundByStoredText[0].SourcePath != "folder:/media/Filmes" {
		t.Fatalf("filmes = %+v", foundByStoredText)
	}

	foundByFolderName, err := repository.SearchVideoPlaylists("aventura", 10)
	if err != nil {
		t.Fatalf("search aventura: %v", err)
	}
	if len(foundByFolderName) != 1 || foundByFolderName[0].Name != "Aventura" {
		t.Fatalf("aventura = %+v", foundByFolderName)
	}
}

func TestSearchGlobalReturnsResolvedVideoPlaylistName_Postgres(t *testing.T) {
	repository, dbContext := newVideoPlaylistPostgresRepository(t)
	seedVideoPlaylist(t, dbContext, "related:movie", "Filmes", "VIDEO_PLAYLIST_NAME_MOVIES")
	useTranslationLocale(t, "en-US")

	response, err := NewService(repository, nil).SearchGlobal("movies", 10)
	if err != nil {
		t.Fatalf("SearchGlobal: %v", err)
	}
	if len(response.Playlists) != 1 || response.Playlists[0].Name != "Movies" {
		t.Fatalf("playlists = %+v", response.Playlists)
	}
}

func TestSearchVideoPlaylistsRanksExactResolvedNameFirst_Postgres(t *testing.T) {
	repository, dbContext := newVideoPlaylistPostgresRepository(t)
	seedVideoPlaylist(t, dbContext, "folder:/media/more", "More Movies Than Ever", "")
	seedVideoPlaylist(t, dbContext, "related:movie", "Filmes", "VIDEO_PLAYLIST_NAME_MOVIES")
	useTranslationLocale(t, "en-US")

	found, err := repository.SearchVideoPlaylists("movies", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(found) != 2 || found[0].SourcePath != "related:movie" {
		t.Fatalf("ranking = %+v", found)
	}
}
