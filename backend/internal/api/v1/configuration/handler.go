package configuration

import (
	"nas-go/api/api"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	logService logger.LoggerServiceInterface
}

func NewHandler(loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{
		logService: loggerService,
	}
}

func (handler *Handler) GetTranslationJson(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetTranslation",
		Description: "Fetching translation file",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	filePath, err := i18n.GetPathFileTranslate()
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.File(filePath)
}

func (handler *Handler) GetAboutHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetAbout",
		Description: "Fetching about information",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	c.JSON(200, gin.H{
		"version":        api.Version,
		"commit_hash":    api.CommitHash,
		"platform":       config.GetBuildConfig("BuildVersion"),
		"path":           config.AppConfig.EntryPoint,
		"lang":           config.AppConfig.Lang,
		"enable_workers": config.AppConfig.EnableWorkers,
		"statup_time":    config.AppConfig.StartupTime.Format("2006-01-02 15:04:05"),
		"gin_mode":       gin.Mode(),
		"gin_version":    gin.Version,
		"go_version":     api.GoVersion,
		"node_version":   api.NodeVersion,
	})
	handler.logService.CompleteWithSuccessLog(loggerModel)
}
