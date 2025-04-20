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

	fileParent := utils.ParseInt(c.DefaultQuery("file_parent", "0"), c)

	filter := FileFilter{
		FileParent: utils.Optional[int]{
			HasValue: fileParent != 0,
			Value:    fileParent,
		},
	}

	pagination, err := handler.service.GetFiles(filter, page, pageSize)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetFilesByPathHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	path := c.DefaultQuery("path", config.AppConfig.EntryPoint)

	pagination, err := handler.service.GetFiles(FileFilter{
		Path: utils.Optional[string]{
			HasValue: true,
			Value:    path,
		},
	}, page, pageSize)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetChildrenByIdHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)
	id := utils.ParseInt(c.Param("id"), c)

	file, err := handler.service.GetFiles(FileFilter{
		ID: utils.Optional[int]{
			HasValue: true,
			Value:    id,
		},
	}, page, pageSize)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination, err := handler.service.GetFiles(FileFilter{
		Path: utils.Optional[string]{
			HasValue: true,
			Value:    file.Items[0].Path,
		},
	}, page, pageSize)

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
