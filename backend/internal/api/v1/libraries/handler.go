package libraries

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(service ServiceInterface, loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{
		service:    service,
		logService: loggerService,
	}
}

func (handler *Handler) GetLibrariesHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "GetLibraries",
		Description: "Fetching configured library paths",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	libraries, err := handler.service.GetLibraries()
	if err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_GET_FILES")})
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, libraries)
}

func (handler *Handler) UpdateLibraryHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "UpdateLibrary",
		Description: "Updating library path",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	category := LibraryCategory(c.Param("category"))
	if !category.IsValid() {
		handler.completeError(logModel, ErrInvalidCategory)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.Translate("LIBRARY_INVALID_CATEGORY", c.Param("category"))})
		return
	}

	var request UpdateLibraryDto
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	library, err := handler.service.UpdateLibrary(category, request)
	if err != nil {
		handler.completeError(logModel, err)
		switch {
		case errors.Is(err, ErrInvalidCategory):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.Translate("LIBRARY_INVALID_CATEGORY", c.Param("category"))})
		case errors.Is(err, ErrPathNotSubfolder):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("LIBRARY_PATH_NOT_SUBFOLDER")})
		case errors.Is(err, ErrPathNotExists):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("LIBRARY_PATH_NOT_EXISTS")})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("SETTINGS_LIBRARY_SAVE_ERROR")})
		}
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, library)
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
