package music

import (
	"encoding/json"
	"errors"
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) UpdateTrackMetadataHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("UpdateMusicTrackMetadata", "Updating music track metadata", c), nil)

	fileID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	var patch TrackMetadataPatch
	if err := c.ShouldBindJSON(&patch); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondInvalidMetadataBody(c, err)
		return
	}

	track, err := handler.service.UpdateTrackMetadata(fileID, patch)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondTrackMetadataError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, track)
}

func respondTrackMetadataError(c *gin.Context, err error) {
	var validationError *fieldValidationError
	if errors.As(err, &validationError) {
		respondRefusedField(c, validationError.Field, validationError.MessageKey)
		return
	}
	respondMusicError(c, err)
}

func respondInvalidMetadataBody(c *gin.Context, err error) {
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) && typeError.Field != "" {
		respondRefusedField(c, typeError.Field, "ERROR_INVALID_REQUEST")
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
}

func respondRefusedField(c *gin.Context, fieldName string, messageKey string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage(messageKey), "field": fieldName})
}
