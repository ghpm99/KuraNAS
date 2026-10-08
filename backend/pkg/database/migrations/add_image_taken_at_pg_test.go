package migrations_test

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database/migrations"
)

func TestAddImageTakenAtMigrationBackfill_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")
	fileModificationTime := time.Date(2018, 2, 3, 4, 5, 6, 0, time.UTC)

	seeds := []struct {
		name             string
		dateTimeOriginal string
		dateTime         string
		expected         time.Time
	}{
		{"original.jpg", "2020:08:09 10:11:12", "2021:01:01 00:00:00", time.Date(2020, 8, 9, 10, 11, 12, 0, time.UTC)},
		{"datetime_only.jpg", "", "2021:01:01 00:00:01", time.Date(2021, 1, 1, 0, 0, 1, 0, time.UTC)},
		{"zero_original.jpg", "0000:00:00 00:00:00", "2019:05:06 07:08:09", time.Date(2019, 5, 6, 7, 8, 9, 0, time.UTC)},
		{"impossible_day.jpg", "2021:02:31 10:00:00", "", fileModificationTime},
		{"garbage.jpg", "yesterday", "2021-01-01", fileModificationTime},
		{"no_exif.jpg", "", "", fileModificationTime},
	}

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE image_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		for _, seed := range seeds {
			var fileID int
			if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
				VALUES ($1, $2, '/lib', '.jpg', 1, $3, now(), 2, '') RETURNING id`,
				seed.name, "/lib/"+seed.name, fileModificationTime).Scan(&fileID); err != nil {
				return err
			}
			if _, err := tx.Exec(`INSERT INTO image_metadata (file_id, path, datetime_original, datetime, taken_at, created_at)
				VALUES ($1, $2, $3, $4, NULL, now())`, fileID, "/lib/"+seed.name, seed.dateTimeOriginal, seed.dateTime); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	runErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(migrations.AddImageTakenAtQuery)
		return err
	})
	if runErr != nil {
		t.Fatalf("run migration: %v", runErr)
	}

	for _, seed := range seeds {
		var takenAt time.Time
		queryErr := dbContext.QueryTx(func(tx *sql.Tx) error {
			return tx.QueryRow(`SELECT im.taken_at FROM image_metadata im JOIN home_file hf ON hf.id = im.file_id WHERE hf.name = $1`, seed.name).Scan(&takenAt)
		})
		if queryErr != nil {
			t.Fatalf("read taken_at of %s: %v", seed.name, queryErr)
		}
		if !takenAt.Equal(seed.expected) {
			t.Fatalf("taken_at of %s = %v, want %v", seed.name, takenAt, seed.expected)
		}
	}
}
