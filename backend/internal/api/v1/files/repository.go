package files

import (
	"database/sql"
	"errors"
	"fmt"
	"nas-go/api/pkg/database/queries"
	"nas-go/api/pkg/utils"
)

type Repository struct {
	DbContext *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{database}
}

func (r *Repository) GetDbContext() *sql.DB {
	return r.DbContext
}

func (r *Repository) GetFiles(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {

	paginationResponse := utils.PaginationResponse[FileModel]{
		Items: []FileModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	fmt.Println("GetFiles: ", filter, pageSize, page)

	rows, err := r.DbContext.Query(
		queries.GetFilesQuery,
		!filter.ID.HasValue,
		filter.ID.Value,
		!filter.Name.HasValue,
		filter.Name.Value,
		!filter.Path.HasValue,
		filter.Path.Value,
		!filter.Format.HasValue,
		filter.Format.Value,
		!filter.Type.HasValue,
		filter.Type.Value,
		!filter.DeletedAt.HasValue,
		filter.DeletedAt.Value,
		pageSize+1,
		page,
	)
	if err != nil {
		return paginationResponse, err
	}
	defer rows.Close()

	for rows.Next() {
		var file FileModel
		if err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.Path,
			&file.Format,
			&file.Size,
			&file.UpdatedAt,
			&file.CreatedAt,
			&file.LastInteraction,
			&file.LastBackup,
			&file.Type,
			&file.CheckSum,
			&file.DeletedAt,
		); err != nil {
			return paginationResponse, err
		}

		paginationResponse.Items = append(paginationResponse.Items, file)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}

func (r *Repository) CreateFile(transaction *sql.Tx, file FileModel) (FileModel, error) {

	fail := func(err error) (FileModel, error) {
		return file, fmt.Errorf("CreateFile: %v", err)
	}

	args := []any{
		file.Name,
		file.Path,
		file.Format,
		file.Size,
		file.UpdatedAt,
		file.CreatedAt,
		file.LastInteraction,
		file.LastBackup,
		file.DeletedAt,
		file.Type,
		file.CheckSum,
	}

	query := queries.InsertFileQuery

	data, err := transaction.Exec(
		query,
		args...,
	)

	if err != nil {
		return fail(err)
	}

	fileId, err := data.LastInsertId()

	if err != nil {
		return fail(err)
	}

	file.ID = int(fileId)

	return file, nil
}

func (r *Repository) UpdateFile(transaction *sql.Tx, file FileModel) (bool, error) {
	fail := func(err error) (bool, error) {
		return false, fmt.Errorf("UpdateFile: %v", err)
	}

	result, err := transaction.Exec(
		queries.UpdateFileQuery,
		&file.ID,
		&file.Name,
		&file.Path,
		&file.Format,
		&file.Size,
		&file.UpdatedAt,
		&file.CreatedAt,
		&file.LastInteraction,
		&file.LastBackup,
		&file.Type,
		&file.CheckSum,
		&file.DeletedAt,
	)

	if err != nil {
		return fail(err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fail(err)
	}

	if rowsAffected > 1 {
		transaction.Rollback()
		return fail(errors.New("multiple rows affected"))
	}

	return rowsAffected == 1, nil
}
