package accesscontrol

import (
	"errors"
	"net/http"
	"net/netip"
	"strconv"

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
	case errors.Is(err, ErrInvalidCIDR), errors.Is(err, ErrEmptyAllowedIPInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ACCESS_CONTROL_INVALID_CIDR")})
	case errors.Is(err, ErrDuplicateAllowedIP):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ACCESS_CONTROL_DUPLICATE_ENTRY")})
	case errors.Is(err, ErrAllowedIPNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ACCESS_CONTROL_ENTRY_NOT_FOUND")})
	case errors.Is(err, ErrInvalidAllowedIPID):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ACCESS_CONTROL_SAVE_ERROR")})
	}
}

func (handler *Handler) GetAllowedIPsHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "GetAllowedIPs",
		Description: "Fetching allowed IPs",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	allowedIPs, err := handler.service.GetAllowedIPs()
	if err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ACCESS_CONTROL_SAVE_ERROR")})
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, allowedIPs)
}

func (handler *Handler) CreateAllowedIPHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "CreateAllowedIP",
		Description: "Creating allowed IP",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	var request CreateAllowedIPDto
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	created, err := handler.service.CreateAllowedIP(request)
	if err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusCreated, created)
}

func (handler *Handler) UpdateAllowedIPHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "UpdateAllowedIP",
		Description: "Updating allowed IP",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		handler.completeError(logModel, ErrInvalidAllowedIPID)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	var request UpdateAllowedIPDto
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	updated, err := handler.service.UpdateAllowedIP(id, request)
	if err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, updated)
}

func (handler *Handler) DeleteAllowedIPHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "DeleteAllowedIP",
		Description: "Deleting allowed IP",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		handler.completeError(logModel, ErrInvalidAllowedIPID)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := handler.service.DeleteAllowedIP(id); err != nil {
		handler.completeError(logModel, err)
		handler.respondServiceError(c, err)
		return
	}

	handler.completeSuccess(logModel)
	c.Status(http.StatusNoContent)
}

// GetClientIPHandler returns the requester's connection IP so the Settings
// screen can offer one-click registration of the current device.
func (handler *Handler) GetClientIPHandler(c *gin.Context) {
	remoteIP := c.RemoteIP()
	if addr, err := netip.ParseAddr(remoteIP); err == nil {
		remoteIP = addr.Unmap().String()
	}
	c.JSON(http.StatusOK, gin.H{"ip": remoteIP})
}
