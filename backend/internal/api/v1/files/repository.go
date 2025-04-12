package files

import (
	"database/sql"
	"errors"
	"fmt"
	"nas-go/api/pkg/database/queries"
	"nas-go/api/pkg/utils"
)

type Repository struct {
	dbContext *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{database}
}

func (r *Repository) GetFiles(filter FileFilter, pagination utils.Pagination) (utils.PaginationResponse[FileModel], error) {

	paginationResponse := utils.PaginationResponse[FileModel]{
		Items:      []FileModel{},
		Pagination: pagination,
	}

	fmt.Println("GetFiles: ", filter.Path, pagination.PageSize, pagination.Page)

	rows, err := r.dbContext.Query(
		queries.GetFilesQuery,
		filter.Path,
		pagination.PageSize+1,
		pagination.Page,
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

func (r *Repository) GetFilesByPath(path string) ([]FileModel, error) {
	rows, err := r.dbContext.Query(
		queries.GetFilesByPathQuery,
		path,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []FileModel
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
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func (r *Repository) GetFileByNameAndPath(name string, path string) (FileModel, error) {
	row := r.dbContext.QueryRow(
		queries.GetFileByNameAndPathQuery,
		name,
		path,
	)

	var file FileModel

	if err := row.Scan(
		&file.ID,
		&file.Name,
		&file.Path,
		&file.Format,
		&file.Size,
		&file.UpdatedAt,
		&file.CreatedAt,
		&file.LastInteraction,
		&file.LastBackup,
	); err != nil {
		return file, err
	}

	return file, nil
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

func (r *Repository) GetPathByFileId(fileId int) (string, error) {
	row := r.dbContext.QueryRow(
		queries.GetPathByFileIdQuery,
		fileId,
	)

	var path string

	if err := row.Scan(
		&path,
	); err != nil {
		return "", err
	}

	return path, nil
}
