package migrations

import "database/sql"

func fileMigrationList() {
	migrationList = append(migrationList,
		migration{
			Name: "20250617_create_home_file_table",
			Migrate: func(tx *sql.Tx) error {
				_, err := tx.Exec(`
				CREATE TABLE
    IF NOT EXISTS "home_file" (
        "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
        "name" varchar(256) NOT NULL,
        "path" varchar(1024) NOT NULL,
        "parent_path" varchar(1024) NOT NULL,
        "format" varchar(256) NOT NULL,
        "size" integer NOT NULL,
        "updated_at" datetime NOT NULL,
        "created_at" datetime NOT NULL,
        "last_interaction" datetime NULL,
        "last_backup" datetime NULL,
        "type" INTEGER,
        "checksum" VARCHAR(64),
        "deleted_at" DATETIME NULL
    );
			`)
				return err
			}},
		migration{
			Name: "20250617_add_file_starred_column",
			Migrate: func(tx *sql.Tx) error {
				_, err := tx.Exec(`
				ALTER TABLE files ADD COLUMN starred BOOLEAN DEFAULT FALSE;
			`)
				return err
			},
		})
}
