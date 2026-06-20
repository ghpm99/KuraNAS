package image

import (
	"errors"
	"net/http"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Handler serves HTTP requests for the image domain.
type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(service ServiceInterface, loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{
		service:    service,
		logService: loggerService,
	}
}

// GetImagesHandler serves GET /files/images — returns a paginated, optionally
// grouped list of images with metadata. The URL path stays under /files/ so the
// HTTP contract is unchanged; only the serving package changed.
func (h *Handler) GetImagesHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        "GetImages",
		Description: "Fetching image files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)
	groupBy, err := ParseImageGroupBy(c.DefaultQuery("group_by", string(ImageGroupByDate)))
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	pagination, err := h.service.GetImages(page, pageSize, groupBy)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, files.ParsePaginationToResponse(pagination))
}

// GetPendingAIClassificationCountHandler serves
// GET /files/images/classification/pending-count — the number of indexed images
// still awaiting AI classification, used by the UI to show/enable the backfill.
func (h *Handler) GetPendingAIClassificationCountHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        "GetPendingAIClassificationCount",
		Description: "Counting images pending AI classification",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	count, err := h.service.GetPendingAIClassificationCount()
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"pending_count": count})
}

// EnqueueClassificationBackfillHandler serves
// POST /files/images/classification/backfill — enqueues a background job that
// reclassifies the images still awaiting AI classification.
func (h *Handler) EnqueueClassificationBackfillHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        "EnqueueClassificationBackfill",
		Description: "Enqueuing AI classification backfill",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	jobID, err := h.service.EnqueueClassificationBackfill()
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		if errors.Is(err, ErrBackfillUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("IMAGE_CLASSIFY_BACKFILL_UNAVAILABLE")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"message": i18n.GetMessage("IMAGE_CLASSIFY_BACKFILL_ENQUEUED"),
	})
}
