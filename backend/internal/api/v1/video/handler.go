package video

import (
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(service ServiceInterface, logService logger.LoggerServiceInterface) *Handler {
	return &Handler{service: service, logService: logService}
}

func (h *Handler) StartPlaybackHandler(c *gin.Context) {
	var req StartPlaybackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.service.StartPlayback(c.ClientIP(), req.VideoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *Handler) GetPlaybackStateHandler(c *gin.Context) {
	session, err := h.service.GetPlaybackState(c.ClientIP())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *Handler) UpdatePlaybackStateHandler(c *gin.Context) {
	var req UpdatePlaybackStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state, err := h.service.UpdatePlaybackState(c.ClientIP(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *Handler) NextVideoHandler(c *gin.Context) {
	session, err := h.service.NextVideo(c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *Handler) PreviousVideoHandler(c *gin.Context) {
	session, err := h.service.PreviousVideo(c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *Handler) GetHomeCatalogHandler(c *gin.Context) {
	limit := utils.ParseInt(c.DefaultQuery("limit", "24"), c)

	catalog, err := h.service.GetHomeCatalog(c.ClientIP(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalog)
}
