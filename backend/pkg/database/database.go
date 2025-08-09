package database

import (
	"database/sql"
	"log"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database/migrations"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func ConfigDatabase() (*sql.DB, error) {

	dbPath := config.GetBuildConfig("DbPath")

	dbPathWithDSN := applyDatabaseConfigDSN(dbPath)

	log.Println("Database path", dbPathWithDSN)
	localDatabase, errSql := sql.Open("sqlite3", dbPathWithDSN)

	if errSql != nil {
		log.Println("Erro ao conectar ao banco de dados SQLite:", errSql)
		return nil, errSql
	}

	log.Println("Successfully connected to database!")
	migrations.Init(localDatabase)
	return localDatabase, nil

}

func applyDatabaseConfigDSN(dbPath string) string {
	dsnQuery := "_busy_timeout=" + strconv.Itoa(config.AppConfig.DbBuzyTimeout)

	dsnQuery += "&_journal_mode=" + config.AppConfig.DbJournalMode

	return dbPath + "?" + dsnQuery
}
