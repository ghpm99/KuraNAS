package image

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func seedImageFiles(t *testing.T, execSeed func(tx *sql.Tx) error) *Repository {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_image_it")
	repository := NewRepository(dbContext)
	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE image_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		return execSeed(tx)
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}
	return repository
}

func TestListImagesWithoutMetadata_Postgres(t *testing.T) {
	repository := seedImageFiles(t, func(tx *sql.Tx) error {
		if _, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum, deleted_at)
			VALUES ('missing.jpg', '/img/missing.jpg', '/img', '.jpg', 1, now(), now(), 2, '', NULL),
			       ('indexed.jpg', '/img/indexed.jpg', '/img', '.jpg', 1, now(), now(), 2, '', NULL),
			       ('deleted.png', '/img/deleted.png', '/img', '.png', 1, now(), now(), 2, '', now()),
			       ('song.mp3', '/img/song.mp3', '/img', '.mp3', 1, now(), now(), 2, '', NULL),
			       ('missing2.heic', '/img/missing2.heic', '/img', '.heic', 1, now(), now(), 2, '', NULL)`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO image_metadata (file_id, path) SELECT id, path FROM home_file WHERE name = 'indexed.jpg'`)
		return err
	})

	firstPage, err := repository.ListImagesWithoutMetadata(0, 1)
	if err != nil {
		t.Fatalf("ListImagesWithoutMetadata returned error: %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].Path != "/img/missing.jpg" {
		t.Fatalf("unexpected first page: %+v", firstPage)
	}

	secondPage, err := repository.ListImagesWithoutMetadata(firstPage[0].FileID, 10)
	if err != nil {
		t.Fatalf("ListImagesWithoutMetadata returned error: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].Path != "/img/missing2.heic" {
		t.Fatalf("expected only the remaining active image without metadata, got %+v", secondPage)
	}
}

func TestUpdateAIClassificationLeavesPendingSet_Postgres(t *testing.T) {
	repository := seedImageFiles(t, func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('a.jpg', '/img/a.jpg', '/img', '.jpg', 1, now(), now(), 2, '')`)
		return err
	})

	var fileID int
	if err := repository.Db.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT id FROM home_file`).Scan(&fileID)
	}); err != nil {
		t.Fatalf("select file id: %v", err)
	}
	if err := repository.Db.ExecTx(func(tx *sql.Tx) error {
		_, upsertErr := repository.UpsertImageMetadata(tx, MetadataModel{
			FileId: fileID, Path: "/img/a.jpg",
			Classification: ClassificationModel{Category: ClassificationCategoryOther, Confidence: 0.35},
		})
		return upsertErr
	}); err != nil {
		t.Fatalf("seed metadata: %v", err)
	}

	pending, err := repository.ListPendingAIClassification(AIClassificationConfidenceThreshold, 0, 10)
	if err != nil || len(pending) != 1 || pending[0].MetadataID == 0 {
		t.Fatalf("expected one pending image with metadata id, got %+v err=%v", pending, err)
	}

	aiResult := ClassificationModel{Category: ClassificationCategoryLandscape, Confidence: 0.9, SuggestedName: "mountain_view"}
	if err := repository.UpdateAIClassification(pending[0].FileID, aiResult); err != nil {
		t.Fatalf("UpdateAIClassification returned error: %v", err)
	}

	stillPending, err := repository.ListPendingAIClassification(AIClassificationConfidenceThreshold, 0, 10)
	if err != nil || len(stillPending) != 0 {
		t.Fatalf("expected empty pending set after AI classification, got %+v err=%v", stillPending, err)
	}

	stored, err := repository.GetImageMetadataByID(pending[0].MetadataID)
	if err != nil {
		t.Fatalf("GetImageMetadataByID returned error: %v", err)
	}
	if stored.Classification.Category != ClassificationCategoryLandscape || stored.Classification.AIClassifiedAt == nil {
		t.Fatalf("unexpected stored classification: %+v", stored.Classification)
	}
}

func TestHeuristicReindexKeepsStoredAIClassification_Postgres(t *testing.T) {
	repository := seedImageFiles(t, func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('a.jpg', '/img/a.jpg', '/img', '.jpg', 1, now(), now(), 2, '')`)
		return err
	})

	var fileID int
	if err := repository.Db.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT id FROM home_file`).Scan(&fileID)
	}); err != nil {
		t.Fatalf("select file id: %v", err)
	}

	upsert := func(classification ClassificationModel) MetadataModel {
		var persisted MetadataModel
		err := repository.Db.ExecTx(func(tx *sql.Tx) error {
			var upsertErr error
			persisted, upsertErr = repository.UpsertImageMetadata(tx, MetadataModel{
				FileId: fileID, Path: "/img/a.jpg", Classification: classification,
			})
			return upsertErr
		})
		if err != nil {
			t.Fatalf("UpsertImageMetadata returned error: %v", err)
		}
		return persisted
	}

	first := upsert(ClassificationModel{Category: ClassificationCategoryOther, Confidence: 0.35})
	if err := repository.UpdateAIClassification(fileID, ClassificationModel{Category: ClassificationCategoryArt, Confidence: 0.8}); err != nil {
		t.Fatalf("UpdateAIClassification returned error: %v", err)
	}
	upsert(ClassificationModel{Category: ClassificationCategoryOther, Confidence: 0.35})

	stored, err := repository.GetImageMetadataByID(first.ID)
	if err != nil {
		t.Fatalf("GetImageMetadataByID returned error: %v", err)
	}
	if stored.Classification.Category != ClassificationCategoryArt || stored.Classification.Confidence != 0.8 {
		t.Fatalf("heuristic re-index must not overwrite the AI classification, got %+v", stored.Classification)
	}
}
