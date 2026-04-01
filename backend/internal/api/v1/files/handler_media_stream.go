package files

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

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

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]int{"id": id, "width": width, "height": height},
	})

	file, err := handler.service.GetFileById(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	thumbnailData, err := handler.service.GetFileThumbnail(file, width, height)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, ErrFileMissingDisk) {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/png", thumbnailData)
}

func (handler *Handler) GetVideoThumbnailHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetVideoThumbnail",
		Description: "Fetching video thumbnail by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)
	width := utils.ParseInt(c.DefaultQuery("width", "320"), c)
	height := utils.ParseInt(c.DefaultQuery("height", "180"), c)

	file, err := handler.service.GetFileById(id)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	thumbnailData, err := handler.service.GetVideoThumbnail(file, width, height)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, ErrFileMissingDisk) {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/png", thumbnailData)
}

func (handler *Handler) GetVideoPreviewHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetVideoPreview",
		Description: "Fetching animated video preview by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)
	width := utils.ParseInt(c.DefaultQuery("width", "320"), c)
	height := utils.ParseInt(c.DefaultQuery("height", "180"), c)

	file, err := handler.service.GetFileById(id)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	previewData, err := handler.service.GetVideoPreviewGif(file, width, height)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, ErrFileMissingDisk) {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Header("Content-Type", "image/gif")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/gif", previewData)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.recentFileService.RegisterAccess(c.ClientIP(), fileBlob.ID, config.AppConfig.RecentFilesKeep)
	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Data(http.StatusOK, mime.TypeByExtension(strings.ToLower(fileBlob.Format)), fileBlob.Blob)
}

// StreamAudioHandler streams audio files with HTTP Range support
func (handler *Handler) StreamAudioHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "StreamAudio",
		Description: "Streaming audio file",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	file, err := handler.service.GetFileById(id)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	exists := handler.service.CheckFileExistsByPath(file.Path)
	if !exists {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("file not found on disk"))
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	audioFile, err := os.Open(file.Path)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer audioFile.Close()

	fileInfo, err := audioFile.Stat()
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	contentType := contentTypeByFormat(file.Format, "audio/mpeg")
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

			audioFile.Seek(start, 0)
			_, err := io.CopyN(c.Writer, audioFile, end-start+1)
			if err != nil {
				handler.Logger.CompleteWithErrorLog(loggerModel, err)
				return
			}

			handler.Logger.CompleteWithSuccessLog(loggerModel)
			return
		}
	}

	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, audioFile)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
}

func (handler *Handler) GetVideosHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetVideos",
		Description: "Fetching video files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	pagination, err := handler.service.GetVideos(page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, ParsePaginationToResponse(pagination))
}

func (handler *Handler) StreamVideoHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "StreamVideo",
		Description: "Streaming video file",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	file, err := handler.service.GetFileById(id)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	exists := handler.service.CheckFileExistsByPath(file.Path)
	if !exists {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("file not found on disk"))
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	videoFile, err := os.Open(file.Path)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer videoFile.Close()

	fileInfo, err := videoFile.Stat()
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
				handler.Logger.CompleteWithErrorLog(loggerModel, err)
				return
			}

			handler.Logger.CompleteWithSuccessLog(loggerModel)
			return
		}
	}

	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, videoFile)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
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
