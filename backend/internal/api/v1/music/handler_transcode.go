package music

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultMaxConcurrentTranscodes = 4

type TranscodeTrackSource interface {
	GetFileById(id int) (files.FileDto, error)
}

type TranscodeHandler struct {
	trackSource         TranscodeTrackSource
	runTranscode        TranscodeRunner
	isTranscoderPresent func() bool
	transcodeSlots      chan struct{}
	logService          logger.LoggerServiceInterface
}

func NewTranscodeHandler(trackSource TranscodeTrackSource, runTranscode TranscodeRunner, isTranscoderPresent func() bool, maxConcurrentTranscodes int, logService logger.LoggerServiceInterface) *TranscodeHandler {
	if maxConcurrentTranscodes <= 0 {
		maxConcurrentTranscodes = defaultMaxConcurrentTranscodes
	}
	return &TranscodeHandler{
		trackSource:         trackSource,
		runTranscode:        runTranscode,
		isTranscoderPresent: isTranscoderPresent,
		transcodeSlots:      make(chan struct{}, maxConcurrentTranscodes),
		logService:          logService,
	}
}

func NewFFmpegTranscodeHandler(trackSource TranscodeTrackSource, logService logger.LoggerServiceInterface) *TranscodeHandler {
	return NewTranscodeHandler(trackSource, RunFFmpegTranscode, IsFFmpegInstalled, defaultMaxConcurrentTranscodes, logService)
}

func (handler *TranscodeHandler) StreamTranscodedTrackHandler(c *gin.Context) {
	fileID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	transcodeFormat, err := ResolveTranscodeFormat(c.DefaultQuery("format", "mp3"))
	startSeconds, startErr := strconv.ParseFloat(c.DefaultQuery("start", "0"), 64)
	if err != nil || startErr != nil || startSeconds < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	playback := openPlaybackLog(handler.logService, logEntry("StreamTranscodedTrack", "Streaming transcoded track", c), startSeconds == 0)

	trackFile, err := handler.trackSource.GetFileById(fileID)
	if err != nil {
		playback.fail(err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}
	sourcePath := trackFile.ResolveContentPath()
	if _, err := os.Stat(sourcePath); err != nil {
		playback.fail(err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}
	if !handler.isTranscoderPresent() {
		playback.fail(errors.New("ffmpeg not available"))
		c.JSON(http.StatusNotImplemented, gin.H{"error": i18n.GetMessage("ERROR_MUSIC_TRANSCODE_UNAVAILABLE")})
		return
	}
	if !handler.tryAcquireSlot() {
		c.Header("Retry-After", "5")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_MUSIC_TRANSCODE_BUSY")})
		return
	}
	defer handler.releaseSlot()

	responseStream := &lazyHeaderWriter{context: c, contentType: transcodeFormat.ContentType}
	arguments := BuildTranscodeArguments(sourcePath, transcodeFormat, startSeconds)
	runErr := handler.runTranscode(c.Request.Context(), arguments, responseStream)
	if c.Request.Context().Err() != nil {
		playback.succeed()
		return
	}
	if runErr != nil {
		playback.fail(runErr)
		if !responseStream.hasStarted {
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_MUSIC_OPERATION_FAILED")})
		}
		return
	}
	playback.succeed()
}

func (handler *TranscodeHandler) tryAcquireSlot() bool {
	select {
	case handler.transcodeSlots <- struct{}{}:
		return true
	default:
		return false
	}
}

func (handler *TranscodeHandler) releaseSlot() {
	<-handler.transcodeSlots
}

type lazyHeaderWriter struct {
	context     *gin.Context
	contentType string
	hasStarted  bool
}

func (writer *lazyHeaderWriter) Write(chunk []byte) (int, error) {
	if !writer.hasStarted {
		writer.hasStarted = true
		writer.context.Header("Content-Type", writer.contentType)
		writer.context.Header("Accept-Ranges", "none")
		writer.context.Header("Cache-Control", "no-store")
		writer.context.Status(http.StatusOK)
	}
	written, err := writer.context.Writer.Write(chunk)
	writer.context.Writer.Flush()
	return written, err
}
