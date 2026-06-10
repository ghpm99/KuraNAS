package files

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	queries "nas-go/api/pkg/database/queries/file"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"
)

func (handler *Handler) GetTotalSpaceUsedHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetTotalSpaceUsed",
		Description: "Fetching total space used",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	totalSpaceUsed, err := handler.service.GetTotalSpaceUsed()

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"total_space_used": totalSpaceUsed})
}

func (handler *Handler) GetTotalFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetTotalFiles",
		Description: "Fetching total files count",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	totalFiles, err := handler.service.GetTotalFiles()

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"total_files": totalFiles})
}

func (handler *Handler) GetTotalDirectoryHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetTotalSpaceUsedByPath",
		Description: "Fetching total space used by path",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	totalSpaceUsed, err := handler.service.GetTotalDirectory()

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"total_directory": totalSpaceUsed})
}

func (handler *Handler) GetReportSizeByFormatHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetReportSizeByFormat",
		Description: "Fetching report size by format",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	report, err := handler.service.GetReportSizeByFormat()

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, report)
}

func (handler *Handler) GetTopFilesBySizeHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetTopFilesBySize",
		Description: "Fetching top files by size",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	limit := utils.ParseInt(c.DefaultQuery("limit", "5"), c)

	topFiles, err := handler.service.GetTopFilesBySize(limit)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	responseFiles := make([]FileDto, len(topFiles))
	for i, f := range topFiles {
		responseFiles[i] = f.ToResponse()
	}
	c.JSON(http.StatusOK, responseFiles)
}

func (handler *Handler) GetDuplicateFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetDuplicateFiles",
		Description: "Fetching duplicate files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	report, err := handler.service.GetDuplicateFiles(page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, report)
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
