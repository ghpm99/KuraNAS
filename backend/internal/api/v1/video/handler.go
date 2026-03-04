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

	session, err := h.service.StartPlayback(c.ClientIP(), req.VideoID, req.PlaylistID)
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

func (h *Handler) RebuildPlaylistsHandler(c *gin.Context) {
	if err := h.service.RebuildSmartPlaylists(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetPlaylistsHandler(c *gin.Context) {
	includeHidden := c.DefaultQuery("include_hidden", "false") == "true"
	playlists, err := h.service.GetPlaylists(includeHidden)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, playlists)
}

func (h *Handler) GetPlaylistByIDHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	playlist, err := h.service.GetPlaylistByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, playlist)
}

func (h *Handler) SetPlaylistHiddenHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	var req SetPlaylistHiddenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.SetPlaylistHidden(id, req.Hidden); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) AddPlaylistVideoHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	var req AddPlaylistVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.AddVideoToPlaylist(id, req.VideoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func (h *Handler) RemovePlaylistVideoHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	videoID := utils.ParseInt(c.Param("videoId"), c)
	if err := h.service.RemoveVideoFromPlaylist(id, videoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) UpdatePlaylistHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	var req UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdatePlaylistName(id, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) ReorderPlaylistHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	var req ReorderPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.ReorderPlaylistItems(id, req.Items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetUnassignedVideosHandler(c *gin.Context) {
	limit := utils.ParseInt(c.DefaultQuery("limit", "2000"), c)
	videos, err := h.service.GetUnassignedVideos(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, videos)
}
