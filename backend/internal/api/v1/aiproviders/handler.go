package aiproviders

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

func (h *Handler) GetProvidersHandler(c *gin.Context) {
	providers, err := h.service.GetProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_AI_PROVIDERS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, providers)
}

func (h *Handler) UpdateProviderHandler(c *gin.Context) {
	name := ProviderName(c.Param("name"))
	if !name.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_AI_PROVIDER_INVALID_NAME")})
		return
	}

	var request UpdateProviderDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	provider, err := h.service.UpdateProvider(name, request)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidProvider):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_AI_PROVIDER_INVALID_NAME")})
		case errors.Is(err, ErrProviderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_AI_PROVIDER_NOT_FOUND")})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_AI_PROVIDER_UPDATE")})
		}
		return
	}

	c.JSON(http.StatusOK, provider)
}
