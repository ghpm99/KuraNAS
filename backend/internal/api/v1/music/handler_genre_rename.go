package music

import (
	"net/http"

	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) RenameGenreHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("RenameMusicGenre", "Renaming music genre", c), nil)

	var request RenameGenreRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	renamedGenre, err := handler.service.RenameGenre(c.Param("key"), request.Name)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondTrackMetadataError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, renamedGenre)
}
