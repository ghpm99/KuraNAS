package files

import (
	"nas-go/api/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) GetUploadStatusHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetUploadStatus",
		Description: "Reading upload progress",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	status, err := handler.service.GetUploadSessionStatus(c.Param("upload_id"))
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_UPLOAD_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{
		"upload_id":     status.UploadID,
		"size":          status.Size,
		"received_size": status.ReceivedSize,
	})
}
