package video

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Handlers moved from the files core. Paths are unchanged (/files/videos,
// /files/video-stream/:id, /files/video-thumbnail/:id, /files/video-preview/:id);
// only the owning package changed.

func browseLogEntry(name, description string, c *gin.Context) logger.LoggerModel {
	return logger.LoggerModel{
		Name:        name,
		Description: description,
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}
}

func (h *Handler) GetVideosHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(browseLogEntry("GetVideos", "Fetching video files", c), nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	pagination, err := h.service.GetVideos(page, pageSize)

	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, files.ParsePaginationToResponse(pagination))
}

func (h *Handler) GetVideoThumbnailHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(browseLogEntry("GetVideoThumbnail", "Fetching video thumbnail by ID", c), nil)

	id := utils.ParseInt(c.Param("id"), c)
	width := utils.ParseInt(c.DefaultQuery("width", "320"), c)
	height := utils.ParseInt(c.DefaultQuery("height", "180"), c)

	file, err := h.filesService.GetFileById(id)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	thumbnailData, err := h.service.GetVideoThumbnail(file, width, height)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, files.ErrFileMissingDisk) {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/png", thumbnailData)
}

func (h *Handler) GetVideoPreviewHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(browseLogEntry("GetVideoPreview", "Fetching animated video preview by ID", c), nil)

	id := utils.ParseInt(c.Param("id"), c)
	width := utils.ParseInt(c.DefaultQuery("width", "320"), c)
	height := utils.ParseInt(c.DefaultQuery("height", "180"), c)

	file, err := h.filesService.GetFileById(id)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	previewData, err := h.service.GetVideoPreviewGif(file, width, height)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, files.ErrFileMissingDisk) {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.Header("Content-Type", "image/gif")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/gif", previewData)
}

func (h *Handler) StreamVideoHandler(c *gin.Context) {
	loggerModel, _ := h.logService.CreateLog(browseLogEntry("StreamVideo", "Streaming video file", c), nil)

	id := utils.ParseInt(c.Param("id"), c)

	file, err := h.filesService.GetFileById(id)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	contentPath := file.ResolveContentPath()
	exists := h.filesService.CheckFileExistsByPath(contentPath)
	if !exists {
		h.logService.CompleteWithErrorLog(loggerModel, fmt.Errorf("file not found on disk"))
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	h.recentFileService.RegisterAccess(c.ClientIP(), file.ID)

	videoFile, err := os.Open(contentPath)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}
	defer videoFile.Close()

	fileInfo, err := videoFile.Stat()
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	contentType := utils.ContentTypeByFormat(file.Format, "video/mp4")
	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=3600")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		start, end, ok := utils.ParseHTTPRange(rangeHeader, fileInfo.Size())
		if ok {
			c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
			c.Header("Content-Length", fmt.Sprintf("%d", end-start+1))
			c.Status(http.StatusPartialContent)

			videoFile.Seek(start, 0)
			_, err := io.CopyN(c.Writer, videoFile, end-start+1)
			if err != nil {
				h.logService.CompleteWithErrorLog(loggerModel, err)
				return
			}

			h.logService.CompleteWithSuccessLog(loggerModel)
			return
		}
	}

	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, videoFile)
	if err != nil {
		h.logService.CompleteWithErrorLog(loggerModel, err)
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
}
