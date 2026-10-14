package video

import (
	"errors"
	"net/http"
	"strconv"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	posterCacheControl = "public, max-age=3600"
	jpegContentType    = "image/jpeg"
	pngContentType     = "image/png"
)

func (h *Handler) GetPlaylistPosterHandler(c *gin.Context) {
	playlistID := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	posterFileID, err := h.service.FindPlaylistPosterFileID(playlistID)
	h.respondWithPoster(c, posterFileID, err)
}

func (h *Handler) GetMoviePosterHandler(c *gin.Context) {
	videoID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	posterFileID, err := h.service.FindMoviePosterFileID(videoID)
	h.respondWithPoster(c, posterFileID, err)
}

func (h *Handler) respondWithPoster(c *gin.Context, posterFileID int, lookupErr error) {
	if lookupErr != nil {
		respondVideoError(c, lookupErr)
		return
	}

	width := utils.ParseInt(c.DefaultQuery("width", strconv.Itoa(files.GridThumbnailSize)), c)
	height := utils.ParseInt(c.DefaultQuery("height", strconv.Itoa(files.GridThumbnailSize*3/2)), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	posterFile, err := h.filesService.GetFileById(posterFileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_POSTER_NOT_FOUND")})
		return
	}

	posterETag := files.ThumbnailETag(posterFile, width, height)
	c.Header("ETag", posterETag)
	c.Header("Cache-Control", posterCacheControl)
	if c.GetHeader("If-None-Match") == posterETag {
		c.Status(http.StatusNotModified)
		return
	}

	posterData, err := h.filesService.GetFileThumbnail(posterFile, width, height)
	if err != nil {
		c.Header("ETag", "")
		c.Header("Cache-Control", "no-store")
		if errors.Is(err, files.ErrFileMissingDisk) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_POSTER_NOT_FOUND")})
			return
		}
		respondVideoError(c, err)
		return
	}

	contentType := posterContentType(posterData)
	c.Data(http.StatusOK, contentType, posterData)
}

func posterContentType(posterData []byte) string {
	if http.DetectContentType(posterData) == jpegContentType {
		return jpegContentType
	}
	return pngContentType
}
