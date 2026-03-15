package configuration

import (
	"database/sql"
	"errors"
	"nas-go/api/api"
	"nas-go/api/internal/config"
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

func (handler *Handler) GetTranslationJson(c *gin.Context) {
	loggerModel := handler.createLog(logger.LoggerModel{
		Name:        "GetTranslation",
		Description: "Fetching translation file",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	filePath := i18n.GetPathFileTranslate()
	if handler.service != nil {
		if runtimePath, err := handler.service.GetTranslationFilePath(); err == nil {
			filePath = runtimePath
		}
	}

	handler.completeSuccess(loggerModel)
	c.File(filePath)
}

func (handler *Handler) GetAboutHandler(c *gin.Context) {
	loggerModel := handler.createLog(logger.LoggerModel{
		Name:        "GetAbout",
		Description: "Fetching about information",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	currentLang := config.AppConfig.Lang
	if handler.service != nil {
		if settings, err := handler.service.GetSettings(); err == nil {
			currentLang = settings.Language.Current
		}
	}

	c.JSON(200, gin.H{
		"version":        api.Version,
		"commit_hash":    api.CommitHash,
		"platform":       config.GetBuildConfig("BuildVersion"),
		"path":           config.AppConfig.EntryPoint,
		"lang":           currentLang,
		"enable_workers": config.AppConfig.EnableWorkers,
		"statup_time":    config.AppConfig.StartupTime.Format("2006-01-02 15:04:05"),
		"gin_mode":       gin.Mode(),
		"gin_version":    gin.Version,
		"go_version":     api.GoVersion,
		"node_version":   api.NodeVersion,
	})
	handler.completeSuccess(loggerModel)
}

func (handler *Handler) GetSettingsHandler(c *gin.Context) {
	loggerModel := handler.createLog(logger.LoggerModel{
		Name:        "GetSettings",
		Description: "Fetching settings",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	if handler.service == nil {
		handler.completeError(loggerModel, errors.New("configuration service is nil"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		return
	}

	settings, err := handler.service.GetSettings()
	if err != nil {
		handler.completeError(loggerModel, err)
		respondConfigurationError(c, err, "ERROR_CONFIGURATION_LOAD_FAILED")
		return
	}

	handler.completeSuccess(loggerModel)
	c.JSON(http.StatusOK, settings)
}

func (handler *Handler) UpdateSettingsHandler(c *gin.Context) {
	loggerModel := handler.createLog(logger.LoggerModel{
		Name:        "UpdateSettings",
		Description: "Updating settings",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	if handler.service == nil {
		handler.completeError(loggerModel, errors.New("configuration service is nil"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_UPDATE_FAILED")})
		return
	}

	var request UpdateSettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	settings, err := handler.service.UpdateSettings(request)
	if err != nil {
		handler.completeError(loggerModel, err)
		respondConfigurationError(c, err, "ERROR_CONFIGURATION_UPDATE_FAILED")
		return
	}

	handler.completeSuccess(loggerModel)
	c.JSON(http.StatusOK, settings)
}

func respondConfigurationError(c *gin.Context, err error, defaultMessageKey string) {
	switch {
	case errors.Is(err, ErrInvalidSettingsRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage(defaultMessageKey)})
	}
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
