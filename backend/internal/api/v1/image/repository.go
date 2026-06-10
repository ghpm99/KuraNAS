package image

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/image"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

// Repository is the image-domain data-access implementation. It is allowed to
// JOIN the home_file table because a package is not the owner of a table.
type Repository struct {
	Db *database.DbContext
}

func NewRepository(db *database.DbContext) *Repository {
	return &Repository{Db: db}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.Db
}

// GetImageMetadataByID loads image_metadata by its primary key.
func (r *Repository) GetImageMetadataByID(id int) (MetadataModel, error) {
	var metadata MetadataModel

	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetImageMetadataByIDQuery, id)

		if err := row.Scan(
			&metadata.ID,
			&metadata.FileId,
			&metadata.Path,
			&metadata.Format,
			&metadata.Mode,
			&metadata.Width,
			&metadata.Height,
			&metadata.DPIX,
			&metadata.DPIY,
			&metadata.XResolution,
			&metadata.YResolution,
			&metadata.ResolutionUnit,
			&metadata.Orientation,
			&metadata.Compression,
			&metadata.Photometric,
			&metadata.ColorSpace,
			&metadata.ComponentsConfig,
			&metadata.ICCProfile,
			&metadata.Make,
			&metadata.Model,
			&metadata.Software,
			&metadata.LensModel,
			&metadata.SerialNumber,
			&metadata.DateTime,
			&metadata.DateTimeOriginal,
			&metadata.DateTimeDigitized,
			&metadata.SubSecTime,
			&metadata.ExposureTime,
			&metadata.FNumber,
			&metadata.ISO,
			&metadata.ShutterSpeed,
			&metadata.ApertureValue,
			&metadata.BrightnessValue,
			&metadata.ExposureBias,
			&metadata.MeteringMode,
			&metadata.Flash,
			&metadata.FocalLength,
			&metadata.WhiteBalance,
			&metadata.ExposureProgram,
			&metadata.MaxApertureValue,
			&metadata.GPSLatitude,
			&metadata.GPSLongitude,
			&metadata.GPSAltitude,
			&metadata.GPSDate,
			&metadata.GPSTime,
			&metadata.ImageDescription,
			&metadata.UserComment,
			&metadata.Copyright,
			&metadata.Artist,
			&metadata.Classification.Category,
			&metadata.Classification.Confidence,
			&metadata.Classification.SuggestedName,
			&metadata.CreatedAt,
		); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return metadata, fmt.Errorf("falha ao obter metadados da imagem: %w", err)
	}

	return metadata, nil
}

// UpsertImageMetadata inserts or updates image_metadata within an existing transaction.
func (r *Repository) UpsertImageMetadata(tx *sql.Tx, metadata MetadataModel) (MetadataModel, error) {
	var id int
	var createdAt time.Time
	args := []any{
		metadata.FileId,
		metadata.Path,
		metadata.Format,
		metadata.Mode,
		metadata.Width,
		metadata.Height,
		metadata.DPIX,
		metadata.DPIY,
		metadata.XResolution,
		metadata.YResolution,
		metadata.ResolutionUnit,
		metadata.Orientation,
		metadata.Compression,
		metadata.Photometric,
		metadata.ColorSpace,
		metadata.ComponentsConfig,
		metadata.ICCProfile,
		metadata.Make,
		metadata.Model,
		metadata.Software,
		metadata.LensModel,
		metadata.SerialNumber,
		metadata.DateTime,
		metadata.DateTimeOriginal,
		metadata.DateTimeDigitized,
		metadata.SubSecTime,
		metadata.ExposureTime,
		metadata.FNumber,
		metadata.ISO,
		metadata.ShutterSpeed,
		metadata.ApertureValue,
		metadata.BrightnessValue,
		metadata.ExposureBias,
		metadata.MeteringMode,
		metadata.Flash,
		metadata.FocalLength,
		metadata.WhiteBalance,
		metadata.ExposureProgram,
		metadata.MaxApertureValue,
		metadata.GPSLatitude,
		metadata.GPSLongitude,
		metadata.GPSAltitude,
		metadata.GPSDate,
		metadata.GPSTime,
		metadata.ImageDescription,
		metadata.UserComment,
		metadata.Copyright,
		metadata.Artist,
		metadata.Classification.Category,
		metadata.Classification.Confidence,
		metadata.Classification.SuggestedName,
		time.Now(),
	}

	row := tx.QueryRow(queries.UpsertImageMetadataQuery, args...)

	err := row.Scan(&id, &createdAt)
	if err != nil {
		return metadata, err
	}

	metadata.ID = id
	metadata.CreatedAt = createdAt
	return metadata, nil
}

