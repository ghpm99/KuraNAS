package image

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestGetImagesToleratesMissingAndNullMetadata_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_image_it")
	repository := NewRepository(dbContext)

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE image_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('bare.jpg', '/img/bare.jpg', '/img', '.jpg', 1, now(), now(), 2, ''),
			       ('sparse.jpg', '/img/sparse.jpg', '/img', '.jpg', 1, now(), now(), 2, '')`); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO image_metadata (file_id, path, created_at)
			SELECT id, path, NULL FROM home_file WHERE name = 'sparse.jpg'`)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	page, err := repository.GetImages(1, 10, ImageGroupByDate)
	if err != nil {
		t.Fatalf("GetImages returned error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 images, got %d", len(page.Items))
	}

	for _, imageFile := range page.Items {
		imageMetadata, isImageMetadata := imageFile.Metadata.(MetadataModel)
		if !isImageMetadata {
			t.Fatalf("unexpected metadata type %T", imageFile.Metadata)
		}
		if imageMetadata.Width != 0 || imageMetadata.Make != "" || imageMetadata.GPSLatitude != 0 {
			t.Fatalf("expected zero values for %s, got %+v", imageFile.Name, imageMetadata)
		}
		if imageMetadata.Classification.Category != ClassificationCategoryOther {
			t.Fatalf("expected category other for %s, got %q", imageFile.Name, imageMetadata.Classification.Category)
		}
	}
}
