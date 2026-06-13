package files

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"nas-go/api/internal/roots"
	queries "nas-go/api/pkg/database/queries/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"
)

func (handler *Handler) GetFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFiles",
		Description: "Fetching files with filter",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]int{"page": page, "page_size": pageSize},
	})

	pagination, err := handler.service.GetActiveFilesPage(page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, ParsePaginationToResponse(pagination))
}

func (handler *Handler) GetFilesByPathHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFilesByPath",
		Description: "Fetching files by path",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	rawPath := c.DefaultQuery("path", "")
	path := roots.ToAbsolutePath(rawPath)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]string{"path": path},
	})

	pagination, err := handler.service.GetFilesByPath(path, page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, ParsePaginationToResponse(pagination))
}

func (handler *Handler) GetChildrenByIdHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetChildrenById",
		Description: "Fetching files by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)
	id := utils.ParseInt(c.Param("id"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]int{"id": id},
	})

	file, err := handler.service.GetFileById(id)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}
	// Soft-deleted rows are not part of the tree: same 404 the active-only
	// filter used to produce.
	if errors.Is(err, sql.ErrNoRows) || file.DeletedAt.HasValue {
		notFoundErr := fmt.Errorf("%s", i18n.GetMessage("ERROR_FILE_NOT_FOUND"))
		handler.Logger.CompleteWithErrorLog(loggerModel, notFoundErr)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	pagination, err := handler.service.GetFilesByPath(file.Path, page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, ParsePaginationToResponse(pagination))
}

func (handler *Handler) UpdateFilesHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "UpdateFiles",
		Description: "Updating files with data",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	data := c.PostForm("data")
	if data == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("data is required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DATA_REQUIRED")})
		return
	}
	loggerModel.SetExtraData(logger.LogExtraData{
		Data: data,
	})
	handler.service.ScanFilesTask(data)
	handler.Logger.CompleteWithSuccessLog(loggerModel)
}

func (handler *Handler) GetFilesTreeHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFilesTree",
		Description: "Fetching files with filter",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	fileParentId := utils.ParseInt(c.DefaultQuery("file_parent", "0"), c)

	fileCategory := FileCategory(c.DefaultQuery("category", string(AllCategory)))

	// Level zero with multiple roots: the tree starts at the list of storage
	// roots, each a navigable directory node. With a single root the legacy
	// shape is kept — level zero lists the root's children directly.
	if fileParentId == 0 && len(roots.Enabled()) > 1 {
		nodes, err := handler.service.GetRootNodes()
		if err != nil {
			handler.Logger.CompleteWithErrorLog(loggerModel, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
			return
		}

		response := utils.PaginationResponse[FileDto]{
			Items:      nodes,
			Pagination: utils.Pagination{Page: 1, PageSize: pageSize},
		}
		handler.Logger.CompleteWithSuccessLog(loggerModel)
		c.JSON(http.StatusOK, ParsePaginationToResponse(response))
		return
	}

	var parentPath string
	if primary, ok := roots.Primary(); ok {
		parentPath = primary.Path
	}
	if fileParentId != 0 {
		fileParent, err := handler.service.GetFileById(fileParentId)
		if err != nil {
			handler.Logger.CompleteWithErrorLog(loggerModel, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
			return
		}
		if fileParent.ID != 0 {
			parentPath = fileParent.Path
		}
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]string{"parent_path": parentPath, "category": string(fileCategory)},
	})

	pagination, err := handler.service.GetChildrenByParentPath(parentPath, fileCategory, page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, ParsePaginationToResponse(pagination))
}

// GetFileStatByPath returns the size and last-known modification time of the
// active (non-deleted) file at the given absolute path. The second return value
// reports whether such a row exists. It runs a single indexed lookup so the
// diff scan can compare one file at a time without batch-loading the table.
func (r *Repository) GetFileStatByPath(path string) (FileStat, bool, error) {
	var stat FileStat
	found := false

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		scanErr := tx.QueryRow(queries.GetFileStatByPathQuery, path).Scan(&stat.Size, &stat.UpdatedAt)
		if scanErr != nil {
			if errors.Is(scanErr, sql.ErrNoRows) {
				return nil
			}
			return scanErr
		}
		found = true
		return nil
	})
	if err != nil {
		return FileStat{}, false, fmt.Errorf("GetFileStatByPath: %w", err)
	}

	return stat, found, nil
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
		&file.PhysicalPath,
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
		return 0, fmt.Errorf("failed to get directory count: %w", err)
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
