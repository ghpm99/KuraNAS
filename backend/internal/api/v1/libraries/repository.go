package libraries

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/libraries"
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

func (r *Repository) GetAll() ([]LibraryModel, error) {
	var libraries []LibraryModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetLibrariesQuery)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m LibraryModel
			if err := rows.Scan(&m.ID, &m.Category, &m.Path, &m.CreatedAt, &m.UpdatedAt); err != nil {
				return err
			}
			libraries = append(libraries, m)
		}
		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("GetAll: %w", err)
	}

	return libraries, nil
}

func (r *Repository) GetByCategory(category LibraryCategory) (LibraryModel, error) {
	var m LibraryModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetLibraryByCategoryQuery, string(category)).Scan(
			&m.ID, &m.Category, &m.Path, &m.CreatedAt, &m.UpdatedAt,
		)
	})

	if err != nil {
		return m, fmt.Errorf("GetByCategory: %w", err)
	}

	return m, nil
}

func (r *Repository) Upsert(tx *sql.Tx, model LibraryModel) (LibraryModel, error) {
	var m LibraryModel

	err := tx.QueryRow(
		queries.UpsertLibraryQuery,
		string(model.Category),
		model.Path,
	).Scan(&m.ID, &m.Category, &m.Path, &m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		return m, fmt.Errorf("Upsert: %w", err)
	}

	return m, nil
}
