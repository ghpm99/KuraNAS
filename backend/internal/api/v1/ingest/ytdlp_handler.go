package ingest

import (
	"net/http"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type YtDlpHandler struct {
	service YtDlpServiceInterface
}

func NewYtDlpHandler(service YtDlpServiceInterface) *YtDlpHandler {
	return &YtDlpHandler{service: service}
}

// GetStatusHandler reports the installed yt-dlp version vs the latest release.
func (h *YtDlpHandler) GetStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.Status())
}

// UpdateHandler applies a verified yt-dlp update. It is the human-triggered side
// of the lifecycle: nothing here runs on a timer.
func (h *YtDlpHandler) UpdateHandler(c *gin.Context) {
	if err := h.service.Update(); err != nil {
		applog.Error("ytdlp: update failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_YTDLP_UPDATE")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("YTDLP_UPDATE_APPLIED")})
}
