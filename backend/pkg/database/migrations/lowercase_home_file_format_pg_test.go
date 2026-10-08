package migrations_test

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database/migrations"
)

func TestLowercaseHomeFileFormatMigration_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	const insertFileQuery = `INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at)
		VALUES ($1, $2, '/media', $3, 1, now(), now())`
	const selectFormatQuery = `SELECT format FROM home_file WHERE name = $1`

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`DELETE FROM home_file WHERE parent_path = '/media'`); err != nil {
			return err
		}
		for _, seed := range []struct{ name, format string }{
			{"IMG_0001.JPG", ".JPG"},
			{"Song.Mp3", ".Mp3"},
			{"clip.mp4", ".mp4"},
		} {
			if _, err := tx.Exec(insertFileQuery, seed.name, "/media/"+seed.name, seed.format); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed home_file: %v", seedErr)
	}
	t.Cleanup(func() {
		_ = dbContext.ExecTx(func(tx *sql.Tx) error {
			_, err := tx.Exec(`DELETE FROM home_file WHERE parent_path = '/media'`)
			return err
		})
	})

	runErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(migrations.LowercaseHomeFileFormatQuery)
		return err
	})
	if runErr != nil {
		t.Fatalf("run migration: %v", runErr)
	}

	expectedFormatsByName := map[string]string{
		"IMG_0001.JPG": ".jpg",
		"Song.Mp3":     ".mp3",
		"clip.mp4":     ".mp4",
	}
	for name, expectedFormat := range expectedFormatsByName {
		var persistedFormat string
		queryErr := dbContext.QueryTx(func(tx *sql.Tx) error {
			return tx.QueryRow(selectFormatQuery, name).Scan(&persistedFormat)
		})
		if queryErr != nil {
			t.Fatalf("read format of %s: %v", name, queryErr)
		}
		if persistedFormat != expectedFormat {
			t.Fatalf("format of %s = %q, want %q", name, persistedFormat, expectedFormat)
		}
	}
}
