package worker

import (
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"os"
	"path/filepath"
	"time"
)

func ScanFilesWorker(service files.ServiceInterface) {
	fmt.Println("🔍 Escaneando arquivos em:", config.AppConfig.EntryPoint)

	err := filepath.Walk(config.AppConfig.EntryPoint, func(path string, info os.FileInfo, err error) error {
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
		pathDir := filepath.Dir(path)
		fmt.Printf("📄 Arquivo: %s, Extensão: %s, Tamanho: %d bytes\n", name, ext, size)
		fileDto, err := service.GetFileByNameAndPath(name, pathDir)

		if err == nil {
			fmt.Printf("❌ Arquivo ja cadastrado %s: %v\n", pathDir, fileDto.ID)
			return nil
		}

		file := files.FileDto{
			Name:            name,
			Path:            pathDir,
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
