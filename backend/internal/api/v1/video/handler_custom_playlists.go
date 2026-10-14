package video

import (
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateCustomPlaylistHandler(c *gin.Context) {
	var request CreateCustomPlaylistRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	createdPlaylist, err := h.service.CreateCustomPlaylist(request)
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusCreated, createdPlaylist)
}

func (h *Handler) DeleteCustomPlaylistHandler(c *gin.Context) {
	playlistID := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	if err := h.service.DeleteCustomPlaylist(playlistID); err != nil {
		respondVideoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
