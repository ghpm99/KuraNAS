package worker

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"os"
	"time"
)

func ScanDirWorker(service files.ServiceInterface, data string) {
	fmt.Println("🔍 Escaneando diretorio...")

	entries, err := os.ReadDir(data)
	if err != nil {
		fmt.Printf("❌ Erro ao ler diretório %s: %v\n", data, err)
		return
	}

	// Map de arquivos do diretório
	dirFileMap := make(map[string]files.FileDto)
	for _, entry := range entries {
		var fileDto = files.FileDto{}
		if err := fileDto.ParseDirEntryToFileDto(entry); err != nil {
			fmt.Printf("Erro ao obter informações: %v\n", err)
			continue
		}
		fileDto.Path = data
		dirFileMap[fileDto.Name] = fileDto
	}

	//Array de arquivos do cache
	cacheFileArray, err := service.GetFiles(files.FileFilter{
		Path: utils.Optional[string]{
			Value:    data,
			HasValue: true,
		},
	}, 1, 1000)

	if err != nil {
		fmt.Printf("Erro ao obter arquivos: %v\n", err)
		return
	}

	for _, file := range cacheFileArray.Items {
		file.DeletedAt = utils.Optional[time.Time]{
			Value:    time.Now(),
			HasValue: true,
		}
	}

	for _, cacheEntry := range cacheFileArray.Items {
		if _, ok := dirFileMap[cacheEntry.Name]; ok {
			delete(dirFileMap, cacheEntry.Name)
			cacheEntry.DeletedAt = utils.Optional[time.Time]{
				Value:    time.Time{},
				HasValue: false,
			}
		}
	}

	fmt.Println("🔍 Arquivos encontrados no cache:", len(cacheFileArray.Items))

	fmt.Println("🔍 Arquivos para deletar do cache:")
	for _, file := range cacheFileArray.Items {
		if !file.DeletedAt.HasValue {
			continue
		}
		fmt.Printf(" - %s\n", file.Name)
		service.UpdateFile(file)
	}

	fmt.Println("🔍 Arquivos novos encontrados no diretório:", len(dirFileMap))
	for _, file := range dirFileMap {
		service.CreateFile(file)
	}

}
