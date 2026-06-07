package distribution

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

// GetDownloadsHandler returns the catalog of distributable client apps. It owns
// one piece of information: the list of available downloads with their metadata.
func (handler *Handler) GetDownloadsHandler(c *gin.Context) {
	items, err := handler.service.ListDownloads()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOADS_LIST")})
		return
	}
	c.JSON(http.StatusOK, items)
}

// DownloadFileHandler streams a single artifact as an attachment. The filename
// comes from the manifest (never from the URL), so the id param cannot be used
// for path traversal.
func (handler *Handler) DownloadFileHandler(c *gin.Context) {
	id := c.Param("id")

	path, filename, err := handler.service.ResolveDownload(id)
	if err != nil {
		if errors.Is(err, ErrArtifactNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_NOT_FOUND")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOADS_LIST")})
		return
	}

	c.FileAttachment(path, filename)
}
