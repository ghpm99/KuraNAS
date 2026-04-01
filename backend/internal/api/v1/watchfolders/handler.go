package watchfolders

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(service ServiceInterface, loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{service: service, logService: loggerService}
}

func (handler *Handler) GetWatchFoldersHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "GetWatchFolders",
		Description: "Fetching watch folders",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	watchFolders, err := handler.service.GetWatchFolders()
	if err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("SETTINGS_WATCH_FOLDER_SAVE_ERROR")})
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, watchFolders)
}

func (handler *Handler) CreateWatchFolderHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "CreateWatchFolder",
		Description: "Creating watch folder",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	var request CreateWatchFolderDto
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	created, err := handler.service.CreateWatchFolder(request)
	if err != nil {
		handler.completeError(logModel, err)
		switch {
		case errors.Is(err, ErrPathNotExists):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("WATCH_FOLDER_PATH_NOT_EXISTS")})
		case errors.Is(err, ErrPathIsSubfolderOfEntryPoint):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("WATCH_FOLDER_PATH_IS_ENTRY_POINT")})
		case errors.Is(err, ErrPathAlreadyWatched):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("WATCH_FOLDER_PATH_ALREADY_WATCHED")})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("SETTINGS_WATCH_FOLDER_SAVE_ERROR")})
		}
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusCreated, created)
}

func (handler *Handler) UpdateWatchFolderHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "UpdateWatchFolder",
		Description: "Updating watch folder",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		handler.completeError(logModel, ErrInvalidWatchFolderID)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	var request UpdateWatchFolderDto
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	updated, err := handler.service.UpdateWatchFolder(id, request)
	if err != nil {
		handler.completeError(logModel, err)
		switch {
		case errors.Is(err, ErrWatchFolderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("WATCH_FOLDER_NOT_FOUND")})
		case errors.Is(err, ErrPathNotExists):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("WATCH_FOLDER_PATH_NOT_EXISTS")})
		case errors.Is(err, ErrPathIsSubfolderOfEntryPoint):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("WATCH_FOLDER_PATH_IS_ENTRY_POINT")})
		case errors.Is(err, ErrPathAlreadyWatched):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("WATCH_FOLDER_PATH_ALREADY_WATCHED")})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("SETTINGS_WATCH_FOLDER_SAVE_ERROR")})
		}
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, updated)
}

func (handler *Handler) DeleteWatchFolderHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "DeleteWatchFolder",
		Description: "Deleting watch folder",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		handler.completeError(logModel, ErrInvalidWatchFolderID)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	err = handler.service.DeleteWatchFolder(id)
	if err != nil {
		handler.completeError(logModel, err)
		switch {
		case errors.Is(err, ErrWatchFolderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("WATCH_FOLDER_NOT_FOUND")})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("SETTINGS_WATCH_FOLDER_SAVE_ERROR")})
		}
		return
	}

	handler.completeSuccess(logModel)
	c.AbortWithStatus(http.StatusNoContent)
}

func (handler *Handler) createLog(logModel logger.LoggerModel) logger.LoggerModel {
	if handler.logService == nil {
		return logModel
	}

	createdLog, err := handler.logService.CreateLog(logModel, nil)
	if err != nil {
		return logModel
	}
	return createdLog
}

func (handler *Handler) completeSuccess(logModel logger.LoggerModel) {
	if handler.logService != nil {
		_ = handler.logService.CompleteWithSuccessLog(logModel)
	}
}

func (handler *Handler) completeError(logModel logger.LoggerModel, err error) {
	if handler.logService != nil {
		_ = handler.logService.CompleteWithErrorLog(logModel, err)
	}
}
