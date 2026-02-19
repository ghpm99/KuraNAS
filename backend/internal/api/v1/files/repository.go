package files

import (
	"database/sql"
	"errors"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/file"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(database *database.DbContext) *Repository {
	return &Repository{database}
}

func (r *Repository) GetDbContext() *database.DbContext {
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
		return paginationResponse, fmt.Errorf("falha na consulta de arquivos: %w", err)
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
		return 0, fmt.Errorf("falha ao obter contagem de diretório: %w", err)
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
		return 0, fmt.Errorf("falha ao obter espaço total usado: %w", err)
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
		return nil, fmt.Errorf("falha ao obter relatório por formato: %w", err)
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
		return nil, fmt.Errorf("falha ao obter top arquivos por tamanho: %w", err)
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
		return paginationResponse, fmt.Errorf("falha ao obter arquivos duplicados: %w", err)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}

func (r *Repository) GetImages(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {

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
		pq.Array(utils.ImageFormats),
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.GetImagesQuery,
			args...,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file FileModel
			var metadata ImageMetadataModel
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
		return paginationResponse, fmt.Errorf("falha na consulta de arquivos: %w", err)
	}

	paginationResponse.UpdatePagination()
	fmt.Println("Paginacao length", len(paginationResponse.Items))

	return paginationResponse, nil
}
