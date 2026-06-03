package aiproviders

import (
	"errors"
	"net/http"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load AI providers"})
		return
	}
	c.JSON(http.StatusOK, providers)
}

func (h *Handler) UpdateProviderHandler(c *gin.Context) {
	name := ProviderName(c.Param("name"))
	if !name.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider name"})
		return
	}

	var request UpdateProviderDto
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	provider, err := h.service.UpdateProvider(name, request)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidProvider):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider name"})
		case errors.Is(err, ErrProviderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update AI provider"})
		}
		return
	}

	c.JSON(http.StatusOK, provider)
}
