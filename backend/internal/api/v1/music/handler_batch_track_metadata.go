package music

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) UpdateTracksMetadataHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("UpdateMusicTracksMetadata", "Updating music tracks metadata in batch", c), nil)

	isDryRun, isDryRunValid := parseDryRunQuery(c)
	if !isDryRunValid {
		respondRefusedField(c, "dry_run", "ERROR_INVALID_REQUEST")
		return
	}

	var request BatchTracksMetadataRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondInvalidMetadataBody(c, err)
		return
	}

	result, err := handler.service.UpdateTracksMetadata(request, isDryRun)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondTrackMetadataError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, result)
}

func parseDryRunQuery(c *gin.Context) (isDryRun bool, isValid bool) {
	rawDryRun, isPresent := c.GetQuery("dry_run")
	if !isPresent {
		return false, true
	}
	isDryRun, err := strconv.ParseBool(rawDryRun)
	return isDryRun, err == nil
}
