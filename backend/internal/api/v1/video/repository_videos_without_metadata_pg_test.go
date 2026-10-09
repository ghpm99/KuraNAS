package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestListVideosWithoutMetadata_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewVideoMetadataRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at)
			VALUES ('missing.mkv', '/v/missing.mkv', '/v', '.mkv', 1, now(), now(), 2, '', NULL),
			       ('indexed.mp4', '/v/indexed.mp4', '/v', '.mp4', 1, now(), now(), 2, '', NULL),
			       ('deleted.avi', '/v/deleted.avi', '/v', '.avi', 1, now(), now(), 2, '', now()),
			       ('song.mp3', '/v/song.mp3', '/v', '.mp3', 1, now(), now(), 2, '', NULL),
			       ('missing2.wmv', '/v/missing2.wmv', '/v', '.wmv', 1, now(), now(), 2, '', NULL)`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO video_metadata (file_id, path) SELECT id, path FROM home_file WHERE name = 'indexed.mp4'`)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	firstPage, err := repository.ListVideosWithoutMetadata(0, 1)
	if err != nil {
		t.Fatalf("ListVideosWithoutMetadata returned error: %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].Path != "/v/missing.mkv" {
		t.Fatalf("unexpected first page: %+v", firstPage)
	}

	secondPage, err := repository.ListVideosWithoutMetadata(firstPage[0].FileID, 10)
	if err != nil {
		t.Fatalf("ListVideosWithoutMetadata returned error: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].Path != "/v/missing2.wmv" {
		t.Fatalf("expected only the remaining active video without metadata, got %+v", secondPage)
	}
}
