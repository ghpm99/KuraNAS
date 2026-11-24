package video

import (
	"database/sql"
	"errors"
	"net/http"
	"os/exec"
	"strconv"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const audioStreamIndexQueryParam = "audio_stream_index"

type AudioTrackFileSource interface {
	GetFileById(id int) (files.FileDto, error)
}

type AudioTrackHandler struct {
	fileSource AudioTrackFileSource
	service    AudioTrackServiceInterface
}

func NewAudioTrackHandler(fileSource AudioTrackFileSource, service AudioTrackServiceInterface) *AudioTrackHandler {
	return &AudioTrackHandler{fileSource: fileSource, service: service}
}

func (h *AudioTrackHandler) ListAudioTracksHandler(c *gin.Context) {
	fileID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	videoFile, err := h.fileSource.GetFileById(fileID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && videoFile.DeletedAt.HasValue) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_NOT_FOUND")})
		return
	}
	if err != nil {
		respondAudioTrackFailure(c, err)
		return
	}
	audioTracks, err := h.service.ListAudioTracks(videoFile)
	if err != nil {
		respondAudioTrackFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, AudioTracksDto{Items: audioTracks})
}

func parseAudioStreamIndex(c *gin.Context) (audioStreamIndex int, isValid bool) {
	rawIndex, isProvided := c.GetQuery(audioStreamIndexQueryParam)
	if !isProvided {
		return defaultAudioStreamIndex, true
	}
	audioStreamIndex, err := strconv.Atoi(rawIndex)
	if err != nil || audioStreamIndex < 0 {
		return 0, false
	}
	return audioStreamIndex, true
}

func isRequestedAudioStreamAvailable(c *gin.Context, service AudioTrackServiceInterface, videoFile files.FileDto, audioStreamIndex int) bool {
	if audioStreamIndex == defaultAudioStreamIndex {
		return true
	}
	if service == nil {
		respondInvalidAudioStream(c)
		return false
	}
	hasAudioStream, err := service.HasAudioStream(videoFile, audioStreamIndex)
	if err != nil {
		respondAudioTrackFailure(c, err)
		return false
	}
	if !hasAudioStream {
		respondInvalidAudioStream(c)
		return false
	}
	return true
}

func respondInvalidAudioStream(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_INVALID_AUDIO_STREAM")})
}

func respondAudioTrackFailure(c *gin.Context, err error) {
	if errors.Is(err, exec.ErrNotFound) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_AUDIO_TRACKS_UNAVAILABLE")})
		return
	}
	applog.ErrorWithStack("video: audio track probe failed", err, "path", c.FullPath(), "ip", c.ClientIP())
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_OPERATION_FAILED")})
}
