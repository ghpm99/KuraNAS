package ollama

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

func (h *Handler) GetStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.GetStatus(c.Request.Context()))
}

func (h *Handler) ListModelsHandler(c *gin.Context) {
	models, err := h.service.ListModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach Ollama daemon"})
		return
	}
	c.JSON(http.StatusOK, models)
}

func (h *Handler) PullModelHandler(c *gin.Context) {
	var request PullModelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model name is required"})
		return
	}

	jobID, err := h.service.PullModel(request.Model)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidModelName):
			c.JSON(http.StatusBadRequest, gin.H{"error": "model name is required"})
		case errors.Is(err, ErrJobsUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "jobs subsystem is not available"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue model download"})
		}
		return
	}

	c.JSON(http.StatusAccepted, PullModelResponse{JobID: jobID})
}

func (h *Handler) DeleteModelHandler(c *gin.Context) {
	name := c.Param("name")
	if err := h.service.DeleteModel(c.Request.Context(), name); err != nil {
		switch {
		case errors.Is(err, ErrInvalidModelName):
			c.JSON(http.StatusBadRequest, gin.H{"error": "model name is required"})
		case errors.Is(err, ErrModelNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to delete model"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": name})
}
