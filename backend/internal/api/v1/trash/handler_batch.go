package trash

import (
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type BatchRequest struct {
	IDs []int `json:"ids" binding:"required"`
}

// BatchFailureDto reports why one item of a batch was left untouched. Path is
// only present when a restore was blocked by an occupied location.
type BatchFailureDto struct {
	ID    int    `json:"id"`
	Error string `json:"error"`
	Path  string `json:"path,omitempty"`
}

type BatchRestoreResponse struct {
	RestoredIDs []int             `json:"restored_ids"`
	Failures    []BatchFailureDto `json:"failures"`
}

type BatchDeleteResponse struct {
	DeletedIDs []int             `json:"deleted_ids"`
	Failures   []BatchFailureDto `json:"failures"`
}

func toBatchFailureDtos(failures []ItemFailure) []BatchFailureDto {
	failureDtos := make([]BatchFailureDto, 0, len(failures))
	for _, failure := range failures {
		_, body := serviceErrorResponse(failure.Err)
		failureDto := BatchFailureDto{ID: failure.ID}
		failureDto.Error, _ = body["error"].(string)
		failureDto.Path, _ = body["path"].(string)
		failureDtos = append(failureDtos, failureDto)
	}
	return failureDtos
}

func (handler *Handler) bindBatchIDs(c *gin.Context, logModel logger.LoggerModel) ([]int, bool) {
	var request BatchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(logModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return nil, false
	}
	if len(request.IDs) == 0 || len(request.IDs) > MaxBatchItems {
		handler.completeError(logModel, ErrInvalidBatchSize)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return nil, false
	}
	return request.IDs, true
}

func (handler *Handler) RestoreTrashItemsHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "RestoreTrashItems",
		Description: "Restoring several items from trash",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	ids, isRequestValid := handler.bindBatchIDs(c, logModel)
	if !isRequestValid {
		return
	}

	restoredIDs, failures := handler.service.RestoreItems(ids)

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, BatchRestoreResponse{RestoredIDs: restoredIDs, Failures: toBatchFailureDtos(failures)})
}

func (handler *Handler) DeleteTrashItemsHandler(c *gin.Context) {
	logModel := handler.createLog(logger.LoggerModel{
		Name:        "DeleteTrashItems",
		Description: "Permanently deleting several trash items",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	ids, isRequestValid := handler.bindBatchIDs(c, logModel)
	if !isRequestValid {
		return
	}

	deletedIDs, failures := handler.service.DeleteItemsPermanently(ids)

	handler.completeSuccess(logModel)
	c.JSON(http.StatusOK, BatchDeleteResponse{DeletedIDs: deletedIDs, Failures: toBatchFailureDtos(failures)})
}
