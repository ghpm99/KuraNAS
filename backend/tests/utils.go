package tests

import (
	"database/sql"
	"log"

	"github.com/DATA-DOG/go-sqlmock"
)

func ConfigInMemoryDatabase() *sql.DB {
	db, _, err := sqlmock.New()
	if err != nil {
		log.Fatalf("Falha ao criar banco de dados em memoria: %v", err)
	}

	return db
}
