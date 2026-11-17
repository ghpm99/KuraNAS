package files

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) UploadChunkHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "UploadChunk",
		Description: "Receiving upload chunk",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	chunk, err := c.FormFile("chunk")
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_UPLOAD_CHUNK_REQUIRED")})
		return
	}
	offset, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("offset")), 10, 64)
	if err != nil || offset < 0 {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_UPLOAD_OFFSET_INVALID")})
		return
	}

	receivedSize, err := handler.service.UploadChunk(strings.TrimSpace(c.PostForm("upload_id")), offset, chunk)
	var offsetMismatch *UploadOffsetMismatchError
	if errors.As(err, &offsetMismatch) {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusConflict, gin.H{
			"error":         i18n.GetMessage("ERROR_UPLOAD_OFFSET_MISMATCH"),
			"received_size": offsetMismatch.ReceivedSize,
		})
		return
	}
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_UPLOAD_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"received_size": receivedSize})
}
