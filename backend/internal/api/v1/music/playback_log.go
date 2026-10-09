package music

import (
	"strconv"
	"strings"

	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

const byteRangeUnitPrefix = "bytes="

type playbackLog struct {
	service    logger.LoggerServiceInterface
	model      logger.LoggerModel
	isRecorded bool
}

func openPlaybackLog(service logger.LoggerServiceInterface, entry logger.LoggerModel, isPlaybackStart bool) playbackLog {
	if !isPlaybackStart {
		return playbackLog{service: service}
	}
	model, _ := service.CreateLog(entry, nil)
	return playbackLog{service: service, model: model, isRecorded: true}
}

func (playback playbackLog) succeed() {
	if playback.isRecorded {
		playback.service.CompleteWithSuccessLog(playback.model)
	}
}

func (playback playbackLog) fail(err error) {
	if playback.isRecorded {
		playback.service.CompleteWithErrorLog(playback.model, err)
	}
}

func isByteRangeAtStart(rangeHeader string) bool {
	if rangeHeader == "" {
		return true
	}
	firstRange, hasBytesUnit := strings.CutPrefix(rangeHeader, byteRangeUnitPrefix)
	if !hasBytesUnit {
		return false
	}
	startText, _, _ := strings.Cut(firstRange, "-")
	startByte, err := strconv.ParseInt(strings.TrimSpace(startText), 10, 64)
	return err == nil && startByte == 0
}

func isAudioStreamStart(c *gin.Context) bool {
	return isByteRangeAtStart(c.GetHeader("Range"))
}
