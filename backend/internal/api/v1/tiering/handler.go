package tiering

import (
	"errors"
	"net/http"
	"strconv"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSettingsHandler(c *gin.Context) {
	settings, err := h.service.GetSettings()
	if err != nil {
		applog.ErrorWithStack("tiering: load settings failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_TIERING_SETTINGS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdateSettingsHandler(c *gin.Context) {
	var request SettingsDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	settings, err := h.service.UpdateSettings(request)
	if err != nil {
		if errors.Is(err, ErrInvalidColdDir) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("TIERING_INVALID_COLD_DIR")})
			return
		}
		applog.ErrorWithStack("tiering: save settings failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_TIERING_SETTINGS_SAVE")})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetStatusHandler(c *gin.Context) {
	status, err := h.service.Status()
	if err != nil {
		applog.ErrorWithStack("tiering: load status failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_TIERING_STATUS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *Handler) GetUsageHandler(c *gin.Context) {
	usage, err := h.service.Usage()
	if err != nil {
		applog.ErrorWithStack("tiering: load usage failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_TIERING_STATUS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, usage)
}

func (h *Handler) PromoteFileHandler(c *gin.Context) {
	fileID, err := strconv.Atoi(c.Param("file_id"))
	if err != nil || fileID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	location, err := h.service.PromoteFile(fileID)
	if err != nil {
		h.respondPromoteError(c, fileID, err)
		return
	}
	c.JSON(http.StatusOK, location)
}

func (h *Handler) respondPromoteError(c *gin.Context, fileID int, err error) {
	switch {
	case errors.Is(err, ErrFileNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
	case errors.Is(err, ErrFileAlreadyHot):
		c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("ERROR_TIERING_FILE_ALREADY_HOT")})
	case errors.Is(err, ErrInsufficientHotSpace):
		applog.Warn("tiering: promote refused, hot disk lacks space", "file_id", fileID, "error", err)
		c.JSON(http.StatusInsufficientStorage, gin.H{"error": i18n.GetMessage("ERROR_TIERING_NO_HOT_SPACE")})
	case errors.Is(err, ErrColdCopyUnavailable):
		applog.ErrorWithStack("tiering: promote failed, cold copy unavailable", err, "file_id", fileID, "ip", c.ClientIP())
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_TIERING_COLD_UNAVAILABLE")})
	default:
		applog.ErrorWithStack("tiering: promote failed", err, "file_id", fileID, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_TIERING_PROMOTE")})
	}
}
