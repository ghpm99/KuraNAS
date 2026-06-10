package video

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	exists := h.filesService.CheckFileExistsByPath(file.Path)
	if !exists {
		h.logService.CompleteWithErrorLog(loggerModel, fmt.Errorf("file not found on disk"))
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	videoFile, err := os.Open(file.Path)
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

	contentType := contentTypeByFormat(file.Format, "video/mp4")
	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=3600")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		start, end, ok := parseHTTPRange(rangeHeader, fileInfo.Size())
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

func parseHTTPRange(rangeHeader string, fileSize int64) (int64, int64, bool) {
	if fileSize <= 0 {
		return 0, 0, false
	}

	parts := strings.SplitN(strings.TrimSpace(rangeHeader), "=", 2)
	if len(parts) != 2 || parts[0] != "bytes" {
		return 0, 0, false
	}

	rangeValue := strings.TrimSpace(parts[1])
	if rangeValue == "" {
		return 0, 0, false
	}

	// Only first range is supported.
	if commaIndex := strings.Index(rangeValue, ","); commaIndex >= 0 {
		rangeValue = strings.TrimSpace(rangeValue[:commaIndex])
	}

	bounds := strings.SplitN(rangeValue, "-", 2)
	if len(bounds) != 2 {
		return 0, 0, false
	}

	startText := strings.TrimSpace(bounds[0])
	endText := strings.TrimSpace(bounds[1])

	var start int64
	var end int64
	var err error

	if startText == "" {
		// Suffix byte range: bytes=-500
		suffixLength, parseErr := strconv.ParseInt(endText, 10, 64)
		if parseErr != nil || suffixLength <= 0 {
			return 0, 0, false
		}
		if suffixLength > fileSize {
			suffixLength = fileSize
		}
		start = fileSize - suffixLength
		end = fileSize - 1
	} else {
		start, err = strconv.ParseInt(startText, 10, 64)
		if err != nil || start < 0 || start >= fileSize {
			return 0, 0, false
		}

		if endText == "" {
			// Open ended range: bytes=500-
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(endText, 10, 64)
			if err != nil {
				return 0, 0, false
			}
			if end >= fileSize {
				end = fileSize - 1
			}
		}
	}

	if end < start {
		return 0, 0, false
	}

	return start, end, true
}

func contentTypeByFormat(format string, fallback string) string {
	ext := strings.TrimSpace(format)
	if ext == "" {
		return fallback
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	contentType := mime.TypeByExtension(strings.ToLower(ext))
	if contentType == "" {
		return fallback
	}
	return contentType
}
