package files

import (
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type initUploadRequest struct {
	FileName       string     `json:"file_name"`
	Size           *int64     `json:"size"`
	TargetFolderID int        `json:"target_folder_id"`
	RelativePath   string     `json:"relative_path"`
	Checksum       string     `json:"checksum"`
	ModifiedAt     *time.Time `json:"modified_at"`
	OnConflict     string     `json:"on_conflict"`
}

type initUploadResponse struct {
	UploadID      string `json:"upload_id,omitempty"`
	ChunkSize     int64  `json:"chunk_size"`
	ReceivedSize  int64  `json:"received_size"`
	Resumed       bool   `json:"resumed"`
	AlreadyExists bool   `json:"already_exists"`
}

func (handler *Handler) InitUploadHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "InitUpload",
		Description: "Opening resumable upload",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var request initUploadRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.Size == nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	initResult, err := handler.service.InitUpload(UploadInitRequest{
		FileName:       request.FileName,
		Size:           *request.Size,
		TargetFolderID: request.TargetFolderID,
		RelativePath:   request.RelativePath,
		Checksum:       request.Checksum,
		ModifiedAt:     request.ModifiedAt,
		OnConflict:     UploadConflictPolicy(request.OnConflict),
	})
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_UPLOAD_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusCreated, initUploadResponse{
		UploadID:      initResult.UploadID,
		ChunkSize:     initResult.ChunkSize,
		ReceivedSize:  initResult.ReceivedSize,
		Resumed:       initResult.Resumed,
		AlreadyExists: initResult.AlreadyExists,
	})
}
