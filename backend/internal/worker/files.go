package worker

import (
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"os"
	"path/filepath"
	"time"
)

func ScanFilesHandler(service *files.Service) {
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
		fileDto, err := service.GetFileByNameAndPath(name, path)

		if err == nil {
			fmt.Printf("❌ Arquivo ja cadastrado %s: %v\n", path, fileDto.ID)
			return nil
		}

		file := files.FileDto{
			Name:            name,
			Path:            path,
			Format:          ext,
			Size:            size,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			LastInteraction: time.Now(),
			LastBackup:      time.Now(),
		}
		fileCreated, err := service.CreateFile(file)

		if err != nil {
			fmt.Printf("❌ Erro ao escanear arquivo %s: %v\n", path, err)
			return nil
		}
		fmt.Printf("✅ Arquivo criado ID: %d\n", fileCreated.ID)
		return nil
	})

	if err != nil {
		log.Printf("❌ Erro ao escanear arquivos: %v", err)
	} else {
		fmt.Println("✅ Escaneamento concluído!")
	}
}
