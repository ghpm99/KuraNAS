package files

import (
	"fmt"

	"nas-go/api/internal/config"
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

	pagination := utils.Pagination{
		Page:     page,
		PageSize: pageSize,
		HasNext:  false,
		HasPrev:  false,
	}

	paginationResponse := utils.PaginationResponse[FileDto]{
		Items:      nil,
		Pagination: pagination,
	}

	path := c.DefaultQuery("path", config.AppConfig.EntryPoint)
	fmt.Println("entrypoint", config.AppConfig.EntryPoint)
	filter := FileFilter{
		Path: path,
	}

	err := handler.service.GetFiles(filter, &paginationResponse)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, paginationResponse)
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
