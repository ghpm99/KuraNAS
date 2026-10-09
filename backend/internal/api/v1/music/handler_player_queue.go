package music

import (
	"nas-go/api/pkg/i18n"
	"net/http"

	"github.com/gin-gonic/gin"
)

func respondInvalidRequest(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
}

func (handler *Handler) GetPlayerQueueHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetPlayerQueue", "Fetching player queue", c), nil)

	clientID, isClientIDValid := resolvePlayerClientID(c)
	if !isClientIDValid {
		respondInvalidRequest(c)
		return
	}

	queue, err := handler.service.GetPlayerQueue(clientID)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, queue)
}

func (handler *Handler) ReplacePlayerQueueHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("ReplacePlayerQueue", "Saving player queue", c), nil)

	clientID, isClientIDValid := resolvePlayerClientID(c)
	if !isClientIDValid {
		respondInvalidRequest(c)
		return
	}

	var request ReplacePlayerQueueRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondInvalidRequest(c)
		return
	}

	if err := handler.service.ReplacePlayerQueue(clientID, request); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.Status(http.StatusNoContent)
}
