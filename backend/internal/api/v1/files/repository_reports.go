package files

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/file"
	"nas-go/api/pkg/utils"
)

func (r *Repository) GetTotalSpaceUsed() (int, error) {
	var totalSpaceUsed int

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.TotalSpaceUsedQuery)

		if err := row.Scan(&totalSpaceUsed); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to get total space used: %w", err)
	}

	return totalSpaceUsed, nil
}

func (r *Repository) GetReportSizeByFormat() ([]SizeReportModel, error) {
	var report []SizeReportModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.CountByFormatQuery, File)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item SizeReportModel
			if err := rows.Scan(&item.Format, &item.Total, &item.Size); err != nil {
				return err
			}
			report = append(report, item)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get report by format: %w", err)
	}

	return report, nil
}

func (r *Repository) GetTopFilesBySize(limit int) ([]FileModel, error) {
	var topFiles []FileModel

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.TopFilesBySizeQuery, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file FileModel
			if err := rows.Scan(
				&file.ID,
				&file.Name,
				&file.Size,
				&file.Path,
			); err != nil {
				return err
			}
			topFiles = append(topFiles, file)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get top files by size: %w", err)
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

	args := []any{
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.GetDuplicateFilesQuery,
			args...,
		)
		if err != nil {
			return err
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
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, duplicate)
		}

		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to get duplicate files: %w", err)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}
