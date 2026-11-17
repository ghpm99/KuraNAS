package autoshutdown

import (
	"errors"
	"net/http"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   ServiceInterface
	countdown CountdownInterface
}

func NewHandler(service ServiceInterface, countdown CountdownInterface) *Handler {
	return &Handler{service: service, countdown: countdown}
}

func (h *Handler) GetSettingsHandler(c *gin.Context) {
	settings, err := h.service.GetSettings()
	if err != nil {
		applog.ErrorWithStack("autoshutdown: load settings failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_AUTO_SHUTDOWN_LOAD_FAILED")})
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
		if errors.Is(err, ErrGracePeriodOutOfRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_AUTO_SHUTDOWN_GRACE_OUT_OF_RANGE")})
			return
		}
		if errors.Is(err, ErrInvalidSettingsRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
			return
		}
		applog.ErrorWithStack("autoshutdown: update settings failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_AUTO_SHUTDOWN_UPDATE_FAILED")})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetSuggestedTimeHandler(c *gin.Context) {
	suggestion, err := h.service.SuggestedTime()
	if err != nil {
		applog.ErrorWithStack("autoshutdown: suggested time failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_AUTO_SHUTDOWN_LOAD_FAILED")})
		return
	}
	c.JSON(http.StatusOK, suggestion)
}

func (h *Handler) GetStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, h.countdown.Status())
}

func (h *Handler) CancelHandler(c *gin.Context) {
	status, err := h.countdown.CancelToday()
	if errors.Is(err, ErrNotInWarningPeriod) {
		c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("ERROR_AUTO_SHUTDOWN_NOT_IN_WARNING")})
		return
	}
	applog.Info("autoshutdown: today's shutdown cancelled", "ip", c.ClientIP())
	c.JSON(http.StatusOK, status)
}
