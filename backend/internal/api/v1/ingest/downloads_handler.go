package ingest

import (
	"errors"
	"net/http"
	"strconv"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultDownloadsPageSize = 20

type DownloadsServiceInterface interface {
	ListDownloads(stateFilter string, page int, pageSize int) (utils.PaginationResponse[DownloadDto], error)
	RetryDownload(downloadID int) (DownloadDto, error)
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

// RetryDownloadHandler serves POST /ingest/downloads/:id/retry: repeats a
// failed or canceled download as a new request and answers 201 with the new item.
func (h *DownloadsHandler) RetryDownloadHandler(c *gin.Context) {
	downloadID, err := strconv.Atoi(c.Param("id"))
	if err != nil || downloadID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_ID")})
		return
	}

	download, err := h.service.RetryDownload(downloadID)
	if err != nil {
		h.respondRetryError(c, err)
		return
	}

	c.JSON(http.StatusCreated, download)
}

func (h *DownloadsHandler) respondRetryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrDownloadNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_RECORD_NOT_FOUND")})
	case errors.Is(err, ErrDownloadNotRetryable):
		c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_NOT_RETRYABLE")})
	case errors.Is(err, ErrInvalidURL):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_URL")})
	case errors.Is(err, ErrURLTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.Translate("ERROR_DOWNLOAD_URL_TOO_LONG", MaxURLLength)})
	case errors.Is(err, ErrInvalidPreset):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_PRESET")})
	case errors.Is(err, ErrInvalidTarget):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_TARGET")})
	case errors.Is(err, ErrInvalidSubfolder):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_INVALID_SUBFOLDER")})
	case errors.Is(err, ErrYtDlpNotInstalled):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_YTDLP_NOT_INSTALLED")})
	case errors.Is(err, ErrJobsUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_JOBS_UNAVAILABLE")})
	default:
		applog.ErrorWithStack("ingest: retry download failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOWNLOAD_ENQUEUE")})
	}
}
