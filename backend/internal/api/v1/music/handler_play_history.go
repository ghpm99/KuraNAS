package music

import (
	"net/http"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) RecordPlayHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("RecordMusicPlay", "Recording music play", c), nil)

	clientID, isClientIDValid := resolvePlayerClientID(c)
	if !isClientIDValid {
		respondInvalidRequest(c)
		return
	}

	var request RecordPlayRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondInvalidRequest(c)
		return
	}

	if err := handler.service.RecordPlay(clientID, request); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.Status(http.StatusNoContent)
}

func (handler *Handler) GetMostPlayedTracksHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	period := PlayPeriod(c.DefaultQuery("period", string(PlayPeriodAll)))
	handler.respondLibraryTracks(c, "GetMostPlayedTracks", "Fetching most played tracks", func() (any, error) {
		return handler.service.GetMostPlayedTracks(period, page, pageSize)
	})
}

func (handler *Handler) GetRecentlyPlayedTracksHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	handler.respondLibraryTracks(c, "GetRecentlyPlayedTracks", "Fetching recently played tracks", func() (any, error) {
		return handler.service.GetRecentlyPlayedTracks(page, pageSize)
	})
}
