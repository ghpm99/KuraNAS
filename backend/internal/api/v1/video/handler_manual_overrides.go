package video

import (
	"errors"
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SetVideoClassificationHandler(c *gin.Context) {
	fileID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	var request SetVideoClassificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	classification, err := h.service.SetVideoClassification(fileID, request.Classification)
	if err != nil {
		respondManualOverrideError(c, err)
		return
	}
	c.JSON(http.StatusOK, classification)
}

func (h *Handler) SetVideoEpisodeHandler(c *gin.Context) {
	fileID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	var request SetVideoEpisodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	episode, err := h.service.SetVideoEpisode(fileID, request.Season, request.Episode)
	if err != nil {
		respondManualOverrideError(c, err)
		return
	}
	c.JSON(http.StatusOK, episode)
}

func respondManualOverrideError(c *gin.Context, err error) {
	var validationError *VideoFieldValidationError
	if errors.As(err, &validationError) {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage(validationError.MessageKey), "field": validationError.Field})
		return
	}
	respondVideoError(c, err)
}
