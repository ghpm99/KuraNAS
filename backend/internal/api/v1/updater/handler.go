package updater

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

func (handler *Handler) GetUpdateStatusHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetUpdateStatus",
		Description: "Checking for updates",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	status, err := handler.service.CheckForUpdate()
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_UPDATE_STATUS")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, status)
}

func (handler *Handler) ApplyUpdateHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "ApplyUpdate",
		Description: "Downloading and applying update",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	if err := handler.service.DownloadAndApply(); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		if translationKey, isRefusal := refusalTranslationKey(err); isRefusal {
			c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage(translationKey)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_UPDATE_APPLY")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("UPDATE_APPLIED")})
}

func refusalTranslationKey(err error) (string, bool) {
	switch {
	case errors.Is(err, errChecksumMissing):
		return "ERROR_UPDATE_CHECKSUM_MISSING", true
	case errors.Is(err, errChecksumNoEntry):
		return "ERROR_UPDATE_CHECKSUM_NO_ENTRY", true
	case errors.Is(err, errChecksumMismatch):
		return "ERROR_UPDATE_CHECKSUM_MISMATCH", true
	case errors.Is(err, errPackageTooLarge):
		return "ERROR_UPDATE_PACKAGE_TOO_LARGE", true
	}
	return "", false
}
