package database

import (
	"database/sql"
	"fmt"
	"log"
	"nas-go/api/pkg/database/queries"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var dbName = "db.sqlite3"

func ConfigDatabase() (*sql.DB, error) {

	dbPath := os.Getenv("DB_PATH")

	if dbPath == "" {
		dbPath = "./" + dbName
	}

	localDatabase, errSql := sql.Open("sqlite3", dbPath)

	if errSql != nil {
		fmt.Println("Erro ao conectar ao banco de dados SQLite:", errSql)
		return nil, errSql
	}

	fmt.Println("Successfully connected to database!")

	createTable(localDatabase)
	return localDatabase, nil

}

func createTable(db *sql.DB) {
	_, err := db.Exec(queries.CreateTableQuery)

	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}
}
