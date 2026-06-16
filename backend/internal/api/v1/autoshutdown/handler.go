package autoshutdown

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
		if errors.Is(err, ErrInvalidSettingsRequest) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_AUTO_SHUTDOWN_UPDATE_FAILED")})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetSuggestedTimeHandler(c *gin.Context) {
	suggestion, err := h.service.SuggestedTime()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_AUTO_SHUTDOWN_LOAD_FAILED")})
		return
	}
	c.JSON(http.StatusOK, suggestion)
}
