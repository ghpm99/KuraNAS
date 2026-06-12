package trash

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/trash"
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

func scanTrashItem(scanner interface{ Scan(...any) error }) (TrashItemModel, error) {
	var model TrashItemModel
	err := scanner.Scan(
		&model.ID,
		&model.OriginalPath,
		&model.TrashPath,
		&model.Size,
		&model.DeletedAt,
	)
	return model, err
}

func (r *Repository) CreateItem(tx *sql.Tx, item TrashItemModel) (TrashItemModel, error) {
	created, err := scanTrashItem(tx.QueryRow(
		queries.InsertTrashItemQuery,
		item.OriginalPath,
		item.TrashPath,
		item.Size,
		item.DeletedAt,
	))
	if err != nil {
		return TrashItemModel{}, fmt.Errorf("CreateItem: %w", err)
	}
	return created, nil
}

func (r *Repository) queryItems(query string, args ...any) ([]TrashItemModel, error) {
	models := make([]TrashItemModel, 0)

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			model, scanErr := scanTrashItem(rows)
			if scanErr != nil {
				return scanErr
			}
			models = append(models, model)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (r *Repository) GetItems(page int, pageSize int) (utils.PaginationResponse[TrashItemModel], error) {
	paginationResponse := utils.PaginationResponse[TrashItemModel]{
		Items: []TrashItemModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
		},
	}

	items, err := r.queryItems(queries.GetTrashItemsQuery, pageSize+1, utils.CalculateOffset(page, pageSize))
	if err != nil {
		return paginationResponse, fmt.Errorf("GetItems: %w", err)
	}

	paginationResponse.Items = items
	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetItemByID(id int) (TrashItemModel, bool, error) {
	var model TrashItemModel
	found := false

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		scanned, scanErr := scanTrashItem(tx.QueryRow(queries.GetTrashItemByIDQuery, id))
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		model = scanned
		found = true
		return nil
	})
	if err != nil {
		return TrashItemModel{}, false, fmt.Errorf("GetItemByID: %w", err)
	}

	return model, found, nil
}

func (r *Repository) GetExpiredItems(cutoff time.Time) ([]TrashItemModel, error) {
	items, err := r.queryItems(queries.GetExpiredTrashItemsQuery, cutoff)
	if err != nil {
		return nil, fmt.Errorf("GetExpiredItems: %w", err)
	}
	return items, nil
}

func (r *Repository) GetAllItems() ([]TrashItemModel, error) {
	items, err := r.queryItems(queries.GetAllTrashItemsQuery)
	if err != nil {
		return nil, fmt.Errorf("GetAllItems: %w", err)
	}
	return items, nil
}

func (r *Repository) DeleteItem(tx *sql.Tx, id int) error {
	result, err := tx.Exec(queries.DeleteTrashItemQuery, id)
	if err != nil {
		return fmt.Errorf("DeleteItem: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteItem rows affected: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) GetRetentionDays() (int, bool, error) {
	days := 0
	found := false

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		var value string
		scanErr := tx.QueryRow(queries.GetRetentionDaysQuery).Scan(&value)
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return nil // an unreadable stored value behaves as unset
		}
		days = parsed
		found = true
		return nil
	})
	if err != nil {
		return 0, false, fmt.Errorf("GetRetentionDays: %w", err)
	}

	return days, found, nil
}

func (r *Repository) SetRetentionDays(days int) error {
	err := r.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(queries.UpsertRetentionDaysQuery, strconv.Itoa(days))
		return execErr
	})
	if err != nil {
		return fmt.Errorf("SetRetentionDays: %w", err)
	}
	return nil
}
