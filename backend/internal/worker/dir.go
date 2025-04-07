package worker

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"os"
	"time"
)

func ScanDirWorker(service *files.Service, data string) {
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
	cacheFileArray, err := service.GetFilesByPath(data)

	if err != nil {
		fmt.Printf("Erro ao obter arquivos: %v\n", err)
		return
	}

	for _, file := range cacheFileArray {
		file.DeletedAt = time.Now()
	}

	for _, cacheEntry := range cacheFileArray {
		if _, ok := dirFileMap[cacheEntry.Name]; ok {
			delete(dirFileMap, cacheEntry.Name)
			cacheEntry.DeletedAt = time.Time{}
		}
	}

	fmt.Println("🔍 Arquivos encontrados no cache:", len(cacheFileArray))

	fmt.Println("🔍 Arquivos para deletar do cache:")
	for _, file := range cacheFileArray {
		if file.DeletedAt.IsZero() {
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
