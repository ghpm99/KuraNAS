package video

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultMaxConcurrentTranscodes = 2

type TranscodeHandler struct {
	fileSource          RemuxFileSource
	codecSource         RemuxCodecSource
	runTranscode        RemuxRunner
	isTranscoderPresent func() bool
	transcodeSlots      conversionSlots
	logService          logger.LoggerServiceInterface
}

func NewTranscodeHandler(fileSource RemuxFileSource, codecSource RemuxCodecSource, runTranscode RemuxRunner, isTranscoderPresent func() bool, maxConcurrentTranscodes int, logService logger.LoggerServiceInterface) *TranscodeHandler {
	if maxConcurrentTranscodes <= 0 {
		maxConcurrentTranscodes = defaultMaxConcurrentTranscodes
	}
	return &TranscodeHandler{
		fileSource:          fileSource,
		codecSource:         codecSource,
		runTranscode:        runTranscode,
		isTranscoderPresent: isTranscoderPresent,
		transcodeSlots:      newConversionSlots(maxConcurrentTranscodes),
		logService:          logService,
	}
}

func NewFFmpegTranscodeHandler(fileSource RemuxFileSource, codecSource RemuxCodecSource, logService logger.LoggerServiceInterface) *TranscodeHandler {
	return NewTranscodeHandler(fileSource, codecSource, RunFFmpegRemux, IsFFmpegInstalled, defaultMaxConcurrentTranscodes, logService)
}

func (handler *TranscodeHandler) StreamTranscodedVideoHandler(c *gin.Context) {
	fileID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	startSeconds, startErr := strconv.ParseFloat(c.DefaultQuery("start", "0"), 64)
	requestedHeight, heightErr := strconv.Atoi(c.DefaultQuery("height", "0"))
	if startErr != nil || startSeconds < 0 || heightErr != nil || requestedHeight < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	loggerModel, _ := handler.logService.CreateLog(browseLogEntry("StreamTranscodedVideo", "Streaming transcoded video", c), nil)

	videoFile, err := handler.fileSource.GetFileById(fileID)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}
	sourcePath := videoFile.ResolveContentPath()
	if _, err := os.Stat(sourcePath); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}
	if !handler.isTranscoderPresent() {
		handler.logService.CompleteWithErrorLog(loggerModel, errors.New("ffmpeg not available"))
		c.JSON(http.StatusNotImplemented, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_TRANSCODE_UNAVAILABLE")})
		return
	}
	if !handler.transcodeSlots.tryAcquire() {
		c.Header("Retry-After", "5")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_TRANSCODE_BUSY")})
		return
	}
	defer handler.transcodeSlots.release()

	arguments := BuildTranscodeArguments(sourcePath, startSeconds, requestedHeight, handler.findSourceHeight(fileID))
	responseStream := &remuxResponseWriter{context: c}
	runErr := handler.runTranscode(c.Request.Context(), arguments, responseStream)
	if c.Request.Context().Err() != nil {
		handler.logService.CompleteWithSuccessLog(loggerModel)
		return
	}
	if runErr != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, runErr)
		if !responseStream.hasStarted {
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_OPERATION_FAILED")})
		}
		return
	}
	handler.logService.CompleteWithSuccessLog(loggerModel)
}

func (handler *TranscodeHandler) findSourceHeight(fileID int) int {
	summary, err := handler.codecSource.GetVideoSummary(fileID)
	if err != nil {
		return 0
	}
	return summary.Height
}
