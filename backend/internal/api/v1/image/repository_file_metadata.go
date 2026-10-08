package image

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/image"
)

type ImageSummaryRepositoryInterface interface {
	GetImageSummaryByFileID(fileID int) (ImageSummaryDto, error)
}

type ImageSummaryRepository struct {
	Db *database.DbContext
}

func NewImageSummaryRepository(db *database.DbContext) *ImageSummaryRepository {
	return &ImageSummaryRepository{Db: db}
}

func (r *ImageSummaryRepository) GetImageSummaryByFileID(fileID int) (ImageSummaryDto, error) {
	var summary ImageSummaryDto

	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetImageSummaryByFileIDQuery, fileID)
		return row.Scan(
			&summary.Width,
			&summary.Height,
			&summary.Make,
			&summary.Model,
			&summary.LensModel,
			&summary.DateTimeOriginal,
			&summary.ExposureTime,
			&summary.FNumber,
			&summary.ISO,
			&summary.FocalLength,
		)
	})
	if err != nil {
		return ImageSummaryDto{}, fmt.Errorf("GetImageSummaryByFileID: %w", err)
	}
	return summary, nil
}
