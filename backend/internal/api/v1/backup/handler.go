package backup

import (
	"errors"
	"net/http"

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
		applog.ErrorWithStack("backup: load settings failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_BACKUP_SETTINGS_LOAD")})
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
		if errors.Is(err, ErrInvalidDestination) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("BACKUP_INVALID_DESTINATION")})
			return
		}
		applog.ErrorWithStack("backup: save settings failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_BACKUP_SETTINGS_SAVE")})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetStatusHandler(c *gin.Context) {
	status, err := h.service.Status()
	if err != nil {
		applog.ErrorWithStack("backup: load status failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_BACKUP_STATUS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *Handler) GetPendingHandler(c *gin.Context) {
	pending, err := h.service.Pending()
	if err != nil {
		applog.ErrorWithStack("backup: load pending failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_BACKUP_STATUS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, pending)
}
