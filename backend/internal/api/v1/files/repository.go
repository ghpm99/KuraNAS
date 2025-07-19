package files

import (
	"database/sql"
	"errors"
	"fmt"

	queries "nas-go/api/pkg/database/queries/file"
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

	args := []any{
		!filter.ID.HasValue,
		filter.ID.Value,
		!filter.Name.HasValue,
		filter.Name.Value,
		!filter.Path.HasValue,
		filter.Path.Value,
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

	rows, err := r.DbContext.Query(
		queries.GetFilesQuery,
		args...,
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
	fail := func(err error) (int, error) {
		return 0, fmt.Errorf("GetDirectoryContentCount: %v", err)
	}
	row := r.DbContext.QueryRow(
		queries.GetChildrenCountQuery,
		parentPath,
		fileId,
	)
	var childrenCount int

	if err := row.Scan(&childrenCount); err != nil {
		return fail(err)
	}

	return childrenCount, nil

}

func (r *Repository) GetCountByType(fileType FileType) (int, error) {
	fail := func(err error) (int, error) {
		return 0, fmt.Errorf("GetCountByType: %v", err)
	}

	row := r.DbContext.QueryRow(
		queries.CountByTypeQuery,
		fileType,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return fail(err)
	}

	return count, nil
}

func (r *Repository) GetTotalSpaceUsed() (int, error) {
	fail := func(err error) (int, error) {
		return 0, fmt.Errorf("GetTotalSpaceUsed: %v", err)
	}

	row := r.DbContext.QueryRow(queries.TotalSpaceUsedQuery)

	var totalSpaceUsed int
	if err := row.Scan(&totalSpaceUsed); err != nil {
		return fail(err)
	}

	return totalSpaceUsed, nil
}

func (r *Repository) GetReportSizeByFormat() ([]SizeReportModel, error) {
	fail := func(err error) ([]SizeReportModel, error) {
		return nil, fmt.Errorf("GetReportSizeByFormat: %v", err)
	}

	rows, err := r.DbContext.Query(queries.CountByFormatQuery, File)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()

	var report []SizeReportModel

	for rows.Next() {
		var item SizeReportModel
		if err := rows.Scan(&item.Format, &item.Total, &item.Size); err != nil {
			return fail(err)
		}
		report = append(report, item)
	}

	return report, nil
}

func (r *Repository) GetTopFilesBySize(limit int) ([]FileModel, error) {
	fail := func(err error) ([]FileModel, error) {
		return nil, fmt.Errorf("GetTopFilesBySize: %v", err)
	}

	rows, err := r.DbContext.Query(queries.TopFilesBySizeQuery, limit)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()

	var topFiles []FileModel

	for rows.Next() {
		var file FileModel
		if err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.Size,
			&file.Path,
		); err != nil {
			return fail(err)
		}
		topFiles = append(topFiles, file)
	}

	return topFiles, nil
}

func (r *Repository) GetDuplicateFiles(page int, pageSize int) (utils.PaginationResponse[DuplicateFilesModel], error) {

	paginationResponse := utils.PaginationResponse[DuplicateFilesModel]{
		Items: []DuplicateFilesModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	fail := func(err error) (utils.PaginationResponse[DuplicateFilesModel], error) {
		return paginationResponse, fmt.Errorf("GetDuplicateFiles: %v", err)
	}

	rows, err := r.DbContext.Query(
		queries.GetDuplicateFilesQuery,
		pageSize+1,
		utils.CalculateOffset(page, pageSize),
	)
	if err != nil {
		return fail(err)
	}
	defer rows.Close()

	for rows.Next() {
		var duplicate DuplicateFilesModel
		if err := rows.Scan(
			&duplicate.Name,
			&duplicate.Size,
			&duplicate.Copies,
			&duplicate.Paths,
		); err != nil {
			return fail(err)
		}
		paginationResponse.Items = append(paginationResponse.Items, duplicate)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}
