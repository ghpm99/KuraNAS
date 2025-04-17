package files

import (
	"fmt"

	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(financialService *Service) *Handler {
	return &Handler{
		service: financialService,
	}
}

func (handler *Handler) GetFilesHandler(c *gin.Context) {

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	fileParent := utils.ParseInt(c.DefaultQuery("file_parent", "0"), c)

	filter := FileFilter{
		FileParent: fileParent,
	}

	pagination, err := handler.service.GetFiles(filter, page, pageSize)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) UpdateFilesHandler(c *gin.Context) {
	data := c.PostForm("data")
	fmt.Println("📁 Recebendo dados para processamento:", data)
	if data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data is required"})
		return
	}
	handler.service.ScanFilesTask(data)
}