// DeleteImageMetadata removes a image_metadata row by ID.
func (r *Repository) DeleteImageMetadata(id int) error {
	err := r.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(queries.DeleteImageMetadataQuery, id)
		if err != nil {
			return fmt.Errorf("falha ao executar query de exclusão: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("falha ao deletar metadados da imagem: %w", err)
	}

	return nil
}

func imageOrderByClause(groupBy ImageGroupBy) string {
	switch groupBy {
	case ImageGroupByType:
		return `COALESCE(NULLIF(hf.format, ''), 'zzzz') ASC, hf.name ASC, hf.id DESC`
	case ImageGroupByName:
		return `hf.name ASC, hf.id DESC`
	case ImageGroupByDate:
		fallthrough
	default:
		return `COALESCE(NULLIF(im.datetime_original, ''), NULLIF(im.datetime, ''), to_char(hf.created_at, 'YYYY:MM:DD HH24:MI:SS')) DESC, hf.id DESC`
	}
}

func getImagesQueryByGroup(groupBy ImageGroupBy) string {
	return strings.Replace(queries.GetImagesQuery, "{{ORDER_BY}}", imageOrderByClause(groupBy), 1)
}

// GetImages returns a paginated list of image files joined with their metadata.
// It JOINs home_file (owned by files) and image_metadata (owned here).
func (r *Repository) GetImages(page int, pageSize int, groupBy ImageGroupBy) (utils.PaginationResponse[files.FileModel], error) {
	paginationResponse := utils.PaginationResponse[files.FileModel]{
		Items: []files.FileModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.ImageFormats),
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			getImagesQueryByGroup(groupBy),
			args...,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file files.FileModel
			var metadata MetadataModel
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
				&metadata.ID,
				&metadata.FileId,
				&metadata.Path,
				&metadata.Format,
				&metadata.Mode,
				&metadata.Width,
				&metadata.Height,
				&metadata.DPIX,
				&metadata.DPIY,
				&metadata.XResolution,
				&metadata.YResolution,
				&metadata.ResolutionUnit,
				&metadata.Orientation,
				&metadata.Compression,
				&metadata.Photometric,
				&metadata.ColorSpace,
				&metadata.ComponentsConfig,
				&metadata.ICCProfile,
				&metadata.Make,
				&metadata.Model,
				&metadata.Software,
				&metadata.LensModel,
				&metadata.SerialNumber,
				&metadata.DateTime,
				&metadata.DateTimeOriginal,
				&metadata.DateTimeDigitized,
				&metadata.SubSecTime,
				&metadata.ExposureTime,
				&metadata.FNumber,
				&metadata.ISO,
				&metadata.ShutterSpeed,
				&metadata.ApertureValue,
				&metadata.BrightnessValue,
				&metadata.ExposureBias,
				&metadata.MeteringMode,
				&metadata.Flash,
				&metadata.FocalLength,
				&metadata.WhiteBalance,
				&metadata.ExposureProgram,
				&metadata.MaxApertureValue,
				&metadata.GPSLatitude,
				&metadata.GPSLongitude,
				&metadata.GPSAltitude,
				&metadata.GPSDate,
				&metadata.GPSTime,
				&metadata.ImageDescription,
				&metadata.UserComment,
				&metadata.Copyright,
				&metadata.Artist,
				&metadata.Classification.Category,
				&metadata.Classification.Confidence,
				&metadata.Classification.SuggestedName,
				&metadata.CreatedAt,
			); err != nil {
				return err
			}

			file.Metadata = metadata

			paginationResponse.Items = append(paginationResponse.Items, file)
		}

		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query images: %w", err)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}
