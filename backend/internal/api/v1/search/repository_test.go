package search

import (
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/search"

	"github.com/DATA-DOG/go-sqlmock"
)

func newSearchRepositoryForTest(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return NewRepository(database.NewDbContext(db)), mock
}

func singleTermMatchArgs(limit int) []driver.Value {
	return []driver.Value{"%mix%", sqlmock.AnyArg(), "mix", "mix%", "%mix%", limit}
}

func singleTermMatchArgsWithFormats(limit int) []driver.Value {
	return []driver.Value{"%mix%", sqlmock.AnyArg(), "mix", "mix%", "%mix%", sqlmock.AnyArg(), limit}
}

func singleTermFilesMatchArgs(limit int) []driver.Value {
	return singleTermMatchArgsWithFormats(limit)
}

func singleTermAudioMatchArgs(limit int) []driver.Value {
	return []driver.Value{"%mix%", sqlmock.AnyArg(), "mix", "mix%", sqlmock.AnyArg(), limit}
}

func TestSearchRepositorySuccessPaths(t *testing.T) {
	repository, mock := newSearchRepositoryForTest(t)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchFilesQuery)).
		WithArgs(singleTermFilesMatchArgs(5)...).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "parent_path", "format", "starred", "size", "updated_at", "is_cold"}).
			AddRow(1, "song.mp3", "/media/song.mp3", "/media", ".mp3", true, int64(2048), now, true))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchFoldersQuery)).
		WithArgs(singleTermMatchArgs(5)...).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "parent_path", "starred", "size", "updated_at", "is_cold"}).
			AddRow(2, "Photos", "/photos", "/", false, int64(0), now, false))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchArtistsQuery)).
		WithArgs(singleTermAudioMatchArgs(5)...).
		WillReturnRows(sqlmock.NewRows([]string{"artist", "track_count", "album_count"}).
			AddRow("Artist", 4, 2))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchAlbumsQuery)).
		WithArgs(singleTermAudioMatchArgs(5)...).
		WillReturnRows(sqlmock.NewRows([]string{"artist", "album", "year", "track_count"}).
			AddRow("Artist", "Album", "2026", 8))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchMusicPlaylistsQuery)).
		WithArgs("mix", 5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "updated_at", "track_count"}).
			AddRow(3, "Playlist", "Desc", true, now, 6))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchVideoPlaylistsQuery)).
		WithArgs("mix", 5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type", "classification", "source_path", "is_auto", "updated_at", "item_count"}).
			AddRow(4, "Series", "series", "series", "/videos/series", true, now, 10))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchVideosQuery)).
		WithArgs(singleTermMatchArgsWithFormats(5)...).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "parent_path", "format", "updated_at"}).
			AddRow(5, "Episode 01", "/videos/episode-01.mkv", "/videos", ".mkv", now))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchImagesQuery)).
		WithArgs(singleTermMatchArgsWithFormats(5)...).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "parent_path", "format", "updated_at", "category", "context"}).
			AddRow(6, "Vacation", "/photos/vacation.jpg", "/photos", ".jpg", now, "photo", "Canon"))
	mock.ExpectRollback()

	if repository.DbContext == nil {
		t.Fatalf("expected repository DbContext")
	}

	if items, err := repository.SearchFiles("mix", 5); err != nil || len(items) != 1 || items[0].ID != 1 || items[0].Size != 2048 || !items[0].IsCold || !items[0].UpdatedAt.Equal(now) {
		t.Fatalf("SearchFiles returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchFolders("mix", 5); err != nil || len(items) != 1 || items[0].ID != 2 || items[0].IsCold || !items[0].UpdatedAt.Equal(now) {
		t.Fatalf("SearchFolders returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchArtists("mix", 5); err != nil || len(items) != 1 || items[0].Artist != "Artist" {
		t.Fatalf("SearchArtists returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchAlbums("mix", 5); err != nil || len(items) != 1 || items[0].Album != "Album" {
		t.Fatalf("SearchAlbums returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchMusicPlaylists("mix", 5); err != nil || len(items) != 1 || items[0].ID != 3 {
		t.Fatalf("SearchMusicPlaylists returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchVideoPlaylists("mix", 5); err != nil || len(items) != 1 || items[0].ID != 4 {
		t.Fatalf("SearchVideoPlaylists returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchVideos("mix", 5); err != nil || len(items) != 1 || items[0].ID != 5 {
		t.Fatalf("SearchVideos returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchImages("mix", 5); err != nil || len(items) != 1 || items[0].ID != 6 {
		t.Fatalf("SearchImages returned %+v err=%v", items, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestSearchRepositoryErrorPaths(t *testing.T) {
	repository, mock := newSearchRepositoryForTest(t)
	errBoom := errors.New("query failed")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchFilesQuery)).
		WithArgs(singleTermFilesMatchArgs(5)...).
		WillReturnError(errBoom)
	mock.ExpectRollback()

	if _, err := repository.SearchFiles("mix", 5); !errors.Is(err, errBoom) {
		t.Fatalf("expected SearchFiles error, got %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchImagesQuery)).
		WithArgs(singleTermMatchArgsWithFormats(5)...).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()

	if _, err := repository.SearchImages("mix", 5); err == nil {
		t.Fatalf("expected SearchImages scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestSearchRepositoryBlankQuerySkipsDatabase(t *testing.T) {
	repository, mock := newSearchRepositoryForTest(t)

	if items, err := repository.SearchFiles("   ", 5); err != nil || len(items) != 0 {
		t.Fatalf("SearchFiles returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchFolders("", 5); err != nil || len(items) != 0 {
		t.Fatalf("SearchFolders returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchArtists("", 5); err != nil || len(items) != 0 {
		t.Fatalf("SearchArtists returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchAlbums("", 5); err != nil || len(items) != 0 {
		t.Fatalf("SearchAlbums returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchVideos("", 5); err != nil || len(items) != 0 {
		t.Fatalf("SearchVideos returned %+v err=%v", items, err)
	}
	if items, err := repository.SearchImages("", 5); err != nil || len(items) != 0 {
		t.Fatalf("SearchImages returned %+v err=%v", items, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected database calls: %v", err)
	}
}

func TestSearchRepositoryMultiWordQueryBuildsPatternsPerTerm(t *testing.T) {
	repository, mock := newSearchRepositoryForTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchFilesQuery)).
		WithArgs("%holiday%", sqlmock.AnyArg(), "beach holiday", "beach holiday%", "%beach holiday%", sqlmock.AnyArg(), 5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "parent_path", "format", "starred", "size", "updated_at", "is_cold"}))
	mock.ExpectRollback()

	if _, err := repository.SearchFiles("  Beach   Holiday ", 5); err != nil {
		t.Fatalf("SearchFiles error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestSearchTracksScansRowsAndWrapsErrors(t *testing.T) {
	errBoom := errors.New("query failed")
	repository, mock := newSearchRepositoryForTest(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchTracksQuery)).
		WithArgs(singleTermMatchArgsWithFormats(5)...).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "artist", "album", "album_owner", "duration", "path"}).
			AddRow(9, "Mix", "Artist", "Album", "Artist", 120.5, "/media/mix.mp3"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchTracksQuery)).
		WithArgs(singleTermMatchArgsWithFormats(5)...).
		WillReturnError(errBoom)
	mock.ExpectRollback()

	items, err := repository.SearchTracks("mix", 5)
	if err != nil || len(items) != 1 || items[0].FileID != 9 || items[0].Duration != 120.5 {
		t.Fatalf("SearchTracks returned %+v err=%v", items, err)
	}
	if _, err := repository.SearchTracks("mix", 5); !errors.Is(err, errBoom) {
		t.Fatalf("expected SearchTracks error, got %v", err)
	}
	if items, err := repository.SearchTracks("   ", 5); err != nil || len(items) != 0 {
		t.Fatalf("blank SearchTracks returned %+v err=%v", items, err)
	}
}

func TestSearchRepositoryFuzzyPaths(t *testing.T) {
	repository, mock := newSearchRepositoryForTest(t)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CheckFuzzySearchSupportQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchFilesFuzzyQuery)).
		WithArgs("relatoro", 5, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "parent_path", "format", "starred", "size", "updated_at", "is_cold"}).
			AddRow(1, "relatorio.pdf", "/docs/relatorio.pdf", "/docs", ".pdf", false, int64(10), now, false))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchFoldersFuzzyQuery)).
		WithArgs("relatoro", 5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "parent_path", "starred", "size", "updated_at", "is_cold"}).
			AddRow(2, "relatorios", "/docs/relatorios", "/docs", false, int64(0), now, true))
	mock.ExpectRollback()

	if !repository.IsFuzzySearchAvailable() || !repository.IsFuzzySearchAvailable() {
		t.Fatal("expected fuzzy search to be available and probed once")
	}
	files, err := repository.SearchFilesFuzzy("  Relatoro ", 5)
	if err != nil || len(files) != 1 || files[0].Name != "relatorio.pdf" {
		t.Fatalf("unexpected fuzzy files: %+v err=%v", files, err)
	}
	folders, err := repository.SearchFoldersFuzzy("relatoro", 5)
	if err != nil || len(folders) != 1 || !folders[0].IsCold {
		t.Fatalf("unexpected fuzzy folders: %+v err=%v", folders, err)
	}
	if blank, blankErr := repository.SearchFilesFuzzy("   ", 5); blankErr != nil || len(blank) != 0 {
		t.Fatalf("blank fuzzy query must be empty, got %+v err=%v", blank, blankErr)
	}
	if blank, blankErr := repository.SearchFoldersFuzzy("   ", 5); blankErr != nil || len(blank) != 0 {
		t.Fatalf("blank fuzzy query must be empty, got %+v err=%v", blank, blankErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSearchRepositoryFuzzyProbeFailureDisablesFuzzySearch(t *testing.T) {
	repository, mock := newSearchRepositoryForTest(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CheckFuzzySearchSupportQuery)).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if repository.IsFuzzySearchAvailable() {
		t.Fatal("probe failure must disable fuzzy search")
	}
}

func TestSearchRepositoryFuzzyQueryErrorsAreWrapped(t *testing.T) {
	repository, mock := newSearchRepositoryForTest(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchFilesFuzzyQuery)).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.SearchFoldersFuzzyQuery)).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if _, err := repository.SearchFilesFuzzy("abc", 5); err == nil {
		t.Fatal("expected files error")
	}
	if _, err := repository.SearchFoldersFuzzy("abc", 5); err == nil {
		t.Fatal("expected folders error")
	}
}
