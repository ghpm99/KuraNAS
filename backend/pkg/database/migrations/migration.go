package migrations

import (
	"database/sql"
)

type migration struct {
	Name    string
	Migrate func(*sql.Tx) error
}

var migrationList = []migration{
	{
		Name:    "create_migrations_table",
		Migrate: createMigrationDatabase,
	},
}

func Init(db *sql.DB) {
	if db == nil {
		panic("Database connection is nil")
	}

	tx, err := db.BeginTx(nil, nil)
	if err != nil {
		panic("Failed to begin transaction: " + err.Error())
	}
	defer tx.Rollback()

	createMigrationDatabase(tx)

	for _, m := range migrationList {
		if err := runMigration(tx, m.Name, m.Migrate); err != nil {
			panic("Failed to run migration " + m.Name + ": " + err.Error())
		}
	}

}

func createMigrationDatabase(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func recordMigration(tx *sql.Tx, name string) error {
	_, err := tx.Exec(`
		INSERT INTO migrations (name) VALUES (?);
	`, name)
	return err
}

func migrationExists(tx *sql.Tx, name string) (bool, error) {
	rows, err := tx.Query(`
		SELECT COUNT(*) FROM migrations WHERE name = ?;
	`, name)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			return false, err
		}
		return count > 0, nil
	}
	return false, nil
}

func runMigration(tx *sql.Tx, name string, migrationFunc func(*sql.Tx) error) error {
	exists, err := migrationExists(tx, name)
	if err != nil {
		return err
	}
	if exists {
		return nil // Migration already applied
	}

	if err := migrationFunc(tx); err != nil {
		return err
	}

	return recordMigration(tx, name)
}

func addMigration(name string, migrationFunc func(*sql.Tx) error) error {
	migrationList = append(migrationList,
		migration{
			Name:    name,
			Migrate: migrationFunc,
		})
	return nil
}
