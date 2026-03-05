package analytics

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

func (handler *Handler) GetOverviewHandler(c *gin.Context) {
	period := c.DefaultQuery("period", "7d")
	overview, err := handler.service.GetOverview(period)
	if err != nil {
		if errors.Is(err, ErrInvalidPeriod) {
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_ANALYTICS_INVALID_PERIOD")})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}
