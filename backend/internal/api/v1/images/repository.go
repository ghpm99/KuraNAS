package images

import (
	"database/sql"
	"nas-go/api/pkg/database/queries"
)

type Repository struct {
	dbContext *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{database}
}

func (r *Repository) GetAllImages() ([]Image, error) {
	rows, err := r.dbContext.Query(queries.GetFilesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var image Image
		if err := rows.Scan(&image.ID, &image.Name, &image.Path); err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}
