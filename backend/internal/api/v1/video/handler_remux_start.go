package video

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"os/exec"
	"strconv"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type RemuxStartFileSource interface {
	GetFileById(id int) (files.FileDto, error)
}

type RemuxStartHandler struct {
	fileSource  RemuxStartFileSource
	probeRunner SubtitleProbeRunner
}

func NewRemuxStartHandler(fileSource RemuxStartFileSource, probeRunner SubtitleProbeRunner) *RemuxStartHandler {
	return &RemuxStartHandler{fileSource: fileSource, probeRunner: probeRunner}
}

func NewFFprobeRemuxStartHandler(fileSource RemuxStartFileSource) *RemuxStartHandler {
	return NewRemuxStartHandler(fileSource, RunFFprobeSubtitleProbe)
}

func (handler *RemuxStartHandler) GetRemuxStartHandler(c *gin.Context) {
	fileID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	atSeconds, err := strconv.ParseFloat(c.Query("at"), 64)
	if err != nil || atSeconds < 0 || math.IsInf(atSeconds, 0) || math.IsNaN(atSeconds) {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	videoFile, err := handler.fileSource.GetFileById(fileID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && videoFile.DeletedAt.HasValue) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_NOT_FOUND")})
		return
	}
	if err != nil {
		respondRemuxStartFailure(c, err)
		return
	}
	probeContext, cancel := context.WithTimeout(c.Request.Context(), ffmpegTimeout())
	defer cancel()
	probeOutput, err := handler.probeRunner(probeContext, buildRemuxStartProbeArguments(videoFile.ResolveContentPath(), atSeconds))
	if err != nil {
		respondRemuxStartFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, RemuxStartDto{StartSeconds: parseRemuxStartSeconds(probeOutput, atSeconds)})
}

func respondRemuxStartFailure(c *gin.Context, err error) {
	if errors.Is(err, exec.ErrNotFound) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_REMUX_START_UNAVAILABLE")})
		return
	}
	applog.ErrorWithStack("video: remux start probe failed", err, "path", c.FullPath(), "ip", c.ClientIP())
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_OPERATION_FAILED")})
}
