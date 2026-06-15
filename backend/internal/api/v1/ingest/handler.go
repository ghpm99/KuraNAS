package ingest

import (
	"errors"
	"net/http"

	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// FetchHandler enqueues a server-side yt-dlp download of the given URL.
func (h *Handler) FetchHandler(c *gin.Context) {
	var request FetchRequestDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_REQUEST")})
		return
	}

	jobID, err := h.service.Fetch(request)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidURL):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_URL")})
		case errors.Is(err, ErrInvalidPreset):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_PRESET")})
		case errors.Is(err, ErrInvalidTarget), errors.Is(err, ErrInvalidSubfolder):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_TARGET")})
		case errors.Is(err, ErrJobsUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_JOBS_UNAVAILABLE")})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_ENQUEUE")})
		}
		return
	}

	c.JSON(http.StatusAccepted, FetchResponseDto{JobID: jobID})
}

// GetTargetsHandler lists the enabled storage roots a download can be saved to.
func (h *Handler) GetTargetsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.ListTargets())
}

// GetPresetsHandler lists the selectable download presets.
func (h *Handler) GetPresetsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.ListPresets())
}
