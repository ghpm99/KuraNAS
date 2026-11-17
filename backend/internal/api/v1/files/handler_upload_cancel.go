package files

import (
	"nas-go/api/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) CancelUploadHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "CancelUpload",
		Description: "Cancelling upload",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	if err := handler.service.CancelUpload(c.Param("upload_id")); err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_UPLOAD_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Status(http.StatusNoContent)
}
