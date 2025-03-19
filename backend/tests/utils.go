package tests

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func ConfigInMemoryDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Falha ao criar banco de dados em memória: %v", err)
	}

	return db
}
