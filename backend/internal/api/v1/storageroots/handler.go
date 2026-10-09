package storageroots

import (
	"errors"
	"net/http"
	"strconv"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(service ServiceInterface, loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{service: service, logService: loggerService}
}

func (handler *Handler) createLog(logModel logger.LoggerModel) logger.LoggerModel {
	if handler.logService == nil {
		return logModel
	}
	created, err := handler.logService.CreateLog(logModel, nil)
	if err != nil {
		return logModel
	}
	return created
}

func (handler *Handler) completeSuccess(logModel logger.LoggerModel) {
	if handler.logService == nil {
		return
	}
	_ = handler.logService.CompleteWithSuccessLog(logModel)
}

func (handler *Handler) completeError(logModel logger.LoggerModel, err error) {
	if handler.logService == nil {
		return
	}
	_ = handler.logService.CompleteWithErrorLog(logModel, err)
}

func (handler *Handler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrRootNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("STORAGE_ROOT_NOT_FOUND")})
	case errors.Is(err, ErrInvalidRootPath):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("STORAGE_ROOT_INVALID_PATH")})
	case errors.Is(err, ErrInvalidRootLabel):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("STORAGE_ROOT_INVALID_LABEL")})
	case errors.Is(err, ErrOverlappingRoot):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("STORAGE_ROOT_OVERLAP")})
	case errors.Is(err, ErrDuplicateRoot):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("STORAGE_ROOT_DUPLICATE")})
	case errors.Is(err, ErrPrimaryRootImmutable):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("STORAGE_ROOT_PRIMARY_IMMUTABLE")})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("STORAGE_ROOT_OPERATION_FAILED")})
	}
}

func parseRootID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func (handler *Handler) GetStorageRootsHandler(c *gin.Context) {
	rootsList, err := handler.service.GetRoots()
	if err != nil {
		applog.ErrorWithStack("storageroots: list roots failed", err, "ip", c.ClientIP())
		handler.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, rootsList)
}

func (handler *Handler) CreateStorageRootHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "CreateStorageRoot",
		Description: "Registering storage root",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	var request CreateStorageRootDto
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	created, err := handler.service.CreateRoot(request)
	if err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusCreated, created)
}

func (handler *Handler) UpdateStorageRootHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "UpdateStorageRoot",
		Description: "Updating storage root",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	id, ok := parseRootID(c)
	if !ok {
		handler.completeError(logModel, ErrRootNotFound)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	var request UpdateStorageRootDto
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	updated, err := handler.service.UpdateRoot(id, request)
	if err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, updated)
}

func (handler *Handler) DeleteStorageRootHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "DeleteStorageRoot",
		Description: "Unregistering storage root",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	id, ok := parseRootID(c)
	if !ok {
		handler.completeError(logModel, ErrRootNotFound)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := handler.service.DeleteRoot(id); err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.Status(http.StatusNoContent)
}
