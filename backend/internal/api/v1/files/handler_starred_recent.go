package files

import (
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) GetStarredFilesHandler(c *gin.Context) {
	handler.serveFilesPage(c, "GetStarredFiles", "Fetching starred files", handler.service.GetStarredFiles)
}

func (handler *Handler) GetRecentlyAccessedFilesHandler(c *gin.Context) {
	handler.serveFilesPage(c, "GetRecentlyAccessedFiles", "Fetching recently accessed files", handler.service.GetRecentlyAccessedFiles)
}

func (handler *Handler) serveFilesPage(
	c *gin.Context,
	logName string,
	logDescription string,
	fetchPage func(page int, pageSize int) (utils.PaginationResponse[FileDto], error),
) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        logName,
		Description: logDescription,
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page, pageSize, isPaginationValid := utils.ParsePagination(c, 15)
	if !isPaginationValid {
		return
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]int{"page": page, "page_size": pageSize},
	})

	pagination, err := fetchPage(page, pageSize)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, ParsePaginationToResponse(pagination))
}
