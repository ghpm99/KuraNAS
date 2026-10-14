package video

import (
	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type streamAccessLog struct {
	logService  logger.LoggerServiceInterface
	context     *gin.Context
	name        string
	description string
	model       logger.LoggerModel
	isOpen      bool
}

func newStreamAccessLog(logService logger.LoggerServiceInterface, c *gin.Context, name, description string) *streamAccessLog {
	return &streamAccessLog{logService: logService, context: c, name: name, description: description}
}

func (accessLog *streamAccessLog) open() {
	if accessLog.isOpen {
		return
	}
	accessLog.model, _ = accessLog.logService.CreateLog(browseLogEntry(accessLog.name, accessLog.description, accessLog.context), nil)
	accessLog.isOpen = true
}

func (accessLog *streamAccessLog) succeed() {
	if !accessLog.isOpen {
		return
	}
	accessLog.logService.CompleteWithSuccessLog(accessLog.model)
}

func (accessLog *streamAccessLog) fail(err error) {
	accessLog.open()
	accessLog.logService.CompleteWithErrorLog(accessLog.model, err)
}

func isPlaybackStartRange(rangeHeader string, parsedRangeStart int64) bool {
	return rangeHeader == "" || parsedRangeStart == 0
}
