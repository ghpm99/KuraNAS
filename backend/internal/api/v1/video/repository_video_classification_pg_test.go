package video

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestVideoClassificationPersistenceAndBackfillSelection_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewVideoMetadataRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE video_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at)
			VALUES ('film.mkv', '/v/film.mkv', '/v', '.mkv', 1, now(), now(), 2, '', NULL),
			       ('legacy.mkv', '/v/legacy.mkv', '/v', '.mkv', 1, now(), now(), 2, '', NULL),
			       ('stale.mkv', '/v/stale.mkv', '/v', '.mkv', 1, now(), now(), 2, '', NULL),
			       ('gone.mkv', '/v/gone.mkv', '/v', '.mkv', 1, now(), now(), 2, '', now())`)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	upsertErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		film := VideoMetadataModel{FileId: 1, Path: "/v/film.mkv", Duration: "7200.0", Height: 1080}
		film.ApplyClassification("film.mkv", "/v/film.mkv", "/v")
		if _, err := repository.UpsertVideoMetadata(tx, film); err != nil {
			return err
		}
		if _, err := repository.UpsertVideoMetadata(tx, VideoMetadataModel{FileId: 2, Path: "/v/legacy.mkv", Duration: "7200.0", Height: 1080}); err != nil {
			return err
		}
		if _, err := repository.UpsertVideoMetadata(tx, VideoMetadataModel{FileId: 3, Path: "/v/stale.mkv", Classification: "clip", ClassificationVersion: CurrentVideoClassificationVersion - 1}); err != nil {
			return err
		}
		_, err := repository.UpsertVideoMetadata(tx, VideoMetadataModel{FileId: 4, Path: "/v/gone.mkv"})
		return err
	})
	if upsertErr != nil {
		t.Fatalf("upsert: %v", upsertErr)
	}

	var persistedClassification string
	var persistedVersion int
	row := dbContext.GetDatabase().QueryRow(`SELECT classification, classification_version FROM video_metadata WHERE file_id = 1`)
	if err := row.Scan(&persistedClassification, &persistedVersion); err != nil {
		t.Fatalf("read persisted classification: %v", err)
	}
	if persistedClassification != "movie" || persistedVersion != CurrentVideoClassificationVersion {
		t.Fatalf("expected movie at current version, got %q v%d", persistedClassification, persistedVersion)
	}

	firstPage, err := repository.ListVideosPendingClassification(0, 1)
	if err != nil {
		t.Fatalf("ListVideosPendingClassification returned error: %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].Name != "legacy.mkv" || firstPage[0].Height != 1080 || firstPage[0].Duration != "7200.0" {
		t.Fatalf("unexpected first pending page: %+v", firstPage)
	}

	secondPage, err := repository.ListVideosPendingClassification(firstPage[0].MetadataID, 10)
	if err != nil {
		t.Fatalf("ListVideosPendingClassification returned error: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].Name != "stale.mkv" {
		t.Fatalf("expected only the stale-version active row, got %+v", secondPage)
	}

	if err := repository.UpdateVideoClassification(firstPage[0].MetadataID, "movie"); err != nil {
		t.Fatalf("UpdateVideoClassification: %v", err)
	}
	if err := repository.UpdateVideoClassification(secondPage[0].MetadataID, "clip"); err != nil {
		t.Fatalf("UpdateVideoClassification: %v", err)
	}

	remaining, err := repository.ListVideosPendingClassification(0, 10)
	if err != nil {
		t.Fatalf("ListVideosPendingClassification returned error: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected nothing pending after backfill, got %+v", remaining)
	}
}
