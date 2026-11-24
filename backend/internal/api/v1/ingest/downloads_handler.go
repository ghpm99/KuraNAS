package ingest

import (
	"errors"
	"net/http"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultDownloadsPageSize = 20

type DownloadsServiceInterface interface {
	ListDownloads(stateFilter string, page int, pageSize int) (utils.PaginationResponse[DownloadDto], error)
}

type DownloadsHandler struct {
	service DownloadsServiceInterface
}

func NewDownloadsHandler(service DownloadsServiceInterface) *DownloadsHandler {
	return &DownloadsHandler{service: service}
}

// ListDownloadsHandler serves GET /ingest/downloads: registered downloads,
// newest first, paginated, optionally filtered by state=active|finished.
func (h *DownloadsHandler) ListDownloadsHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultDownloadsPageSize)
	if !isPaginationValid {
		return
	}

	downloads, err := h.service.ListDownloads(c.Query("state"), page, pageSize)
	if errors.Is(err, ErrInvalidDownloadStateFilter) {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_STATE_FILTER")})
		return
	}
	if err != nil {
		applog.ErrorWithStack("ingest: list downloads failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_LIST")})
		return
	}

	c.JSON(http.StatusOK, downloads)
}
