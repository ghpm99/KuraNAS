package database

import (
	"database/sql"
	"fmt"
	"log"
	"nas-go/api/internal/config"
	diaryQueries "nas-go/api/pkg/database/queries/diary"
	fileQueries "nas-go/api/pkg/database/queries/file"

	_ "github.com/mattn/go-sqlite3"
)

func ConfigDatabase() (*sql.DB, error) {

	dbPath := config.GetBuildConfig("DbPath")

	fmt.Println("Database path", dbPath)
	localDatabase, errSql := sql.Open("sqlite3", dbPath)

	if errSql != nil {
		fmt.Println("Erro ao conectar ao banco de dados SQLite:", errSql)
		return nil, errSql
	}

	fmt.Println("Successfully connected to database!")

	createFileTable(localDatabase)
	createDiaryTable(localDatabase)
	return localDatabase, nil

}

func createDiaryTable(db *sql.DB) {
	_, err := db.Exec(diaryQueries.CreateTableQuery)

	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}

}

func createFileTable(db *sql.DB) {

	_, err := db.Exec(fileQueries.CreateTableQuery)

	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}
}
