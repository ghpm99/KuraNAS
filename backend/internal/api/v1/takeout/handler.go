package takeout

import (
	"errors"
	"fmt"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(service ServiceInterface, loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{
		service:    service,
		logService: loggerService,
	}
}

func (h *Handler) InitUploadHandler(c *gin.Context) {
	logModel := h.createLog(logger.LoggerModel{
		Name:        "InitTakeoutUpload",
		Description: "Initializing takeout chunked upload",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	var dto InitTakeoutUploadDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		h.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	result, err := h.service.InitUpload(dto)
	if err != nil {
		h.completeError(logModel, err)
		switch {
		case errors.Is(err, ErrInvalidUploadSize):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_SIZE_INVALID")})
			return
		case errors.Is(err, ErrInvalidUploadFileName):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_FILE_NAME_INVALID")})
			return
		case errors.Is(err, ErrInsufficientStorage):
			c.JSON(http.StatusInsufficientStorage, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_INSUFFICIENT_STORAGE")})
			return
		}
		applog.ErrorWithStack("takeout: failed to init upload session", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_INIT_ERROR")})
		return
	}

	h.completeSuccess(logModel)
	c.JSON(http.StatusOK, result)
}

func (h *Handler) UploadChunkHandler(c *gin.Context) {
	logModel := h.createLog(logger.LoggerModel{
		Name:        "UploadTakeoutChunk",
		Description: "Uploading chunk for takeout import",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	file, err := c.FormFile("chunk")
	if err != nil {
		h.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	dto := UploadTakeoutChunkDto{
		UploadID: c.PostForm("upload_id"),
	}
	if dto.UploadID == "" {
		h.completeError(logModel, fmt.Errorf("upload_id is required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if offset := c.PostForm("offset"); offset != "" {
		parsed, parseErr := strconv.ParseInt(offset, 10, 64)
		if parseErr != nil {
			h.completeError(logModel, parseErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
			return
		}
		dto.Offset = parsed
	}

	if err := h.service.UploadChunk(file, dto); err != nil {
		h.completeError(logModel, err)
		switch {
		case errors.Is(err, ErrUploadSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("TAKEOUT_SESSION_NOT_FOUND")})
		case errors.Is(err, ErrUploadOffsetMismatch):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("TAKEOUT_OFFSET_MISMATCH")})
		case errors.Is(err, ErrUploadChunkExceedsSize):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_CHUNK_EXCEEDS_SIZE")})
		default:
			applog.ErrorWithStack("takeout: failed to store upload chunk", err, "upload_id", dto.UploadID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_CHUNK_ERROR")})
		}
		return
	}

	h.completeSuccess(logModel)
	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *Handler) CompleteUploadHandler(c *gin.Context) {
	logModel := h.createLog(logger.LoggerModel{
		Name:        "CompleteTakeoutUpload",
		Description: "Completing takeout upload and scheduling import",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	var dto CompleteTakeoutUploadDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		h.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	result, err := h.service.CompleteUpload(dto)
	if err != nil {
		h.completeError(logModel, err)
		switch {
		case errors.Is(err, ErrUploadSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("TAKEOUT_SESSION_NOT_FOUND")})
		case errors.Is(err, ErrUploadAlreadyCompleted):
			c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_ALREADY_COMPLETED")})
		case errors.Is(err, ErrUploadIncomplete):
			c.JSON(http.StatusBadRequest, gin.H{"error": incompleteUploadMessage(err)})
		case errors.Is(err, ErrInvalidZipFile):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("TAKEOUT_INVALID_ZIP")})
		default:
			applog.ErrorWithStack("takeout: failed to complete upload", err, "upload_id", dto.UploadID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_COMPLETE_ERROR")})
		}
		return
	}

	h.completeSuccess(logModel)
	c.JSON(http.StatusAccepted, result)
}

func (h *Handler) GetUploadStatusHandler(c *gin.Context) {
	logModel := h.createLog(logger.LoggerModel{
		Name:        "GetTakeoutUploadStatus",
		Description: "Reading takeout upload session progress",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	status, err := h.service.GetUploadStatus(c.Param("upload_id"))
	if err != nil {
		h.completeError(logModel, err)
		if errors.Is(err, ErrUploadSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("TAKEOUT_SESSION_NOT_FOUND")})
			return
		}
		applog.ErrorWithStack("takeout: failed to read upload status", err, "upload_id", c.Param("upload_id"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_STATUS_ERROR")})
		return
	}

	h.completeSuccess(logModel)
	c.JSON(http.StatusOK, status)
}

func (h *Handler) CancelUploadHandler(c *gin.Context) {
	logModel := h.createLog(logger.LoggerModel{
		Name:        "CancelTakeoutUpload",
		Description: "Cancelling takeout upload session",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	if err := h.service.CancelUpload(c.Param("upload_id")); err != nil {
		h.completeError(logModel, err)
		if errors.Is(err, ErrUploadSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("TAKEOUT_SESSION_NOT_FOUND")})
			return
		}
		applog.ErrorWithStack("takeout: failed to cancel upload session", err, "upload_id", c.Param("upload_id"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("TAKEOUT_UPLOAD_CANCEL_ERROR")})
		return
	}

	h.completeSuccess(logModel)
	c.Status(http.StatusNoContent)
}

func incompleteUploadMessage(err error) string {
	var incompleteError UploadIncompleteError
	if !errors.As(err, &incompleteError) {
		return i18n.GetMessage("TAKEOUT_UPLOAD_INCOMPLETE_GENERIC")
	}
	return i18n.Translate("TAKEOUT_UPLOAD_INCOMPLETE", incompleteError.ExpectedSize, incompleteError.ReceivedSize)
}

func (h *Handler) createLog(logModel logger.LoggerModel) logger.LoggerModel {
	if h.logService == nil {
		return logModel
	}

	createdLog, err := h.logService.CreateLog(logModel, nil)
	if err != nil {
		return logModel
	}

	return createdLog
}

func (h *Handler) completeSuccess(logModel logger.LoggerModel) {
	if h.logService != nil {
		_ = h.logService.CompleteWithSuccessLog(logModel)
	}
}

func (h *Handler) completeError(logModel logger.LoggerModel, err error) {
	if h.logService != nil {
		_ = h.logService.CompleteWithErrorLog(logModel, err)
	}
}
