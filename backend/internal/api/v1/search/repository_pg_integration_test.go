package search

import (
	"database/sql"
	"slices"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
)

type searchSeedFile struct {
	name         string
	format       string
	fileType     int
	starred      bool
	updatedAt    time.Time
	isDeleted    bool
	physicalPath string
}

func seedHomeFile(t *testing.T, dbContext *database.DbContext, seed searchSeedFile) int {
	t.Helper()
	var fileID int
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		deletedAt := sql.NullTime{Time: seed.updatedAt, Valid: seed.isDeleted}
		return tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, starred, deleted_at, physical_path)
			VALUES ($1, $2, '/lib', $3, 1, $4, now(), $5, '', $6, $7, NULLIF($8, '')) RETURNING id`,
			seed.name, "/lib/"+seed.name, seed.format, seed.updatedAt, seed.fileType, seed.starred, deletedAt, seed.physicalPath).Scan(&fileID)
	})
	if err != nil {
		t.Fatalf("seed %q: %v", seed.name, err)
	}
	return fileID
}

func seedAudioMetadata(t *testing.T, dbContext *database.DbContext, fileID int, artist string, albumArtist string, album string) {
	t.Helper()
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`INSERT INTO audio_metadata (file_id, path, artist, album_artist, album, year)
			VALUES ($1, $2, $3, $4, $5, '2020')`, fileID, "/lib/audio", artist, albumArtist, album)
		return execErr
	})
	if err != nil {
		t.Fatalf("seed audio metadata: %v", err)
	}
}

func seedAudioTrackMetadata(t *testing.T, dbContext *database.DbContext, fileID int, title string, artist string, album string) {
	t.Helper()
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`INSERT INTO audio_metadata (file_id, path, title, artist, album, length)
			VALUES ($1, $2, $3, $4, $5, 215.5)`, fileID, "/lib/track", title, artist, album)
		return execErr
	})
	if err != nil {
		t.Fatalf("seed audio track metadata: %v", err)
	}
}

func trackTitles(results []TrackResultModel) []string {
	titles := make([]string, 0, len(results))
	for _, result := range results {
		titles = append(titles, result.Title)
	}
	return titles
}

func seedImageMetadata(t *testing.T, dbContext *database.DbContext, fileID int, aiSearchText string) {
	t.Helper()
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`INSERT INTO image_metadata (file_id, path, created_at, ai_search_text)
			VALUES ($1, $2, now(), $3)`, fileID, "/lib/image", aiSearchText)
		return execErr
	})
	if err != nil {
		t.Fatalf("seed image metadata: %v", err)
	}
}

func newSearchPostgresRepository(t *testing.T) (*Repository, *database.DbContext) {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_search_it")
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`TRUNCATE audio_metadata, image_metadata, home_file RESTART IDENTITY CASCADE`)
		return execErr
	})
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return NewRepository(dbContext), dbContext
}

func fileNames(results []FileResultModel) []string {
	names := make([]string, 0, len(results))
	for _, result := range results {
		names = append(names, result.Name)
	}
	return names
}

func TestSearchFilesMatchesAllTermsInAnyOrder_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	seedHomeFile(t, dbContext, searchSeedFile{name: "Holiday Beach Trip.txt", format: ".txt", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "trip_beach holiday final.txt", format: ".txt", fileType: 2, updatedAt: now.Add(-time.Hour)})
	seedHomeFile(t, dbContext, searchSeedFile{name: "beach only.txt", format: ".txt", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "holiday plans.txt", format: ".txt", fileType: 2, updatedAt: now})

	results, err := repository.SearchFiles("holiday  BEACH", 10)
	if err != nil {
		t.Fatalf("SearchFiles: %v", err)
	}
	expected := []string{"Holiday Beach Trip.txt", "trip_beach holiday final.txt"}
	if got := fileNames(results); !slices.Equal(got, expected) {
		t.Fatalf("names = %v, want %v", got, expected)
	}
}

func TestSearchFilesTreatsLikeWildcardsAsLiterals_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	seedHomeFile(t, dbContext, searchSeedFile{name: "100%_real.txt", format: ".txt", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "100abreal.txt", format: ".txt", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "path\\name.txt", format: ".txt", fileType: 2, updatedAt: now})

	results, err := repository.SearchFiles("100%_", 10)
	if err != nil {
		t.Fatalf("SearchFiles: %v", err)
	}
	if got := fileNames(results); !slices.Equal(got, []string{"100%_real.txt"}) {
		t.Fatalf("names = %v", got)
	}

	results, err = repository.SearchFiles(`h\n`, 10)
	if err != nil {
		t.Fatalf("SearchFiles backslash: %v", err)
	}
	if got := fileNames(results); !slices.Equal(got, []string{"path\\name.txt"}) {
		t.Fatalf("backslash names = %v", got)
	}
}

func TestSearchFilesRanksExactThenPrefixThenSubstringThenStarred_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	seedHomeFile(t, dbContext, searchSeedFile{name: "my report final", format: ".txt", fileType: 2, updatedAt: now, starred: true})
	seedHomeFile(t, dbContext, searchSeedFile{name: "report old", format: ".txt", fileType: 2, updatedAt: now.Add(-time.Hour)})
	seedHomeFile(t, dbContext, searchSeedFile{name: "report", format: ".txt", fileType: 2, updatedAt: now.Add(-2 * time.Hour)})
	seedHomeFile(t, dbContext, searchSeedFile{name: "report new", format: ".txt", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "annual REPORT", format: ".txt", fileType: 2, updatedAt: now.Add(-3 * time.Hour), starred: true})

	results, err := repository.SearchFiles("Report", 10)
	if err != nil {
		t.Fatalf("SearchFiles: %v", err)
	}
	expected := []string{"report", "report new", "report old", "my report final", "annual REPORT"}
	if got := fileNames(results); !slices.Equal(got, expected) {
		t.Fatalf("names = %v, want %v", got, expected)
	}
}

func TestSearchFilesExcludesDeletedRowsAndHonorsLimit_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	seedHomeFile(t, dbContext, searchSeedFile{name: "ghost report", format: ".txt", fileType: 2, updatedAt: now, isDeleted: true})
	seedHomeFile(t, dbContext, searchSeedFile{name: "report a", format: ".txt", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "report b", format: ".txt", fileType: 2, updatedAt: now.Add(-time.Minute)})
	seedHomeFile(t, dbContext, searchSeedFile{name: "report folder", format: "", fileType: 1, updatedAt: now})

	results, err := repository.SearchFiles("report", 1)
	if err != nil {
		t.Fatalf("SearchFiles: %v", err)
	}
	if got := fileNames(results); !slices.Equal(got, []string{"report a"}) {
		t.Fatalf("names = %v", got)
	}

	folders, err := repository.SearchFolders("report", 10)
	if err != nil {
		t.Fatalf("SearchFolders: %v", err)
	}
	if len(folders) != 1 || folders[0].Name != "report folder" {
		t.Fatalf("folders = %+v", folders)
	}
}

func TestSearchFilesIgnoresPathMatches_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	seedHomeFile(t, dbContext, searchSeedFile{name: "unrelated.txt", format: ".txt", fileType: 2, updatedAt: time.Now().UTC()})

	results, err := repository.SearchFiles("lib", 10)
	if err != nil {
		t.Fatalf("SearchFiles: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no path-only matches, got %v", fileNames(results))
	}
}

func TestSearchVideosFiltersByFormatAndTerms_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	seedHomeFile(t, dbContext, searchSeedFile{name: "Summer Trip.mp4", format: ".mp4", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "Summer Trip.txt", format: ".txt", fileType: 2, updatedAt: now})

	results, err := repository.SearchVideos("trip summer", 10)
	if err != nil {
		t.Fatalf("SearchVideos: %v", err)
	}
	if len(results) != 1 || results[0].Name != "Summer Trip.mp4" {
		t.Fatalf("videos = %+v", results)
	}
}

func TestSearchImagesMatchesNameOrCaption_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	byName := seedHomeFile(t, dbContext, searchSeedFile{name: "sunset.jpg", format: ".jpg", fileType: 2, updatedAt: now})
	byCaption := seedHomeFile(t, dbContext, searchSeedFile{name: "IMG_0001.jpg", format: ".jpg", fileType: 2, updatedAt: now.Add(-time.Hour)})
	unrelated := seedHomeFile(t, dbContext, searchSeedFile{name: "IMG_0002.jpg", format: ".jpg", fileType: 2, updatedAt: now})
	deleted := seedHomeFile(t, dbContext, searchSeedFile{name: "IMG_0003.jpg", format: ".jpg", fileType: 2, updatedAt: now, isDeleted: true})
	seedImageMetadata(t, dbContext, byName, "a photo")
	seedImageMetadata(t, dbContext, byCaption, "a golden sunset over the sea")
	seedImageMetadata(t, dbContext, unrelated, "a dog")
	seedImageMetadata(t, dbContext, deleted, "sunset deleted")

	results, err := repository.SearchImages("SUNSET golden", 10)
	if err != nil {
		t.Fatalf("SearchImages: %v", err)
	}
	if len(results) != 1 || results[0].ID != byCaption {
		t.Fatalf("multi-term caption results = %+v", results)
	}

	results, err = repository.SearchImages("sunset", 10)
	if err != nil {
		t.Fatalf("SearchImages: %v", err)
	}
	ids := []int{}
	for _, result := range results {
		ids = append(ids, result.ID)
	}
	if !slices.Equal(ids, []int{byName, byCaption}) {
		t.Fatalf("ids = %v, want %v", ids, []int{byName, byCaption})
	}
}

func TestSearchArtistsAndAlbumsGroupMatchingTracks_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	first := seedHomeFile(t, dbContext, searchSeedFile{name: "one.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	second := seedHomeFile(t, dbContext, searchSeedFile{name: "two.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	other := seedHomeFile(t, dbContext, searchSeedFile{name: "three.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	compilation := seedHomeFile(t, dbContext, searchSeedFile{name: "four.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	seedAudioMetadata(t, dbContext, first, "Pink Floyd", "", "The Wall")
	seedAudioMetadata(t, dbContext, second, "Pink Floyd", "Pink Floyd", "The Wall")
	seedAudioMetadata(t, dbContext, other, "Someone Else", "", "Other Album")
	seedAudioMetadata(t, dbContext, compilation, "Pink Floyd", "Various", "Mix")

	artists, err := repository.SearchArtists("floyd PINK", 10)
	if err != nil {
		t.Fatalf("SearchArtists: %v", err)
	}
	if len(artists) != 1 || artists[0].Artist != "Pink Floyd" || artists[0].TrackCount != 2 {
		t.Fatalf("artists = %+v", artists)
	}

	albums, err := repository.SearchAlbums("wall", 10)
	if err != nil {
		t.Fatalf("SearchAlbums: %v", err)
	}
	if len(albums) != 1 || albums[0].Album != "The Wall" || albums[0].TrackCount != 2 {
		t.Fatalf("albums = %+v", albums)
	}
}

func TestSearchFilesExcludesMediaFormats_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	seedHomeFile(t, dbContext, searchSeedFile{name: "holiday notes.txt", format: ".txt", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "holiday song.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "holiday clip.mp4", format: ".mp4", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "holiday photo.jpg", format: ".jpg", fileType: 2, updatedAt: now})

	results, err := repository.SearchFiles("holiday", 10)
	if err != nil {
		t.Fatalf("SearchFiles: %v", err)
	}
	if got := fileNames(results); !slices.Equal(got, []string{"holiday notes.txt"}) {
		t.Fatalf("names = %v", got)
	}
}

func TestSearchTracksMatchesTitleArtistAlbumAndFileName_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	byTitle := seedHomeFile(t, dbContext, searchSeedFile{name: "01.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	byArtist := seedHomeFile(t, dbContext, searchSeedFile{name: "02.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	byAlbum := seedHomeFile(t, dbContext, searchSeedFile{name: "03.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "floyd untagged.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	other := seedHomeFile(t, dbContext, searchSeedFile{name: "04.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	seedHomeFile(t, dbContext, searchSeedFile{name: "floyd video.mp4", format: ".mp4", fileType: 2, updatedAt: now})
	deleted := seedHomeFile(t, dbContext, searchSeedFile{name: "05.mp3", format: ".mp3", fileType: 2, updatedAt: now, isDeleted: true})
	seedAudioTrackMetadata(t, dbContext, byTitle, "Floyd Song", "Someone", "Other")
	seedAudioTrackMetadata(t, dbContext, byArtist, "Money", "Pink Floyd", "Other")
	seedAudioTrackMetadata(t, dbContext, byAlbum, "Time", "Someone", "Floyd Hits")
	seedAudioTrackMetadata(t, dbContext, other, "Unrelated", "Nobody", "Nothing")
	seedAudioTrackMetadata(t, dbContext, deleted, "Floyd Deleted", "Nobody", "Nothing")

	results, err := repository.SearchTracks("floyd", 10)
	if err != nil {
		t.Fatalf("SearchTracks: %v", err)
	}
	titles := trackTitles(results)
	expectedTitles := []string{"Floyd Song", "floyd untagged.mp3", "Money", "Time"}
	slices.Sort(titles)
	slices.Sort(expectedTitles)
	if !slices.Equal(titles, expectedTitles) {
		t.Fatalf("titles = %v, want %v", titles, expectedTitles)
	}
	for _, result := range results {
		if result.Title == "Money" && (result.Artist != "Pink Floyd" || result.Album != "Other" || result.AlbumOwner != "Pink Floyd" || result.Duration != 215.5 || result.Path != "/lib/02.mp3") {
			t.Fatalf("money track = %+v", result)
		}
		if result.Title == "floyd untagged.mp3" && (result.Artist != "" || result.Duration != 0) {
			t.Fatalf("untagged track = %+v", result)
		}
	}
}

func TestSearchTracksRanksExactThenPrefixThenSubstringAndHonorsLimit_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	titles := []string{"My Love Story", "Love Story", "Love", "Lovely Day"}
	for _, title := range titles {
		fileID := seedHomeFile(t, dbContext, searchSeedFile{name: title + ".mp3", format: ".mp3", fileType: 2, updatedAt: now})
		seedAudioTrackMetadata(t, dbContext, fileID, title, "Artist", "Album")
	}

	results, err := repository.SearchTracks("love", 10)
	if err != nil {
		t.Fatalf("SearchTracks: %v", err)
	}
	if got := trackTitles(results); !slices.Equal(got, []string{"Love", "Love Story", "Lovely Day", "My Love Story"}) {
		t.Fatalf("ranked titles = %v", got)
	}

	limited, err := repository.SearchTracks("love", 2)
	if err != nil || len(limited) != 2 {
		t.Fatalf("limited = %+v err=%v", limited, err)
	}
}

func TestSearchTracksRequiresAllTermsAcrossFields_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC()
	match := seedHomeFile(t, dbContext, searchSeedFile{name: "a.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	artistOnly := seedHomeFile(t, dbContext, searchSeedFile{name: "b.mp3", format: ".mp3", fileType: 2, updatedAt: now})
	seedAudioTrackMetadata(t, dbContext, match, "Money", "Pink Floyd", "Dark Side")
	seedAudioTrackMetadata(t, dbContext, artistOnly, "Time", "Pink Floyd", "Dark Side")

	results, err := repository.SearchTracks("floyd MONEY", 10)
	if err != nil {
		t.Fatalf("SearchTracks: %v", err)
	}
	if len(results) != 1 || results[0].FileID != match {
		t.Fatalf("results = %+v", results)
	}
}

func TestSearchFilesReportsSizeUpdatedAtAndColdTier_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	now := time.Now().UTC().Truncate(time.Second)
	seedHomeFile(t, dbContext, searchSeedFile{name: "report cold.txt", format: ".txt", fileType: 2, updatedAt: now, physicalPath: "/cold/report cold.txt"})
	seedHomeFile(t, dbContext, searchSeedFile{name: "report hot.txt", format: ".txt", fileType: 2, updatedAt: now.Add(-time.Hour)})

	results, err := repository.SearchFiles("report", 10)
	if err != nil {
		t.Fatalf("SearchFiles: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %+v", results)
	}
	coldFile, hotFile := results[0], results[1]
	if coldFile.Name != "report cold.txt" || !coldFile.IsCold || coldFile.Size != 1 || !coldFile.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected cold file: %+v", coldFile)
	}
	if hotFile.IsCold {
		t.Fatalf("hot file reported as cold: %+v", hotFile)
	}
}
