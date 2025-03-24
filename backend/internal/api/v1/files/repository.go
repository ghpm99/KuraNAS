package files

import (
	"database/sql"
	"nas-go/api/pkg/database/queries"
	"nas-go/api/pkg/utils"
)

type Repository struct {
	dbContext *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{database}
}

func (r *Repository) GetFiles(pagination utils.Pagination) (utils.PaginationResponse[FileModel], error) {

	paginationResponse := utils.PaginationResponse[FileModel]{
		Items:      nil,
		Pagination: pagination,
	}

	rows, err := r.dbContext.Query(
		queries.GetFilesQuery,
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
		); err != nil {
			return paginationResponse, err
		}
		paginationResponse.Items = append(paginationResponse.Items, file)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}
