package systemevent

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/systemevent"
)

type Repository struct {
	dbContext *database.DbContext
}

func NewRepository(dbContext *database.DbContext) *Repository {
	return &Repository{dbContext: dbContext}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.dbContext
}

func (r *Repository) Insert(tx *sql.Tx, event EventModel) error {
	if tx == nil {
		return fmt.Errorf("insert system event: tx is nil")
	}

	_, err := tx.Exec(
		queries.InsertSystemEventQuery,
		event.EventTime,
		event.EventTimeDisplay,
		event.EventType,
		event.Description,
		event.Source,
		event.HostName,
		event.ProcessID,
		event.ExtraData,
	)
	if err != nil {
		return fmt.Errorf("insert system event: %w", err)
	}

	return nil
}
