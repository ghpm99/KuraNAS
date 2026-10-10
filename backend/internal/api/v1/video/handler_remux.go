package video

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultMaxConcurrentRemuxes = 4

type RemuxFileSource interface {
	GetFileById(id int) (files.FileDto, error)
}

type RemuxCodecSource interface {
	GetVideoSummary(fileID int) (VideoSummaryDto, error)
}

type RemuxHandler struct {
	fileSource       RemuxFileSource
	codecSource      RemuxCodecSource
	runRemux         RemuxRunner
	isRemuxerPresent func() bool
	remuxSlots       conversionSlots
	logService       logger.LoggerServiceInterface
}

func NewRemuxHandler(fileSource RemuxFileSource, codecSource RemuxCodecSource, runRemux RemuxRunner, isRemuxerPresent func() bool, maxConcurrentRemuxes int, logService logger.LoggerServiceInterface) *RemuxHandler {
	if maxConcurrentRemuxes <= 0 {
		maxConcurrentRemuxes = defaultMaxConcurrentRemuxes
	}
	return &RemuxHandler{
		fileSource:       fileSource,
		codecSource:      codecSource,
		runRemux:         runRemux,
		isRemuxerPresent: isRemuxerPresent,
		remuxSlots:       newConversionSlots(maxConcurrentRemuxes),
		logService:       logService,
	}
}

func NewFFmpegRemuxHandler(fileSource RemuxFileSource, codecSource RemuxCodecSource, logService logger.LoggerServiceInterface) *RemuxHandler {
	return NewRemuxHandler(fileSource, codecSource, RunFFmpegRemux, IsFFmpegInstalled, defaultMaxConcurrentRemuxes, logService)
}

func (handler *RemuxHandler) StreamRemuxedVideoHandler(c *gin.Context) {
	fileID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	startSeconds, startErr := strconv.ParseFloat(c.DefaultQuery("start", "0"), 64)
	if startErr != nil || startSeconds < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	loggerModel, _ := handler.logService.CreateLog(browseLogEntry("StreamRemuxedVideo", "Streaming remuxed video", c), nil)

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
	if !handler.hasRemuxableVideoCodec(fileID) {
		handler.logService.CompleteWithErrorLog(loggerModel, errors.New("video codec requires transcode"))
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_REQUIRES_TRANSCODE")})
		return
	}
	if !handler.isRemuxerPresent() {
		handler.logService.CompleteWithErrorLog(loggerModel, errors.New("ffmpeg not available"))
		c.JSON(http.StatusNotImplemented, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_REMUX_UNAVAILABLE")})
		return
	}
	if !handler.tryAcquireSlot() {
		c.Header("Retry-After", "5")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_REMUX_BUSY")})
		return
	}
	defer handler.releaseSlot()

	responseStream := &remuxResponseWriter{context: c}
	runErr := handler.runRemux(c.Request.Context(), BuildRemuxArguments(sourcePath, startSeconds), responseStream)
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

func (handler *RemuxHandler) hasRemuxableVideoCodec(fileID int) bool {
	summary, err := handler.codecSource.GetVideoSummary(fileID)
	if err != nil {
		return false
	}
	return IsRemuxableVideoCodec(summary.CodecName)
}

func (handler *RemuxHandler) tryAcquireSlot() bool {
	return handler.remuxSlots.tryAcquire()
}

func (handler *RemuxHandler) releaseSlot() {
	handler.remuxSlots.release()
}

type remuxResponseWriter struct {
	context    *gin.Context
	hasStarted bool
}

func (writer *remuxResponseWriter) Write(chunk []byte) (int, error) {
	if !writer.hasStarted {
		writer.hasStarted = true
		writer.context.Header("Content-Type", remuxContentType)
		writer.context.Header("Accept-Ranges", "none")
		writer.context.Header("Cache-Control", "no-store")
		writer.context.Status(http.StatusOK)
	}
	written, err := writer.context.Writer.Write(chunk)
	writer.context.Writer.Flush()
	return written, err
}
