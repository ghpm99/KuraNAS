package image

import (
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultLibraryPageSize = 60

type LibraryHandler struct {
	service    LibraryServiceInterface
	logService logger.LoggerServiceInterface
}

func NewLibraryHandler(service LibraryServiceInterface, logService logger.LoggerServiceInterface) *LibraryHandler {
	return &LibraryHandler{service: service, logService: logService}
}

func (h *LibraryHandler) startLog(c *gin.Context, name string, description string) logger.LoggerModel {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        name,
		Description: description,
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	return loggerModel
}

func (h *LibraryHandler) rejectInvalidRequest(c *gin.Context, loggerModel logger.LoggerModel, err error) {
	h.logService.CompleteWithErrorLog(loggerModel, err)
	c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage(libraryErrorMessageKey(err))})
}

func (h *LibraryHandler) failInternally(c *gin.Context, loggerModel logger.LoggerModel, err error) {
	h.logService.CompleteWithErrorLog(loggerModel, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
}

// ListLibraryImagesHandler serves GET /image/library.
func (h *LibraryHandler) ListLibraryImagesHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "ListLibraryImages", "Listing gallery images")

	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultLibraryPageSize)
	if !isPaginationValid {
		return
	}

	request, err := parseLibraryListRequest(c)
	if err != nil {
		h.rejectInvalidRequest(c, loggerModel, err)
		return
	}
	request.Page = page
	request.PageSize = pageSize

	libraryPage, err := h.service.ListLibraryImages(request)
	if err != nil {
		h.failInternally(c, loggerModel, err)
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, libraryPage)
}

// CountLibraryImagesHandler serves GET /image/library/count.
func (h *LibraryHandler) CountLibraryImagesHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "CountLibraryImages", "Counting gallery images")

	filter, err := parseLibraryFilter(c)
	if err != nil {
		h.rejectInvalidRequest(c, loggerModel, err)
		return
	}

	count, err := h.service.CountLibraryImages(filter)
	if err != nil {
		h.failInternally(c, loggerModel, err)
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, count)
}

// ListLibraryTimelineHandler serves GET /image/library/timeline.
func (h *LibraryHandler) ListLibraryTimelineHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "ListLibraryTimeline", "Listing gallery timeline")

	filter, err := parseLibraryFilter(c)
	if err != nil {
		h.rejectInvalidRequest(c, loggerModel, err)
		return
	}

	buckets, err := h.service.ListLibraryTimeline(filter)
	if err != nil {
		h.failInternally(c, loggerModel, err)
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, buckets)
}
