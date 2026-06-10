package image

import (
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
	groupBy, err := files.ParseImageGroupBy(c.DefaultQuery("group_by", string(files.ImageGroupByDate)))
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
