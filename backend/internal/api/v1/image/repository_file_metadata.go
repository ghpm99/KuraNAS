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
	var takenAt sql.NullTime
	var gpsLatitude, gpsLongitude sql.NullFloat64

	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetImageSummaryByFileIDQuery, fileID)
		scanErr := row.Scan(
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
			&summary.Software,
			&summary.Description,
			&takenAt,
			&gpsLatitude,
			&gpsLongitude,
			&summary.ClassificationConfidence,
			&summary.SuggestedName,
		)
		if scanErr != nil {
			return scanErr
		}
		if takenAt.Valid {
			summary.TakenAt = &takenAt.Time
		}
		if gpsLatitude.Valid && gpsLongitude.Valid {
			summary.GPSLatitude = &gpsLatitude.Float64
			summary.GPSLongitude = &gpsLongitude.Float64
		}
		return nil
	})
	if err != nil {
		return ImageSummaryDto{}, fmt.Errorf("GetImageSummaryByFileID: %w", err)
	}
	return summary, nil
}
