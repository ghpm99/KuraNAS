package trash

import (
	"errors"
	"net/http"
	"strconv"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

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
	case errors.Is(err, ErrItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("TRASH_ITEM_NOT_FOUND")})
	case errors.Is(err, ErrRestoreConflict):
		c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("TRASH_RESTORE_CONFLICT")})
	case errors.Is(err, ErrInvalidRetention):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("TRASH_OPERATION_FAILED")})
	}
}

func parseItemID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func (handler *Handler) GetTrashItemsHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "GetTrashItems",
		Description: "Listing trash items",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	items, err := handler.service.GetItems(page, pageSize)
	if err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, items)
}

func (handler *Handler) RestoreTrashItemHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "RestoreTrashItem",
		Description: "Restoring item from trash",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	id, ok := parseItemID(c)
	if !ok {
		handler.completeError(logModel, ErrItemNotFound)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	restoredPath, err := handler.service.RestoreItem(id)
	if err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("TRASH_RESTORE_SUCCESS"), "path": restoredPath})
}

func (handler *Handler) DeleteTrashItemHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "DeleteTrashItem",
		Description: "Permanently deleting one trash item",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	id, ok := parseItemID(c)
	if !ok {
		handler.completeError(logModel, ErrItemNotFound)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := handler.service.DeleteItemPermanently(id); err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("TRASH_DELETE_SUCCESS")})
}

func (handler *Handler) EmptyTrashHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "EmptyTrash",
		Description: "Emptying the trash",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	purged, err := handler.service.EmptyTrash()
	if err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("TRASH_EMPTY_SUCCESS"), "purged": purged})
}

func (handler *Handler) GetTrashRetentionHandler(c *gin.Context) {
	days, err := handler.service.GetRetentionDays()
	if err != nil {
		handler.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, RetentionDto{Days: days})
}

func (handler *Handler) UpdateTrashRetentionHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "UpdateTrashRetention",
		Description: "Updating trash retention policy",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	var request RetentionDto
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := handler.service.SetRetentionDays(request.Days); err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, RetentionDto{Days: request.Days})
}
