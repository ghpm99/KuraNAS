package images

import (
	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(financialService *Service) *Handler {
	return &Handler{service: financialService}
}

func (handler *Handler) GetAllImagesHandler(c *gin.Context) {

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	pagination := utils.Pagination{
		Page:     page,
		PageSize: pageSize,
		HasNext:  false,
		HasPrev:  false,
	}

	images, err := handler.service.GetAllImages(pagination)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}
