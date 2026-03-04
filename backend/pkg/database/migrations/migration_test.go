package migrations

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func openMigrationDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestInitPanicsWhenDBIsNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when database is nil")
		}
	}()
	Init(nil)
}

func TestCreateMigrationDatabaseAndRecord(t *testing.T) {
	db := openMigrationDB(t)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := createMigrationDatabase(tx); err != nil {
		t.Fatalf("createMigrationDatabase returned error: %v", err)
	}

	var tableCount int
	if err := tx.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='migrations'`).Scan(&tableCount); err != nil {
		t.Fatalf("failed to query migrations table: %v", err)
	}
	if tableCount != 1 {
		t.Fatalf("expected migrations table to exist, got %d", tableCount)
	}

	if err := recordMigration(tx, "m001"); err != nil {
		t.Fatalf("recordMigration returned error: %v", err)
	}

	var applied int
	if err := tx.QueryRow(`SELECT count(*) FROM migrations`).Scan(&applied); err != nil {
		t.Fatalf("failed to query applied migrations: %v", err)
	}
	if applied != 1 {
		t.Fatalf("expected 1 applied migration, got %d", applied)
	}
}

func TestMigrationExistsAndRunMigration(t *testing.T) {
	db := openMigrationDB(t)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback()

	if err := createMigrationDatabase(tx); err != nil {
		t.Fatalf("failed to create migration table: %v", err)
	}

	name := "test_migration"
	calls := 0
	run := func(tx *sql.Tx) error {
		calls++
		_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS migration_target (id INTEGER)`)
		return err
	}

	if err := runMigration(tx, name, run); err != nil {
		t.Fatalf("runMigration returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected migration func to execute once, got %d", calls)
	}

	exists, err := migrationExists(tx, name)
	if err != nil {
		t.Fatalf("migrationExists returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected migration to exist")
	}

	if err := runMigration(tx, name, run); err != nil {
		t.Fatalf("runMigration second call returned error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected migration func to be skipped on second call, got %d", calls)
	}
}

func TestDefaultMigrationFuncAndAddMigration(t *testing.T) {
	db := openMigrationDB(t)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback()

	fn := defaultMigrationFunc(`CREATE TABLE sample_table (id INTEGER)`)
	if err := fn(tx); err != nil {
		t.Fatalf("default migration func returned error: %v", err)
	}

	oldList := migrationList
	migrationList = nil
	t.Cleanup(func() {
		migrationList = oldList
	})

	addMigration("one", fn)
	if len(migrationList) != 1 {
		t.Fatalf("expected 1 migration in list, got %d", len(migrationList))
	}
}
