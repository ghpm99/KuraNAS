package files

import (
	"bytes"
	"fmt"
	"image/jpeg"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(financialService ServiceInterface) *Handler {
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

func (handler *Handler) GetFilesThreeHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	fileParentId := utils.ParseInt(c.DefaultQuery("file_parent", "0"), c)

	fileFilter := FileFilter{}

	if fileParentId != 0 {
		fileParent, err := handler.service.GetFileById(fileParentId)
		if err != nil {
			fmt.Println("Error getting file by ID:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if fileParent.ID != 0 {
			fileFilter.ParentPath = utils.Optional[string]{
				HasValue: true,
				Value:    fileParent.Path,
			}
		}
	} else {
		fileFilter.ParentPath = utils.Optional[string]{
			HasValue: true,
			Value:    config.AppConfig.EntryPoint,
		}
	}

	pagination, err := handler.service.GetFiles(fileFilter, page, pageSize)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetFileThumbnailHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)

	file, err := handler.service.GetFileById(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error1": err.Error()})
		return
	}

	thumbnail, err := handler.service.GetFileThumbnail(file, 320)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error2": err.Error()})
		return
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, thumbnail, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error3": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", buf.Bytes())
}
