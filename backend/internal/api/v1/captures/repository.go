package captures

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/captures"
	"nas-go/api/pkg/utils"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(db *database.DbContext) *Repository {
	return &Repository{DbContext: db}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
}

func (r *Repository) CreateCapture(transaction *sql.Tx, capture CaptureModel) (CaptureModel, error) {
	args := []any{
		capture.Name,
		capture.FileName,
		capture.FilePath,
		capture.MediaType,
		capture.MimeType,
		capture.Size,
		capture.CreatedAt,
	}

	var id int
	err := transaction.QueryRow(queries.InsertCaptureQuery, args...).Scan(&id)
	if err != nil {
		return capture, fmt.Errorf("CreateCapture: %w", err)
	}

	capture.ID = id
	return capture, nil
}

func (r *Repository) GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureModel], error) {
	paginationResponse := utils.PaginationResponse[CaptureModel]{
		Items: []CaptureModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		!filter.Name.HasValue,
		filter.Name.Value,
		!filter.MediaType.HasValue,
		filter.MediaType.Value,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetCapturesQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m CaptureModel
			if err := rows.Scan(
				&m.ID,
				&m.Name,
				&m.FileName,
				&m.FilePath,
				&m.MediaType,
				&m.MimeType,
				&m.Size,
				&m.CreatedAt,
			); err != nil {
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, m)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("GetCaptures: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetCaptureByID(id int) (CaptureModel, error) {
	var m CaptureModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetCaptureByIDQuery, id).Scan(
			&m.ID,
			&m.Name,
			&m.FileName,
			&m.FilePath,
			&m.MediaType,
			&m.MimeType,
			&m.Size,
			&m.CreatedAt,
		)
	})

	if err != nil {
		return m, fmt.Errorf("GetCaptureByID: %w", err)
	}

	return m, nil
}

func (r *Repository) DeleteCapture(transaction *sql.Tx, id int) error {
	result, err := transaction.Exec(queries.DeleteCaptureQuery, id)
	if err != nil {
		return fmt.Errorf("DeleteCapture: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteCapture: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("DeleteCapture: capture not found")
	}

	return nil
}
