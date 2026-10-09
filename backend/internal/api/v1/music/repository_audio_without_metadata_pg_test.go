package music

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestListAudioWithoutMetadata_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_music_it")
	repository := NewAudioMetadataRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist_track, playlist, audio_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at)
			VALUES ('missing.m4a', '/m/missing.m4a', '/m', '.m4a', 1, now(), now(), 2, '', NULL),
			       ('indexed.mp3', '/m/indexed.mp3', '/m', '.mp3', 1, now(), now(), 2, '', NULL),
			       ('deleted.ogg', '/m/deleted.ogg', '/m', '.ogg', 1, now(), now(), 2, '', now()),
			       ('cover.jpg', '/m/cover.jpg', '/m', '.jpg', 1, now(), now(), 2, '', NULL),
			       ('missing2.opus', '/m/missing2.opus', '/m', '.opus', 1, now(), now(), 2, '', NULL)`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO audio_metadata (file_id, path) SELECT id, path FROM home_file WHERE name = 'indexed.mp3'`)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	firstPage, err := repository.ListAudioWithoutMetadata(0, 1)
	if err != nil {
		t.Fatalf("ListAudioWithoutMetadata returned error: %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].Path != "/m/missing.m4a" {
		t.Fatalf("unexpected first page: %+v", firstPage)
	}

	secondPage, err := repository.ListAudioWithoutMetadata(firstPage[0].FileID, 10)
	if err != nil {
		t.Fatalf("ListAudioWithoutMetadata returned error: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].Path != "/m/missing2.opus" {
		t.Fatalf("expected only the remaining active audio without metadata, got %+v", secondPage)
	}
}
