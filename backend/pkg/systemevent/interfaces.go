package systemevent

import (
	"database/sql"
	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	Insert(tx *sql.Tx, event EventModel) error
}

type ServiceInterface interface {
	RecordStartup() error
	RecordShutdown() error
}
