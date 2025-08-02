package migrations

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
)

type migration struct {
	Name    string
	Migrate func(*sql.Tx) error
}

//go:embed queries/create_migrations_table.sql
var createMigrationDatabaseQuery string

//go:embed queries/insert_migration.sql
var insertMigrationQuery string

//go:embed queries/migration_exists.sql
var migrationExistsQuery string

var migrationList = []migration{
	{
		Name:    "create_migrations_table",
		Migrate: createMigrationDatabase,
	},
}

func Init(db *sql.DB) {
	if db == nil {
		log.Println("Database connection is nil")
		panic("Database connection is nil")
	}
	initMigrationList()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		panic("Failed to begin transaction: " + err.Error())
	}
	defer tx.Rollback()

	createMigrationDatabase(tx)

	for _, m := range migrationList {
		if err := runMigration(tx, m.Name, m.Migrate); err != nil {
			log.Printf("Failed to run migration %s: %v", m.Name, err)
			panic("Failed to run migration " + m.Name + ": " + err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		panic("Failed to commit transaction: " + err.Error())
	}
	log.Println("All migrations applied successfully")

}

func initMigrationList() {
	logMigrationList()
	diaryMigrationList()
	fileMigrationList()
}

func createMigrationDatabase(tx *sql.Tx) error {
	_, err := tx.Exec(createMigrationDatabaseQuery)
	return err
}

func recordMigration(tx *sql.Tx, name string) error {
	_, err := tx.Exec(insertMigrationQuery, name)
	return err
}

func migrationExists(tx *sql.Tx, name string) (bool, error) {
	rows, err := tx.Query(migrationExistsQuery, name)
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
		return nil
	}

	if err := migrationFunc(tx); err != nil {
		return err
	}

	return recordMigration(tx, name)
}

func addMigration(name string, migrationFunc func(*sql.Tx) error) {
	migrationList = append(migrationList,
		migration{
			Name:    name,
			Migrate: migrationFunc,
		})

}
