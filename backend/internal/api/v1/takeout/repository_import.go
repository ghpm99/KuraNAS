package takeout

import (
	"database/sql"
	"errors"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/takeout"
	"nas-go/api/pkg/utils"
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

func (r *ImportRepository) UpdateImportProgress(jobID int, processedEntries int, totalEntries int) error {
	return r.execForJob("UpdateImportProgress", func(transaction *sql.Tx) error {
		_, err := transaction.Exec(queries.UpdateTakeoutImportProgressQuery, jobID, totalEntries, processedEntries)
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

func (r *ImportRepository) ListImports(page int, pageSize int) (utils.PaginationResponse[TakeoutImportDto], error) {
	pagination := utils.PaginationResponse[TakeoutImportDto]{
		Items:      []TakeoutImportDto{},
		Pagination: utils.Pagination{Page: page, PageSize: pageSize},
	}

	err := r.DbContext.QueryTx(func(transaction *sql.Tx) error {
		rows, err := transaction.Query(queries.SelectTakeoutImportsPageQuery, pageSize+1, utils.CalculateOffset(page, pageSize))
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			importDto, err := scanImport(rows)
			if err != nil {
				return err
			}
			pagination.Items = append(pagination.Items, importDto)
		}
		return rows.Err()
	})
	if err != nil {
		return pagination, fmt.Errorf("ListImports: %w", err)
	}

	pagination.UpdatePagination()
	return pagination, nil
}

type importRowScanner interface {
	Scan(destinations ...any) error
}

func scanImport(row importRowScanner) (TakeoutImportDto, error) {
	var importDto TakeoutImportDto
	err := row.Scan(
		&importDto.ID,
		&importDto.JobID,
		&importDto.FileName,
		&importDto.Size,
		&importDto.State,
		&importDto.TotalEntries,
		&importDto.ProcessedEntries,
		&importDto.Imported,
		&importDto.AlreadyExisting,
		&importDto.Ignored,
		&importDto.FailedEntries,
		&importDto.Error,
		&importDto.CreatedAt,
		&importDto.EndedAt,
	)
	return importDto, err
}

func (r *ImportRepository) GetImport(importID int) (TakeoutImportDto, error) {
	var importDto TakeoutImportDto
	err := r.DbContext.QueryTx(func(transaction *sql.Tx) error {
		scanned, err := scanImport(transaction.QueryRow(queries.SelectTakeoutImportByIDQuery, importID))
		importDto = scanned
		return err
	})
	if errors.Is(err, sql.ErrNoRows) {
		return TakeoutImportDto{}, ErrImportNotFound
	}
	if err != nil {
		return TakeoutImportDto{}, fmt.Errorf("GetImport: %w", err)
	}
	return importDto, nil
}
