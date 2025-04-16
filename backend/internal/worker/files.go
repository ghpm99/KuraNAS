package worker

import (
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"os"
	"path/filepath"
)

func ScanFilesWorker(service files.ServiceInterface) {
	fmt.Println("🔍 Escaneando arquivos...")

	err := filepath.Walk(config.AppConfig.EntryPoint, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("❌ Erro ao escanear arquivo %s: %v\n", path, err)
			return nil
		}
		name := info.Name()
		pathDir := filepath.Dir(path)
		fileDto, err := service.GetFileByNameAndPath(name, pathDir)

		if err := fileDto.ParseFileInfoToFileDto(info); err != nil {
			fmt.Printf("Erro ao obter informações: %v\n", err)
			return nil
		}

		if fileDto.ID != 0 {
			updated, err := service.UpdateFile(fileDto)
			if err != nil || !updated {
				fmt.Printf("❌ Erro ao atualizar arquivo %s: %v\n", path, err)
				return nil
			}
			fmt.Printf("✅ Arquivo atualizado ID: %d\n", fileDto.ID)
			return nil
		}

		fileCreated, err := service.CreateFile(fileDto)

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

func scanDir(path string, info os.FileInfo) {

}

func scanFile(path string, info os.FileInfo) {

}
