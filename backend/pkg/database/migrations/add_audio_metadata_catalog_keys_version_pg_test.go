package migrations_test

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database/migrations"
)

func TestAddAudioMetadataCatalogKeysVersionMigrationMarksExistingRowsAsOutdated_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist_track, playlist, audio_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`DROP VIEW IF EXISTS audio_metadata_effective`); err != nil {
			return err
		}
		if _, err := tx.Exec(`ALTER TABLE audio_metadata DROP COLUMN IF EXISTS catalog_keys_version`); err != nil {
			return err
		}
		var fileID int
		if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('legacy.mp3', '/m/legacy.mp3', '/m', '.mp3', 1, now(), now(), 2, '') RETURNING id`).Scan(&fileID); err != nil {
			return err
		}
		_, err := tx.Exec(`INSERT INTO audio_metadata (file_id, path, catalog_artist_key) VALUES ($1, '/m/legacy.mp3', 'artist')`, fileID)
		return err
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	for runNumber := 1; runNumber <= 2; runNumber++ {
		runErr := dbContext.ExecTx(func(tx *sql.Tx) error {
			_, err := tx.Exec(migrations.AddAudioMetadataCatalogKeysVersionQuery)
			return err
		})
		if runErr != nil {
			t.Fatalf("run %d: %v", runNumber, runErr)
		}
	}

	var keysVersion int
	var isNullable string
	readErr := dbContext.QueryTx(func(tx *sql.Tx) error {
		if err := tx.QueryRow(`SELECT catalog_keys_version FROM audio_metadata WHERE path = '/m/legacy.mp3'`).Scan(&keysVersion); err != nil {
			return err
		}
		return tx.QueryRow(`SELECT is_nullable FROM information_schema.columns WHERE table_name = 'audio_metadata' AND column_name = 'catalog_keys_version'`).Scan(&isNullable)
	})
	if readErr != nil {
		t.Fatalf("read back: %v", readErr)
	}
	if keysVersion != 0 || isNullable != "NO" {
		t.Fatalf("version = %d nullable = %s, want 0 and NO", keysVersion, isNullable)
	}
}
