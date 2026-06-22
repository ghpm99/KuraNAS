package ollama

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

func (h *Handler) GetStatusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.GetStatus(c.Request.Context()))
}

func (h *Handler) ListModelsHandler(c *gin.Context) {
	models, err := h.service.ListModels(c.Request.Context())
	if err != nil {
		applog.ErrorWithStack("ollama: list models failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusBadGateway, gin.H{"error": i18n.GetMessage("ERROR_OLLAMA_DAEMON_UNREACHABLE")})
		return
	}
	c.JSON(http.StatusOK, models)
}

func (h *Handler) PullModelHandler(c *gin.Context) {
	var request PullModelRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_OLLAMA_MODEL_NAME_REQUIRED")})
		return
	}

	jobID, err := h.service.PullModel(request.Model)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidModelName):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_OLLAMA_MODEL_NAME_REQUIRED")})
		case errors.Is(err, ErrJobsUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_OLLAMA_JOBS_UNAVAILABLE")})
		default:
			applog.ErrorWithStack("ollama: enqueue pull failed", err, "model", request.Model, "ip", c.ClientIP())
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_OLLAMA_ENQUEUE_DOWNLOAD")})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_OLLAMA_MODEL_NAME_REQUIRED")})
		case errors.Is(err, ErrModelNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_OLLAMA_MODEL_NOT_FOUND")})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"error": i18n.GetMessage("ERROR_OLLAMA_DELETE_MODEL")})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": name})
}
