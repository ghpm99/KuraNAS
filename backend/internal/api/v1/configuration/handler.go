package configuration

import (
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
