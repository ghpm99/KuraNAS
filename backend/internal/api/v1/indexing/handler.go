package indexing

import (
	"errors"
	"net/http"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(service ServiceInterface, loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{service: service, logService: loggerService}
}

// ReindexHandler serves POST /indexing/reindex: 202 with the job id, 400 for an
// invalid scope or missing identifier, 404 for an unknown root or folder and
// 409 (with the running job id) when the same scope is already being reindexed.
func (handler *Handler) ReindexHandler(c *gin.Context) {
	var request ReindexRequestDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_REINDEX_INVALID_SCOPE")})
		return
	}

	logModel := handler.createLog("Reindex", c.ClientIP())
	jobID, err := handler.service.Reindex(request)
	if err != nil {
		handler.respondReindexError(c, logModel, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusAccepted, ReindexJobDto{JobID: jobID})
}

// UpdateFilesHandler serves the legacy POST /files/update, equivalent to a
// reindex of every root. A reindex already running answers with its job id.
func (handler *Handler) UpdateFilesHandler(c *gin.Context) {
	if c.PostForm("data") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DATA_REQUIRED")})
		return
	}

	logModel := handler.createLog("UpdateFiles", c.ClientIP())
	jobID, err := handler.service.Reindex(ReindexRequestDto{Scope: ScopeAll})

	var alreadyRunning *AlreadyRunningError
	if errors.As(err, &alreadyRunning) {
		handler.completeSuccess(logModel)
		c.JSON(http.StatusOK, ReindexJobDto{JobID: alreadyRunning.JobID})
		return
	}
	if err != nil {
		handler.respondReindexError(c, logModel, err)
		return
	}

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, ReindexJobDto{JobID: jobID})
}

func (handler *Handler) respondReindexError(c *gin.Context, logModel logger.LoggerModel, err error) {
	var alreadyRunning *AlreadyRunningError
	switch {
	case errors.Is(err, ErrInvalidScope):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_REINDEX_INVALID_SCOPE")})
	case errors.Is(err, ErrIdentifierRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_REINDEX_IDENTIFIER_REQUIRED")})
	case errors.Is(err, ErrRootNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_REINDEX_ROOT_NOT_FOUND")})
	case errors.Is(err, ErrFolderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_REINDEX_FOLDER_NOT_FOUND")})
	case errors.As(err, &alreadyRunning):
		c.JSON(http.StatusConflict, ReindexConflictDto{
			Error: i18n.GetMessage("ERROR_REINDEX_ALREADY_RUNNING"),
			JobID: alreadyRunning.JobID,
		})
	default:
		applog.Error("reindex failed", "error", err.Error())
		handler.completeError(logModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
	}
}

func (handler *Handler) createLog(name string, clientIP string) logger.LoggerModel {
	logModel := logger.LoggerModel{
		Name:      name,
		Level:     logger.LogLevelInfo,
		Status:    logger.LogStatusPending,
		IPAddress: clientIP,
	}
	if handler.logService == nil {
		return logModel
	}
	created, err := handler.logService.CreateLog(logModel, nil)
	if err != nil {
		return logModel
	}
	return created
}

func (handler *Handler) completeSuccess(logModel logger.LoggerModel) {
	if handler.logService == nil {
		return
	}
	_ = handler.logService.CompleteWithSuccessLog(logModel)
}

func (handler *Handler) completeError(logModel logger.LoggerModel, err error) {
	if handler.logService == nil {
		return
	}
	_ = handler.logService.CompleteWithErrorLog(logModel, err)
}
