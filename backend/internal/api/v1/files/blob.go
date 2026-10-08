package files

import (
	"database/sql"
	"errors"
	"mime"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
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
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]int{"id": id, "width": width, "height": height},
	})

	file, err := handler.service.GetFileById(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	thumbnailETag := ThumbnailETag(file, width, height)
	c.Header("ETag", thumbnailETag)
	c.Header("Cache-Control", "public, max-age=3600")
	if c.GetHeader("If-None-Match") == thumbnailETag {
		handler.Logger.CompleteWithSuccessLog(loggerModel)
		c.Status(http.StatusNotModified)
		return
	}

	thumbnailData, err := handler.service.GetFileThumbnail(file, width, height)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.Header("ETag", "")
		c.Header("Cache-Control", "no-store")
		if errors.Is(err, ErrFileMissingDisk) {
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	contentType := thumbnailContentType(thumbnailData)
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, thumbnailData)
}

func (handler *Handler) GetBlobFileHandler(c *gin.Context) {
	handler.serveFileById(c, "GetBlobFile", "Fetching file by ID", false)
}

func (handler *Handler) DownloadFileHandler(c *gin.Context) {
	handler.serveFileById(c, "DownloadFile", "Downloading file by ID", true)
}

func (handler *Handler) serveFileById(c *gin.Context, logName string, logDescription string, isAttachment bool) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        logName,
		Description: logDescription,
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: id,
	})

	file, err := handler.service.GetFileById(id)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		respondFileLookupError(c, err)
		return
	}

	if file.Type == Directory && isAttachment {
		handler.Logger.CompleteWithSuccessLog(loggerModel)
		handler.streamFolderAsZip(c, file)
		return
	}

	if file.Type == Directory {
		handler.Logger.CompleteWithErrorLog(loggerModel, ErrInvalidFormat)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	content, err := os.Open(file.ResolveContentPath())
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		respondContentOpenError(c, err)
		return
	}
	defer content.Close()

	contentInfo, err := content.Stat()
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.recentFileService.RegisterAccess(c.ClientIP(), file.ID)
	handler.Logger.CompleteWithSuccessLog(loggerModel)

	if contentType := mime.TypeByExtension(strings.ToLower(file.Format)); contentType != "" {
		c.Header("Content-Type", contentType)
	}
	if isAttachment {
		c.Header("Content-Disposition", buildAttachmentDisposition(file.Name))
	}
	http.ServeContent(c.Writer, c.Request, file.Name, contentInfo.ModTime(), content)
}

func buildAttachmentDisposition(fileName string) string {
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": fileName})
	if disposition == "" {
		return "attachment"
	}
	return disposition
}

func respondFileLookupError(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}
	applog.ErrorWithStack("files: lookup failed", err, "ip", c.ClientIP())
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
}

func respondContentOpenError(c *gin.Context, err error) {
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}
	applog.ErrorWithStack("files: open content failed", err, "ip", c.ClientIP())
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
}

func thumbnailContentType(thumbnailData []byte) string {
	detectedType := http.DetectContentType(thumbnailData)
	if detectedType == "image/jpeg" {
		return detectedType
	}
	return "image/png"
}
