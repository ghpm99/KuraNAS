package database

import (
	"database/sql"
	"fmt"
	"log"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database/migrations"

	_ "github.com/lib/pq"
)

func ConfigDatabase() (*sql.DB, error) {

	localDatabase, errSql := sql.Open("postgres", applyDatabaseConfig())

	if errSql != nil {
		log.Println("Erro ao conectar ao banco de dados:", errSql)
		return nil, errSql
	}

	log.Println("Successfully connected to database!")
	migrations.Init(localDatabase)
	return localDatabase, nil

}

func applyDatabaseConfig() string {
	psqlSetup := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		config.AppConfig.DbHost,
		config.AppConfig.DbPort,
		config.AppConfig.DbUser,
		config.AppConfig.DbName,
		config.AppConfig.DbPassword,
	)

	return psqlSetup
}
