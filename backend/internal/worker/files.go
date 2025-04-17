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
		fileDto, err := service.GetFileByNameAndPath(name, path)

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
		} else {
			fileDto.Path = path
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

	findFilesDeleted(service)
}

func findFilesDeleted(service files.ServiceInterface) {
	var currentPage = 1
	var pagination, error = service.GetFiles(files.FileFilter{}, currentPage, 20)
	if error != nil {
		log.Printf("❌ Erro ao buscar arquivos: %v", error)
	}

	var filesArray []files.FileDto = pagination.Items
	for {
		for _, file := range filesArray {
			fmt.Printf("✅ Arquivo deletado ID: %d, %v\n", file.ID, file.Name)
		}
		if !pagination.Pagination.HasNext {
			break
		}
		currentPage++
		var pagination, err = service.GetFiles(files.FileFilter{}, currentPage, 20)
		if err != nil {
			log.Printf("❌ Erro ao buscar arquivos: %v", err)
			break
		}
		fmt.Println("findFilesDeleted", len(pagination.Items))
		fmt.Println("findFilesDeleted Items:", pagination.Items)
		filesArray = pagination.Items
	}

}
