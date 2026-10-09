package music

import (
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) GetLibraryQueueByArtistHandler(c *gin.Context) {
	artistKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicQueueByArtist", "Fetching music queue by artist", func() (any, error) {
		return handler.service.GetLibraryQueueByArtist(artistKey)
	})
}

func (handler *Handler) GetLibraryQueueByAlbumHandler(c *gin.Context) {
	albumKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicQueueByAlbum", "Fetching music queue by album", func() (any, error) {
		return handler.service.GetLibraryQueueByAlbum(albumKey)
	})
}

func (handler *Handler) GetLibraryQueueByGenreHandler(c *gin.Context) {
	genreKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicQueueByGenre", "Fetching music queue by genre", func() (any, error) {
		return handler.service.GetLibraryQueueByGenre(genreKey)
	})
}

func (handler *Handler) GetLibraryQueueByFolderHandler(c *gin.Context) {
	folderKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicQueueByFolder", "Fetching music queue by folder", func() (any, error) {
		return handler.service.GetLibraryQueueByFolder(folderKey)
	})
}

func (handler *Handler) GetPlaylistQueueHandler(c *gin.Context) {
	playlistID := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	clientID := c.ClientIP()
	handler.respondLibraryTracks(c, "GetPlaylistQueue", "Fetching playlist queue", func() (any, error) {
		return handler.service.GetPlaylistQueue(clientID, playlistID)
	})
}
