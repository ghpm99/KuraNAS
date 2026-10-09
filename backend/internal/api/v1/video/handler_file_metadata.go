package video

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type VideoSummaryHandler struct {
	service    VideoSummaryServiceInterface
	logService logger.LoggerServiceInterface
}

func NewVideoSummaryHandler(service VideoSummaryServiceInterface, logService logger.LoggerServiceInterface) *VideoSummaryHandler {
	return &VideoSummaryHandler{service: service, logService: logService}
}

func (h *VideoSummaryHandler) GetVideoSummaryHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        "GetVideoFileMetadata",
		Description: "Fetching video metadata summary",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	fileID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	loggerModel.SetExtraData(logger.LogExtraData{Data: map[string]int{"file_id": fileID}})

	summary, err := h.service.GetVideoSummary(fileID)
	if errors.Is(err, sql.ErrNoRows) {
		message := i18n.GetMessage("ERROR_FILE_METADATA_NOT_FOUND")
		h.logService.CompleteWithErrorLog(loggerModel, fmt.Errorf("%s", message))
		c.JSON(http.StatusNotFound, gin.H{"error": message})
		return
	}
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, summary)
}
