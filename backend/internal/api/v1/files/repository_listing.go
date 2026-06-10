package files

import (
	"database/sql"
	"errors"
	"fmt"

	queries "nas-go/api/pkg/database/queries/file"
	"nas-go/api/pkg/utils"
)

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

	args := []any{
		!filter.ID.HasValue,
		filter.ID.Value,
		!filter.Name.HasValue,
		filter.Name.Value,
		!filter.Path.HasValue,
		filter.Path.Value,
		!filter.PathPrefix.HasValue,
		filter.PathPrefix.Value,
		!filter.ParentPath.HasValue,
		filter.ParentPath.Value,
		!filter.Format.HasValue,
		filter.Format.Value,
		!filter.Type.HasValue,
		filter.Type.Value,
		!filter.DeletedAt.HasValue,
		filter.DeletedAt.Value,
		filter.Category,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.GetFilesQuery,
			args...,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file FileModel
			if err := rows.Scan(
				&file.ID,
				&file.Name,
				&file.Path,
				&file.ParentPath,
				&file.Format,
				&file.Size,
				&file.UpdatedAt,
				&file.CreatedAt,
				&file.LastInteraction,
				&file.LastBackup,
				&file.Type,
				&file.CheckSum,
				&file.DeletedAt,
				&file.Starred,
			); err != nil {
				return err
			}

			paginationResponse.Items = append(paginationResponse.Items, file)
		}

		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query files: %w", err)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}

// GetFileStatByPath returns the size and last-known modification time of the
// active (non-deleted) file at the given absolute path. The second return value
// reports whether such a row exists. It runs a single indexed lookup so the
// diff scan can compare one file at a time without batch-loading the table.
func (r *Repository) GetFileStatByPath(path string) (FileStat, bool, error) {
	var stat FileStat
	found := false

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		scanErr := tx.QueryRow(queries.GetFileStatByPathQuery, path).Scan(&stat.Size, &stat.UpdatedAt)
		if scanErr != nil {
			if errors.Is(scanErr, sql.ErrNoRows) {
				return nil
			}
			return scanErr
		}
		found = true
		return nil
	})
	if err != nil {
		return FileStat{}, false, fmt.Errorf("GetFileStatByPath: %w", err)
	}

	return stat, found, nil
}

func (r *Repository) CreateFile(transaction *sql.Tx, file FileModel) (FileModel, error) {

	fail := func(err error) (FileModel, error) {
		return file, fmt.Errorf("CreateFile: %v", err)
	}

	args := []any{
		file.Name,
		file.Path,
		file.ParentPath,
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

	var fileId int
	err := transaction.QueryRow(
		query,
		args...,
	).Scan(&fileId)

	if err != nil {
		return fail(err)
	}

	file.ID = fileId

	return file, nil
}

func (r *Repository) UpdateFile(transaction *sql.Tx, file FileModel) (bool, error) {
	fail := func(err error) (bool, error) {
		return false, fmt.Errorf("UpdateFile: %v", err)
	}

	result, err := transaction.Exec(
		queries.UpdateFileQuery,
		&file.Name,
		&file.Path,
		&file.ParentPath,
		&file.Format,
		&file.Size,
		&file.UpdatedAt,
		&file.CreatedAt,
		&file.LastInteraction,
		&file.LastBackup,
		&file.Type,
		&file.CheckSum,
		&file.DeletedAt,
		&file.Starred,
		&file.ID,
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

func (r *Repository) GetDirectoryContentCount(fileId int, parentPath string) (int, error) {
	var childrenCount int

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {

		row := tx.QueryRow(
			queries.GetChildrenCountQuery,
			parentPath,
			fileId,
		)

		if err := row.Scan(&childrenCount); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to get directory count: %w", err)
	}

	return childrenCount, nil
}

func (r *Repository) GetCountByType(fileType FileType) (int, error) {

	var count int
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {

		row := tx.QueryRow(
			queries.CountByTypeQuery,
			fileType,
		)

		if err := row.Scan(&count); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("GetCountByType: %v", err)
	}

	return count, nil
}
