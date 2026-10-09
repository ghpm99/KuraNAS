package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestGetVideosToleratesMissingAndNullMetadata_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('bare.mp4', '/v/bare.mp4', '/v', '.mp4', 1, now(), now(), 2, ''),
			       ('sparse.mp4', '/v/sparse.mp4', '/v', '.mp4', 1, now(), now(), 2, '')`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO video_metadata (file_id, path, created_at)
			SELECT id, path, NULL FROM home_file WHERE name = 'sparse.mp4'`)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	page, err := repository.GetVideos(1, 10)
	if err != nil {
		t.Fatalf("GetVideos returned error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(page.Items))
	}
	for _, videoFile := range page.Items {
		videoMetadata, isVideoMetadata := videoFile.Metadata.(VideoMetadataModel)
		if !isVideoMetadata {
			t.Fatalf("unexpected metadata type %T", videoFile.Metadata)
		}
		if videoMetadata.Width != 0 || videoMetadata.CodecName != "" || videoMetadata.Duration != "" {
			t.Fatalf("expected zero values for %s, got %+v", videoFile.Name, videoMetadata)
		}
	}
}
