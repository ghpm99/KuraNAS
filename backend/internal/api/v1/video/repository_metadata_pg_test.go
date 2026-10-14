package video

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

func TestGetVideoMetadataByIDReturnsEveryColumn_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewVideoMetadataRepository(dbContext)

	var fileID int
	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		return tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('full.mp4', '/v/full.mp4', '/v', '.mp4', 1, now(), now(), 2, '') RETURNING id`).Scan(&fileID)
	})
	if seedErr != nil {
		t.Fatalf("seed file: %v", seedErr)
	}

	expectedMetadata := VideoMetadataModel{
		FileId:                fileID,
		Path:                  "/v/full.mp4",
		FormatName:            "mov,mp4",
		Size:                  "1000",
		Duration:              "60.5",
		Width:                 1920,
		Height:                1080,
		FrameRate:             29.5,
		NbFrames:              1800,
		BitRate:               "2500",
		CodecName:             "h264",
		CodecLongName:         "H.264 / AVC",
		PixFmt:                "yuv420p",
		Level:                 4,
		Profile:               "main",
		AspectRatio:           "16:9",
		AudioCodec:            "aac",
		AudioChannels:         2,
		AudioSampleRate:       "44100",
		AudioBitRate:          "128000",
		Classification:        "movie",
		ClassificationVersion: 3,
	}

	var upsertedMetadata VideoMetadataModel
	upsertErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		var err error
		upsertedMetadata, err = repository.UpsertVideoMetadata(tx, expectedMetadata)
		return err
	})
	if upsertErr != nil {
		t.Fatalf("upsert: %v", upsertErr)
	}

	storedMetadata, err := repository.GetVideoMetadataByID(upsertedMetadata.ID)
	if err != nil {
		t.Fatalf("GetVideoMetadataByID: %v", err)
	}

	expectedMetadata.ID = upsertedMetadata.ID
	expectedMetadata.CreatedAt = upsertedMetadata.CreatedAt
	if !storedMetadata.CreatedAt.Truncate(time.Millisecond).Equal(expectedMetadata.CreatedAt.Truncate(time.Millisecond)) || storedMetadata.CreatedAt.IsZero() {
		t.Fatalf("created_at mismatch: stored=%v expected=%v", storedMetadata.CreatedAt, expectedMetadata.CreatedAt)
	}
	storedMetadata.CreatedAt = expectedMetadata.CreatedAt
	if storedMetadata != expectedMetadata {
		t.Fatalf("metadata mismatch:\nstored   %+v\nexpected %+v", storedMetadata, expectedMetadata)
	}
}

func TestGetVideoMetadataByIDToleratesNullColumns_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewVideoMetadataRepository(dbContext)

	var metadataID int
	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('sparse.mp4', '/v/sparse.mp4', '/v', '.mp4', 1, now(), now(), 2, '')`); err != nil {
			return err
		}
		return tx.QueryRow(`INSERT INTO video_metadata (file_id, path, created_at)
			SELECT id, path, NULL FROM home_file RETURNING id`).Scan(&metadataID)
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	storedMetadata, err := repository.GetVideoMetadataByID(metadataID)
	if err != nil {
		t.Fatalf("GetVideoMetadataByID: %v", err)
	}
	if storedMetadata.ID != metadataID || storedMetadata.Path != "/v/sparse.mp4" || storedMetadata.Width != 0 || storedMetadata.CodecName != "" {
		t.Fatalf("unexpected sparse metadata: %+v", storedMetadata)
	}
}
