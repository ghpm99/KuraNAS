package files

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/utils"
)

func (r *Repository) GetTopFilesBySizePage(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	topFiles := []FileModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.TopFilesBySizePageQuery, pageSize+1, utils.CalculateOffset(page, pageSize))
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
				&file.Type,
				&file.Size,
				&file.UpdatedAt,
				&file.CreatedAt,
				&file.PhysicalPath,
			); err != nil {
				return err
			}
			topFiles = append(topFiles, file)
		}
		return rows.Err()
	})
	if err != nil {
		return utils.PaginationResponse[FileModel]{}, fmt.Errorf("failed to get top files by size page: %w", err)
	}

	hasNext := len(topFiles) > pageSize
	if hasNext {
		topFiles = topFiles[:pageSize]
	}
	return utils.PaginationResponse[FileModel]{
		Items: topFiles,
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  hasNext,
			HasPrev:  page > 1,
		},
	}, nil
}
