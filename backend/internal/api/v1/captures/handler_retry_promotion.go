package captures

import (
	"errors"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RetryCapturePromotionResponseDto struct {
	JobID int `json:"job_id"`
}

func (h *Handler) RetryCapturePromotionHandler(c *gin.Context) {
	captureID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_ID")})
		return
	}

	jobID, err := h.service.RetryCapturePromotion(captureID)
	if errors.Is(err, ErrCaptureNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_CAPTURE_NOT_FOUND")})
		return
	}
	if errors.Is(err, ErrCaptureNotRetryable) {
		c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("ERROR_CAPTURE_PROMOTE_NOT_RETRYABLE")})
		return
	}
	if err != nil {
		applog.Error("captures: retry promotion failed", "capture_id", captureID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CAPTURE_PROMOTE_RETRY_FAILED")})
		return
	}

	c.JSON(http.StatusAccepted, RetryCapturePromotionResponseDto{JobID: jobID})
}
