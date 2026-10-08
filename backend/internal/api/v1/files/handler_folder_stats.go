package files

import (
	"database/sql"
	"errors"
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) GetFolderStatsHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFolderStats",
		Description: "Computing recursive folder stats",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	loggerModel.SetExtraData(logger.LogExtraData{Data: map[string]int{"id": id}})

	stats, err := handler.service.GetFolderStats(id)
	if errors.Is(err, sql.ErrNoRows) {
		handler.respondNotFound(c, loggerModel, "ERROR_FILE_NOT_FOUND")
		return
	}
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, stats)
}
