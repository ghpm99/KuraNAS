package files

import (
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

const MaxBatchStarredIDs = 500

type SetFilesStarredRequest struct {
	IDs     []int `json:"ids" binding:"required"`
	Starred *bool `json:"starred" binding:"required"`
}

type SetFilesStarredResponse struct {
	Starred     bool  `json:"starred"`
	UpdatedIDs  []int `json:"updated_ids"`
	NotFoundIDs []int `json:"not_found_ids"`
}

func (handler *Handler) SetFilesStarredHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "SetFilesStarred",
		Description: "Setting the starred state of several files by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var request SetFilesStarredRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if len(request.IDs) == 0 || len(request.IDs) > MaxBatchStarredIDs {
		handler.Logger.CompleteWithErrorLog(loggerModel, ErrInvalidBatchSize)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	requestedIDs := removeDuplicateIDs(request.IDs)
	loggerModel.SetExtraData(logger.LogExtraData{Data: requestedIDs})

	updatedIDs, err := handler.service.SetFilesStarred(requestedIDs, *request.Starred)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, SetFilesStarredResponse{
		Starred:     *request.Starred,
		UpdatedIDs:  nonNilIDs(updatedIDs),
		NotFoundIDs: idsMissingFrom(requestedIDs, updatedIDs),
	})
}

func removeDuplicateIDs(ids []int) []int {
	seenIDs := make(map[int]struct{}, len(ids))
	uniqueIDs := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, isDuplicate := seenIDs[id]; isDuplicate {
			continue
		}
		seenIDs[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	return uniqueIDs
}

func idsMissingFrom(requestedIDs []int, foundIDs []int) []int {
	foundIDSet := make(map[int]struct{}, len(foundIDs))
	for _, id := range foundIDs {
		foundIDSet[id] = struct{}{}
	}
	missingIDs := []int{}
	for _, id := range requestedIDs {
		if _, isFound := foundIDSet[id]; !isFound {
			missingIDs = append(missingIDs, id)
		}
	}
	return missingIDs
}

func nonNilIDs(ids []int) []int {
	if ids == nil {
		return []int{}
	}
	return ids
}
