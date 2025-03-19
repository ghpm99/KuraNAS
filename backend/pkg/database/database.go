package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func ConfigDatabase() (*sql.DB, error) {

	dbPath := os.Getenv("DB_PATH")

	if dbPath == "" {
		dbPath = "./database.db"
	}

	localDatabase, errSql := sql.Open("sqlite3", dbPath)

	if errSql != nil {
		fmt.Println("Erro ao conectar ao banco de dados SQLite:", errSql)
		return nil, errSql
	}

	fmt.Println("Successfully connected to database!")
	return localDatabase, nil

}
