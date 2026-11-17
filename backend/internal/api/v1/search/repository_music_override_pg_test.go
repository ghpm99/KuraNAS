package search

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/pkg/database"
)

func seedAudioOverride(t *testing.T, dbContext *database.DbContext, fileID int, title string, artist string, album string) {
	t.Helper()
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`INSERT INTO audio_metadata_override (file_id, title, artist, album) VALUES ($1, $2, $3, $4)`,
			fileID, title, artist, album)
		return execErr
	})
	if err != nil {
		t.Fatalf("seed audio override: %v", err)
	}
}

func seedOverriddenTrack(t *testing.T, dbContext *database.DbContext) {
	t.Helper()
	fileID := seedHomeFile(t, dbContext, searchSeedFile{name: "track01.mp3", format: ".mp3", fileType: 2, updatedAt: time.Now().UTC()})
	seedAudioTrackMetadata(t, dbContext, fileID, "Oldtitle", "Oldartist", "Oldalbum")
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`UPDATE audio_metadata SET year = '2020' WHERE file_id = $1`, fileID)
		return execErr
	})
	if err != nil {
		t.Fatalf("seed audio year: %v", err)
	}
	seedAudioOverride(t, dbContext, fileID, "Newtitle", "Newartist", "Newalbum")
}

func TestSearchTracksUsesOverriddenValuesInsteadOfTags_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	seedOverriddenTrack(t, dbContext)

	foundByNewTitle, err := repository.SearchTracks("newtitle", 10)
	if err != nil || len(foundByNewTitle) != 1 || foundByNewTitle[0].Title != "Newtitle" || foundByNewTitle[0].Artist != "Newartist" || foundByNewTitle[0].Album != "Newalbum" {
		t.Fatalf("search by corrected title = %+v err=%v", foundByNewTitle, err)
	}
	for _, replacedTerm := range []string{"oldtitle", "oldartist", "oldalbum"} {
		foundByOldValue, err := repository.SearchTracks(replacedTerm, 10)
		if err != nil || len(foundByOldValue) != 0 {
			t.Fatalf("search by replaced tag %q must find nothing, got %+v err=%v", replacedTerm, foundByOldValue, err)
		}
	}
}

func TestSearchArtistsAndAlbumsUseOverriddenValuesInsteadOfTags_Postgres(t *testing.T) {
	repository, dbContext := newSearchPostgresRepository(t)
	seedOverriddenTrack(t, dbContext)

	artists, err := repository.SearchArtists("newartist", 10)
	if err != nil || len(artists) != 1 || artists[0].Artist != "Newartist" {
		t.Fatalf("search artist by corrected name = %+v err=%v", artists, err)
	}
	oldArtists, err := repository.SearchArtists("oldartist", 10)
	if err != nil || len(oldArtists) != 0 {
		t.Fatalf("search artist by replaced tag must find nothing, got %+v err=%v", oldArtists, err)
	}

	albums, err := repository.SearchAlbums("newalbum", 10)
	if err != nil || len(albums) != 1 || albums[0].Album != "Newalbum" || albums[0].Artist != "Newartist" {
		t.Fatalf("search album by corrected name = %+v err=%v", albums, err)
	}
	oldAlbums, err := repository.SearchAlbums("oldalbum", 10)
	if err != nil || len(oldAlbums) != 0 {
		t.Fatalf("search album by replaced tag must find nothing, got %+v err=%v", oldAlbums, err)
	}
}
