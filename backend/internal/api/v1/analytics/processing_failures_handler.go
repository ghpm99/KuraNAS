package analytics

import (
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultProcessingFailuresPageSize = 20

func ParseProcessingStage(rawStage string) (ProcessingStage, bool) {
	if rawStage == "" {
		return "", true
	}
	stage := ProcessingStage(rawStage)
	_, isKnownStage := processingFailureMessageKeys[stage]
	return stage, isKnownStage
}

func (handler *Handler) GetProcessingFailuresHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultProcessingFailuresPageSize)
	if !isPaginationValid {
		return
	}
	stage, isStageValid := ParseProcessingStage(c.Query("stage"))
	if !isStageValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_ANALYTICS_INVALID_PROCESSING_STAGE")})
		return
	}
	failuresPage, err := handler.service.GetProcessingFailuresPage(stage, page, pageSize)
	handler.respond(c, failuresPage, err)
}
