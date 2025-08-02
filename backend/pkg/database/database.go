package database

import (
	"database/sql"
	"log"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database/migrations"

	_ "github.com/mattn/go-sqlite3"
)

func ConfigDatabase() (*sql.DB, error) {

	dbPath := config.GetBuildConfig("DbPath")

	log.Println("Database path", dbPath)
	localDatabase, errSql := sql.Open("sqlite3", dbPath)

	if errSql != nil {
		log.Println("Erro ao conectar ao banco de dados SQLite:", errSql)
		return nil, errSql
	}

	log.Println("Successfully connected to database!")
	migrations.Init(localDatabase)
	return localDatabase, nil

}
