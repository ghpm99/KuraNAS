package files

import (
	"errors"
	"github.com/gin-gonic/gin"
	"mime"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"
	"strings"
)

func (handler *Handler) GetFileThumbnailHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFileThumbnail",
		Description: "Fetching file thumbnail by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)
	width := utils.ParseInt(c.DefaultQuery("width", "320"), c)
	height := utils.ParseInt(c.DefaultQuery("height", "320"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]int{"id": id, "width": width, "height": height},
	})

	file, err := handler.service.GetFileById(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	thumbnailData, err := handler.service.GetFileThumbnail(file, width, height)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, ErrFileMissingDisk) {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/png", thumbnailData)
}

func (handler *Handler) GetBlobFileHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetBlobFile",
		Description: "Fetching file by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: id,
	})

	fileBlob, err := handler.service.GetFileBlobById(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.recentFileService.RegisterAccess(c.ClientIP(), fileBlob.ID, config.AppConfig.RecentFilesKeep)
	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Data(http.StatusOK, mime.TypeByExtension(strings.ToLower(fileBlob.Format)), fileBlob.Blob)
}
