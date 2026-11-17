package migrations_test

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
)

func TestCreateTakeoutPhotoSidecarMigrationShape_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE takeout_photo_sidecar`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO takeout_photo_sidecar (path) VALUES ('/lib/takeout/a.jpg')`); err != nil {
			return err
		}

		var description, latitude, longitude sql.NullString
		var hasCreatedAt bool
		if err := tx.QueryRow(`SELECT description, latitude::text, longitude::text, created_at IS NOT NULL
			FROM takeout_photo_sidecar WHERE path = '/lib/takeout/a.jpg'`).Scan(&description, &latitude, &longitude, &hasCreatedAt); err != nil {
			return err
		}
		if description.Valid || latitude.Valid || longitude.Valid {
			t.Fatalf("description, latitude and longitude must default to NULL, got %+v %+v %+v", description, latitude, longitude)
		}
		if !hasCreatedAt {
			t.Fatal("created_at must default to the insertion time")
		}

		_, duplicateErr := tx.Exec(`INSERT INTO takeout_photo_sidecar (path) VALUES ('/lib/takeout/a.jpg')`)
		if duplicateErr == nil {
			t.Fatal("path must be the primary key and reject a duplicate")
		}
		return duplicateErr
	})
	if err == nil {
		t.Fatal("expected the duplicate insert to abort the transaction")
	}
}
