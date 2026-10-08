package files

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) GetFileLocationHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFileLocation",
		Description: "Fetching file disk location",
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

	location, err := handler.service.GetFileLocation(id)
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
	c.JSON(http.StatusOK, location)
}

func (handler *Handler) GetFileByDiskPathHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFileByDiskPath",
		Description: "Fetching file by disk path",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	diskPath := strings.TrimSpace(c.Query("path"))
	if diskPath == "" {
		missingPathErr := fmt.Errorf("%s", i18n.GetMessage("ERROR_DISK_PATH_REQUIRED"))
		handler.Logger.CompleteWithErrorLog(loggerModel, missingPathErr)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DISK_PATH_REQUIRED")})
		return
	}

	loggerModel.SetExtraData(logger.LogExtraData{Data: map[string]string{"path": diskPath}})

	file, err := handler.service.GetActiveFileByDiskPath(diskPath)
	if errors.Is(err, sql.ErrNoRows) {
		handler.respondNotFound(c, loggerModel, "ERROR_FILE_DISK_PATH_NOT_FOUND")
		return
	}
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, file.ToResponse())
}

func (handler *Handler) respondNotFound(c *gin.Context, loggerModel logger.LoggerModel, messageKey string) {
	message := i18n.GetMessage(messageKey)
	handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("%s", message))
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}
