package music

import (
	"database/sql"
	"errors"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestGetLibraryTrackByIDOnlyReturnsActiveAudio_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_music_track_by_id_it")
	repository := NewRepository(dbContext)
	service := &Service{Repository: repository}

	var audioFileID, deletedAudioFileID, documentFileID int
	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist_track, playlist, audio_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('song.mp3', '/m/song.mp3', '/m', '.mp3', 1, now(), now(), 2, '') RETURNING id`).Scan(&audioFileID); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO audio_metadata (file_id, path, title, artist, created_at) VALUES ($1, '/m/song.mp3', 'Song', 'Band', now())`, audioFileID); err != nil {
			return err
		}
		if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, deleted_at, type, checksum)
			VALUES ('gone.mp3', '/m/gone.mp3', '/m', '.mp3', 1, now(), now(), now(), 2, '') RETURNING id`).Scan(&deletedAudioFileID); err != nil {
			return err
		}
		return tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('notes.txt', '/m/notes.txt', '/m', '.txt', 1, now(), now(), 2, '') RETURNING id`).Scan(&documentFileID)
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	track, err := service.GetLibraryTrackByID(audioFileID)
	if err != nil {
		t.Fatalf("GetLibraryTrackByID audio: %v", err)
	}
	metadata, isAudioMetadata := track.Metadata.(AudioMetadataModel)
	if track.ID != audioFileID || !isAudioMetadata || metadata.Title != "Song" || metadata.Artist != "Band" {
		t.Fatalf("unexpected track %+v", track)
	}

	for name, fileID := range map[string]int{"deleted": deletedAudioFileID, "not audio": documentFileID, "unknown": 999999} {
		if _, err := service.GetLibraryTrackByID(fileID); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("%s: expected sql.ErrNoRows, got %v", name, err)
		}
	}
}
