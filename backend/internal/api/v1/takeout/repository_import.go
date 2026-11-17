package takeout

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/takeout"
)

type ImportRepository struct {
	DbContext *database.DbContext
}

func NewImportRepository(dbContext *database.DbContext) *ImportRepository {
	return &ImportRepository{DbContext: dbContext}
}

func (r *ImportRepository) CreateQueuedImport(transaction *sql.Tx, jobID int, fileName string, size int64) (int, error) {
	var importID int
	err := transaction.QueryRow(queries.InsertTakeoutImportQuery, jobID, fileName, size).Scan(&importID)
	if err != nil {
		return 0, fmt.Errorf("CreateQueuedImport: %w", err)
	}
	return importID, nil
}

func (r *ImportRepository) MarkImportRunning(jobID int) error {
	return r.execForJob("MarkImportRunning", func(transaction *sql.Tx) error {
		_, err := transaction.Exec(queries.MarkTakeoutImportRunningQuery, jobID)
		return err
	})
}

func (r *ImportRepository) CloseImportCompleted(jobID int, summary ImportSummary) error {
	return r.execForJob("CloseImportCompleted", func(transaction *sql.Tx) error {
		_, err := transaction.Exec(
			queries.CloseTakeoutImportCompletedQuery,
			jobID,
			summary.TotalEntries,
			summary.Imported,
			summary.AlreadyExisting,
			summary.Ignored,
			summary.FailedEntries,
		)
		return err
	})
}

func (r *ImportRepository) CloseImportFailed(jobID int, failureMessage string) error {
	return r.execForJob("CloseImportFailed", func(transaction *sql.Tx) error {
		_, err := transaction.Exec(queries.CloseTakeoutImportFailedQuery, jobID, failureMessage)
		return err
	})
}

func (r *ImportRepository) execForJob(operation string, execute func(transaction *sql.Tx) error) error {
	if err := r.DbContext.ExecTx(execute); err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	return nil
}
