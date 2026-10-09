package music

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CoverServiceInterface interface {
	GetTrackCover(fileID int, size int) (Cover, error)
	GetAlbumCover(albumKey string, size int) (Cover, error)
}

type CoverHandler struct {
	service    CoverServiceInterface
	logService logger.LoggerServiceInterface
}

func NewCoverHandler(service CoverServiceInterface, logService logger.LoggerServiceInterface) *CoverHandler {
	return &CoverHandler{service: service, logService: logService}
}

func (handler *CoverHandler) GetTrackCoverHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetTrackCover", "Fetching track cover", c), nil)

	fileID := utils.ParseInt(c.Param("file_id"), c)
	size := utils.ParseInt(c.DefaultQuery("size", "0"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	cover, err := handler.service.GetTrackCover(fileID, size)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondCoverError(c, err)
		return
	}
	handler.logService.CompleteWithSuccessLog(loggerModel)
	respondCover(c, cover)
}

func (handler *CoverHandler) GetAlbumCoverHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetAlbumCover", "Fetching album cover", c), nil)

	size := utils.ParseInt(c.DefaultQuery("size", "0"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	cover, err := handler.service.GetAlbumCover(c.Param("key"), size)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondCoverError(c, err)
		return
	}
	handler.logService.CompleteWithSuccessLog(loggerModel)
	respondCover(c, cover)
}

func respondCover(c *gin.Context, cover Cover) {
	c.Header("ETag", cover.ETag)
	c.Header("Cache-Control", "public, max-age=3600")
	if c.GetHeader("If-None-Match") == cover.ETag {
		c.Status(http.StatusNotModified)
		return
	}
	c.Data(http.StatusOK, "image/jpeg", cover.Data)
}

func respondCoverError(c *gin.Context, err error) {
	if errors.Is(err, ErrCoverNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_MUSIC_COVER_NOT_FOUND")})
		return
	}
	respondMusicError(c, err)
}
