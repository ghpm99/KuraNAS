package files

import (
	"fmt"
	"log"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(financialService *Service) *Handler {
	return &Handler{service: financialService}
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

	err := handler.service.GetFiles(&paginationResponse)

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

	task := worker.Task{Data: data}

	worker.Tasks <- task
}

func (handler *Handler) ScanFilesHandler() {
	fmt.Println("🔍 Escaneando arquivos...")

	basePath := "/mnt/d/"

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("❌ Erro ao escanear arquivo %s: %v\n", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		name := info.Name()
		ext := filepath.Ext(name)
		size := info.Size()

		fmt.Printf("📄 Arquivo: %s, Extensão: %s, Tamanho: %d bytes\n", name, ext, size)

		return nil
	})

	if err != nil {
		log.Printf("❌ Erro ao escanear arquivos: %v", err)
	} else {
		fmt.Println("✅ Escaneamento concluído!")
	}
}
