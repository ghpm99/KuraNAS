package diary

import "database/sql"

type Repository struct {
	DbContext *sql.DB
}
