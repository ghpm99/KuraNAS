package files

import (
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type completeUploadRequest struct {
	UploadID string `json:"upload_id"`
}

type completeUploadResponse struct {
	File  UploadFileResult `json:"file"`
	JobID int              `json:"job_id"`
}

func (handler *Handler) CompleteUploadHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "CompleteUpload",
		Description: "Publishing resumable upload",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var request completeUploadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	completeResult, err := handler.service.CompleteUpload(request.UploadID)
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_UPLOAD_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusCreated, completeUploadResponse{File: toClientUploadFileResult(completeResult.File), JobID: completeResult.JobID})
}
