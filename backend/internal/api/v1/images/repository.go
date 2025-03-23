package images

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

func (r *Repository) GetAllImages(pagination utils.Pagination) (utils.PaginationResponse[ImageModel], error) {
	paginationResponse := utils.PaginationResponse[ImageModel]{
		Items: nil,
		Pagination: utils.Pagination{
			Page:     pagination.Page,
			PageSize: pagination.PageSize + 1,
			HasNext:  false,
			HasPrev:  false,
		},
	}
	rows, err := r.dbContext.Query(
		queries.GetFilesQuery,
		pagination.PageSize,
		pagination.Page,
	)
	if err != nil {
		return paginationResponse, err
	}
	defer rows.Close()

	for rows.Next() {
		var image ImageModel
		if err := rows.Scan(
			&image.ID,
			&image.Name,
			&image.Path,
			&image.Format,
			&image.Size,
			&image.UpdatedAt,
			&image.CreatedAt,
			&image.LastInteraction,
			&image.LastBackup,
		); err != nil {
			return paginationResponse, err
		}
		paginationResponse.Items = append(paginationResponse.Items, image)
	}

	if len(paginationResponse.Items) > pagination.PageSize {
		paginationResponse.Items = paginationResponse.Items[:len(paginationResponse.Items)-1]
		paginationResponse.Pagination.HasNext = true
	}

	return paginationResponse, nil
}
