package music

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestMusicListingsTolerateMissingAndNullMetadata_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_music_it")
	repository := NewRepository(dbContext)

	var bareFileID, sparseFileID, playlistID int
	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist_track, playlist, audio_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('bare.mp3', '/m/bare.mp3', '/m', '.mp3', 1, now(), now(), 2, '') RETURNING id`).Scan(&bareFileID); err != nil {
			return err
		}
		if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('sparse.mp3', '/m/sparse.mp3', '/m', '.mp3', 1, now(), now(), 2, '') RETURNING id`).Scan(&sparseFileID); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO audio_metadata (file_id, path, created_at) VALUES ($1, '/m/sparse.mp3', NULL)`, sparseFileID); err != nil {
			return err
		}
		if err := tx.QueryRow(`INSERT INTO playlist (name) VALUES ('p') RETURNING id`).Scan(&playlistID); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO playlist_track (playlist_id, file_id, position) VALUES ($1, $2, 1), ($1, $3, 2)`,
			playlistID, bareFileID, sparseFileID)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	musicPage, err := repository.GetMusic(1, 10)
	if err != nil || len(musicPage.Items) != 2 {
		t.Fatalf("GetMusic len=%d err=%v", len(musicPage.Items), err)
	}

	libraryPage, err := repository.GetLibraryTracks(1, 10)
	if err != nil || len(libraryPage.Items) != 2 {
		t.Fatalf("GetLibraryTracks len=%d err=%v", len(libraryPage.Items), err)
	}

	libraryFiles, err := repository.GetLibraryFilesByIDs([]int{bareFileID, sparseFileID})
	if err != nil || len(libraryFiles) != 2 {
		t.Fatalf("GetLibraryFilesByIDs len=%d err=%v", len(libraryFiles), err)
	}

	playlistPage, err := repository.GetPlaylistTracks(playlistID, 1, 10)
	if err != nil || len(playlistPage.Items) != 2 {
		t.Fatalf("GetPlaylistTracks len=%d err=%v", len(playlistPage.Items), err)
	}
	for _, track := range playlistPage.Items {
		if track.MetadataTitle != "" || track.MetadataBitrate != 0 || track.MetadataLength != 0 {
			t.Fatalf("expected zero metadata values, got %+v", track)
		}
	}
}
