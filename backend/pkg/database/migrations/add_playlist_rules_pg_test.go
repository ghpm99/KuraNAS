package migrations_test

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database/migrations"
)

func TestAddPlaylistRulesBackfillsKindFromFlags_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO playlist (name, description, is_system, is_ai_generated) VALUES
			('manual', '', FALSE, FALSE), ('system', '', TRUE, FALSE), ('ai', '', FALSE, TRUE)`); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO playlist (name, description, is_system, kind, rules) VALUES ('rule', '', FALSE, 'rule', '{"match":"any"}')`); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE playlist SET kind = 'manual' WHERE name <> 'rule'`); err != nil {
			return err
		}
		if _, err := tx.Exec(migrations.AddPlaylistRulesQuery); err != nil {
			return err
		}

		kindByName := map[string]string{}
		rows, err := tx.Query(`SELECT name, kind FROM playlist`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var name, kind string
			if err := rows.Scan(&name, &kind); err != nil {
				return err
			}
			kindByName[name] = kind
		}
		for name, expectedKind := range map[string]string{"manual": "manual", "system": "system", "ai": "ai", "rule": "rule"} {
			if kindByName[name] != expectedKind {
				t.Fatalf("playlist %s: expected kind %s, got %s", name, expectedKind, kindByName[name])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("backfill: %v", err)
	}
}

func TestPlaylistRulesSchemaConstraints_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_migrations_it")

	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`TRUNCATE playlist, home_file RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		var playlistID, fileID int
		if err := tx.QueryRow(`INSERT INTO playlist (name, description, is_system) VALUES ('p', '', FALSE) RETURNING id`).Scan(&playlistID); err != nil {
			return err
		}
		var defaultKind string
		var rulesAreNull bool
		if err := tx.QueryRow(`SELECT kind, rules IS NULL FROM playlist WHERE id = $1`, playlistID).Scan(&defaultKind, &rulesAreNull); err != nil {
			return err
		}
		if defaultKind != "manual" || !rulesAreNull {
			t.Fatalf("defaults must be manual and NULL rules, got %s %v", defaultKind, rulesAreNull)
		}
		if err := tx.QueryRow(`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, type, checksum)
			VALUES ('a.mp3', '/a.mp3', '/', '.mp3', 1, now(), now(), 2, '') RETURNING id`).Scan(&fileID); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO playlist_rule_exclusion (playlist_id, file_id) VALUES ($1, $2)`, playlistID, fileID); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM playlist WHERE id = $1`, playlistID); err != nil {
			return err
		}
		var remainingExclusions int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM playlist_rule_exclusion`).Scan(&remainingExclusions); err != nil {
			return err
		}
		if remainingExclusions != 0 {
			t.Fatalf("exclusions must cascade with the playlist, %d left", remainingExclusions)
		}

		_, invalidKindErr := tx.Exec(`INSERT INTO playlist (name, description, is_system, kind) VALUES ('bad', '', FALSE, 'bogus')`)
		if invalidKindErr == nil {
			t.Fatal("kind outside manual/rule/system/ai must violate the CHECK constraint")
		}
		return invalidKindErr
	})
	if err == nil {
		t.Fatal("expected the invalid kind insert to abort the transaction")
	}
}
