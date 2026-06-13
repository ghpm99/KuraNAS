package tiering

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

func (h *Handler) GetSettingsHandler(c *gin.Context) {
	settings, err := h.service.GetSettings()
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_TIERING_SETTINGS_SAVE")})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetStatusHandler(c *gin.Context) {
	status, err := h.service.Status()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_TIERING_STATUS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *Handler) GetUsageHandler(c *gin.Context) {
	usage, err := h.service.Usage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_TIERING_STATUS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, usage)
}
