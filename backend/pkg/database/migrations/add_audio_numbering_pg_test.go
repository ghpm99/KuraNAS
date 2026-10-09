package migrations_test

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database/migrations"
)

func TestAddAudioMetadataDiscTrackNumbersMigrationBackfill_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	seeds := []struct {
		name          string
		trackNumber   string
		expectedNo    sql.NullInt64
		expectedTotal sql.NullInt64
	}{
		{"with_total.mp3", "3/12", sql.NullInt64{Int64: 3, Valid: true}, sql.NullInt64{Int64: 12, Valid: true}},
		{"plain.mp3", "7", sql.NullInt64{Int64: 7, Valid: true}, sql.NullInt64{}},
		{"garbage.mp3", "n/a", sql.NullInt64{}, sql.NullInt64{}},
	}

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist_track, playlist, audio_metadata, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		for _, seed := range seeds {
			var fileID int
			if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
				VALUES ($1, $2, '/m', '.mp3', 1, now(), now(), 2, '') RETURNING id`, seed.name, "/m/"+seed.name).Scan(&fileID); err != nil {
				return err
			}
			if _, err := tx.Exec(`INSERT INTO audio_metadata (file_id, path, track_number) VALUES ($1, $2, $3)`,
				fileID, "/m/"+seed.name, seed.trackNumber); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	runErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(migrations.AddAudioMetadataDiscTrackNumbersQuery)
		return err
	})
	if runErr != nil {
		t.Fatalf("run migration: %v", runErr)
	}

	for _, seed := range seeds {
		var trackNo, trackTotal sql.NullInt64
		queryErr := dbContext.QueryTx(func(tx *sql.Tx) error {
			return tx.QueryRow(`SELECT track_no, track_total FROM audio_metadata WHERE path = $1`, "/m/"+seed.name).Scan(&trackNo, &trackTotal)
		})
		if queryErr != nil {
			t.Fatalf("%s: read back: %v", seed.name, queryErr)
		}
		if trackNo != seed.expectedNo || trackTotal != seed.expectedTotal {
			t.Errorf("%s: got %v/%v, want %v/%v", seed.name, trackNo, trackTotal, seed.expectedNo, seed.expectedTotal)
		}
	}
}
