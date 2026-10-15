package migrations_test

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database/migrations"
)

func TestAddHomeFileStarredAtMigrationBackfill_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")
	lastInteraction := time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC)

	seeds := []struct {
		name            string
		isStarred       bool
		lastInteraction any
		expected        any
	}{
		{"starred_interacted.txt", true, lastInteraction, lastInteraction},
		{"starred_never_interacted.txt", true, nil, updatedAt},
		{"not_starred.txt", false, lastInteraction, nil},
	}

	seedErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`DROP INDEX IF EXISTS home_file_active_starred_by_starred_at`); err != nil {
			return err
		}
		if _, err := tx.Exec(`ALTER TABLE home_file DROP COLUMN IF EXISTS starred_at`); err != nil {
			return err
		}
		for _, seed := range seeds {
			if _, err := tx.Exec(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, last_interaction, type, checksum, starred)
				VALUES ($1, $2, '/lib', '.txt', 1, $3, $3, $4, 2, '', $5)`,
				seed.name, "/lib/"+seed.name, updatedAt, seed.lastInteraction, seed.isStarred); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed: %v", seedErr)
	}

	for attempt := 0; attempt < 2; attempt++ {
		runErr := dbContext.ExecTx(func(tx *sql.Tx) error {
			_, err := tx.Exec(migrations.AddHomeFileStarredAtQuery)
			return err
		})
		if runErr != nil {
			t.Fatalf("run migration attempt %d: %v", attempt, runErr)
		}
	}

	for _, seed := range seeds {
		var starredAt sql.NullTime
		queryErr := dbContext.QueryTx(func(tx *sql.Tx) error {
			return tx.QueryRow(`SELECT starred_at FROM home_file WHERE name = $1`, seed.name).Scan(&starredAt)
		})
		if queryErr != nil {
			t.Fatalf("read starred_at of %s: %v", seed.name, queryErr)
		}
		expectedTime, hasExpectedTime := seed.expected.(time.Time)
		if hasExpectedTime != starredAt.Valid || (hasExpectedTime && !starredAt.Time.Equal(expectedTime)) {
			t.Fatalf("starred_at of %s = %v, want %v", seed.name, starredAt, seed.expected)
		}
	}
}
