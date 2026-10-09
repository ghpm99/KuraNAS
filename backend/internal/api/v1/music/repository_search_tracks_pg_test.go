package music

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestSearchLibraryTracksMatchesMetadataAndPaginates_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_music_search_it")
	repository := NewRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist_track, playlist, audio_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		trackFiles := []struct{ name, title, artist, album string }{
			{"a.mp3", "Bohemian Rhapsody", "Queen", "Opera"},
			{"b.mp3", "Another One", "Queen", "Jazz"},
			{"c.mp3", "Imagine", "Lennon", "Imagine"},
		}
		for _, trackFile := range trackFiles {
			var fileID int
			if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
				VALUES ($1::text, '/m/' || $1::text, '/m', '.mp3', 1, now(), now(), 2, '') RETURNING id`, trackFile.name).Scan(&fileID); err != nil {
				return err
			}
			if _, err := tx.Exec(`INSERT INTO audio_metadata (file_id, path, title, artist, album, created_at) VALUES ($1, '/m/' || $2, $3, $4, $5, now())`,
				fileID, trackFile.name, trackFile.title, trackFile.artist, trackFile.album); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	byArtist, err := repository.SearchLibraryTracks("queen", 1, 10)
	if err != nil || len(byArtist.Items) != 2 {
		t.Fatalf("artist search len=%d err=%v", len(byArtist.Items), err)
	}

	byTwoTerms, err := repository.SearchLibraryTracks("queen bohemian", 1, 10)
	if err != nil || len(byTwoTerms.Items) != 1 {
		t.Fatalf("two-term search len=%d err=%v", len(byTwoTerms.Items), err)
	}

	firstPage, err := repository.SearchLibraryTracks("queen", 1, 1)
	if err != nil || len(firstPage.Items) != 1 || !firstPage.Pagination.HasNext {
		t.Fatalf("first page len=%d hasNext=%v err=%v", len(firstPage.Items), firstPage.Pagination.HasNext, err)
	}

	wildcard, err := repository.SearchLibraryTracks("%", 1, 10)
	if err != nil || len(wildcard.Items) != 0 {
		t.Fatalf("wildcard must be escaped len=%d err=%v", len(wildcard.Items), err)
	}

	blank, err := repository.SearchLibraryTracks("  ", 1, 10)
	if err != nil || len(blank.Items) != 0 {
		t.Fatalf("blank len=%d err=%v", len(blank.Items), err)
	}
}
