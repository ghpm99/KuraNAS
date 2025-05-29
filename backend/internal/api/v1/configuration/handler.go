package configuration

import (
	"nas-go/api/api"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

func GetTranslationJson(c *gin.Context) {
	filePath, err := i18n.GetPathFileTranslate()
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.File(filePath)
}

func GetAboutHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"version":        api.Version,
		"commit_hash":    api.CommitHash,
		"platform":       config.GetBuildConfig("BuildVersion"),
		"path":           config.AppConfig.EntryPoint,
		"lang":           config.AppConfig.Lang,
		"enable_workers": config.AppConfig.EnableWorkers,
		"statup_time":    config.AppConfig.StartupTime.Format("2006-01-02 15:04:05"),
	})
}
