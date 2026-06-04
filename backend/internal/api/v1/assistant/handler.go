package assistant

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) ChatHandler(c *gin.Context) {
	if handler.service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		return
	}

	var req ChatRequestDto
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	response, err := handler.service.Chat(req.Messages)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidConversation):
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		case errors.Is(err, ErrAIUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}
