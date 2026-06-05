package analytics

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) respond(c *gin.Context, payload any, err error) {
	if err != nil {
		if errors.Is(err, ErrInvalidPeriod) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_ANALYTICS_INVALID_PERIOD")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_ANALYTICS_LOAD")})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func parseLimit(c *gin.Context, def, max int) int {
	value, err := strconv.Atoi(c.Query("limit"))
	if err != nil || value <= 0 {
		return def
	}
	if value > max {
		return max
	}
	return value
}

func (handler *Handler) GetStorageHandler(c *gin.Context) {
	result, err := handler.service.GetStorage(c.DefaultQuery("period", "7d"))
	handler.respond(c, result, err)
}

func (handler *Handler) GetTimeSeriesHandler(c *gin.Context) {
	result, err := handler.service.GetTimeSeries(c.DefaultQuery("period", "7d"))
	handler.respond(c, result, err)
}

func (handler *Handler) GetTypesHandler(c *gin.Context) {
	result, err := handler.service.GetTypes()
	handler.respond(c, result, err)
}

func (handler *Handler) GetExtensionsHandler(c *gin.Context) {
	result, err := handler.service.GetExtensions(parseLimit(c, 12, 100))
	handler.respond(c, result, err)
}

func (handler *Handler) GetRecentFilesHandler(c *gin.Context) {
	result, err := handler.service.GetRecentFiles(parseLimit(c, 50, 200))
	handler.respond(c, result, err)
}

func (handler *Handler) GetTopFoldersHandler(c *gin.Context) {
	result, err := handler.service.GetTopFolders(parseLimit(c, 20, 100))
	handler.respond(c, result, err)
}

func (handler *Handler) GetHotFoldersHandler(c *gin.Context) {
	result, err := handler.service.GetHotFolders(c.DefaultQuery("period", "7d"), parseLimit(c, 3, 50))
	handler.respond(c, result, err)
}

func (handler *Handler) GetDuplicatesHandler(c *gin.Context) {
	result, err := handler.service.GetDuplicatesSummary()
	handler.respond(c, result, err)
}

func (handler *Handler) GetDuplicateGroupsHandler(c *gin.Context) {
	result, err := handler.service.GetDuplicateGroups(parseLimit(c, 20, 100))
	handler.respond(c, result, err)
}

func (handler *Handler) GetLibraryHandler(c *gin.Context) {
	result, err := handler.service.GetLibrary()
	handler.respond(c, result, err)
}

func (handler *Handler) GetProcessingHandler(c *gin.Context) {
	result, err := handler.service.GetProcessing()
	handler.respond(c, result, err)
}

func (handler *Handler) GetHealthHandler(c *gin.Context) {
	result, err := handler.service.GetHealth()
	handler.respond(c, result, err)
}

func (handler *Handler) GetAIUsageHandler(c *gin.Context) {
	result, err := handler.service.GetAIUsage()
	handler.respond(c, result, err)
}

func (handler *Handler) GetInsightsHandler(c *gin.Context) {
	result, err := handler.service.GetInsights(c.DefaultQuery("period", "7d"))
	handler.respond(c, gin.H{"insights": result}, err)
}
