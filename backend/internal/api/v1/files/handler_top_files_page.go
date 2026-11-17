package files

import (
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultTopFilesBySizePageSize = 20

func (handler *Handler) respondTopFilesBySizePage(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultTopFilesBySizePageSize)
	if !isPaginationValid {
		return
	}

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetTopFilesBySizePage",
		Description: "Fetching a page of the largest files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	topFilesPage, err := handler.service.GetTopFilesBySizePage(page, pageSize)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, ParsePaginationToResponse(topFilesPage))
}
