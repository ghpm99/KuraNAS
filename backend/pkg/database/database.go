package database

import (
	"database/sql"
	"fmt"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database/migrations"

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
	migrations.Init(localDatabase)
	return localDatabase, nil

}
