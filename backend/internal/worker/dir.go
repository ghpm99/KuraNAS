package worker

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"os"
)

func ScanDirHandler(service *files.Service, data string) {
	fmt.Println("🔍 Escaneando diretorio...")

	entries, err := os.ReadDir(data)
	if err != nil {
		fmt.Printf("❌ Erro ao ler diretório %s: %v\n", data, err)
		return
	}

	//Array de arquivos para adicionar
	var dirFileArray []files.FileDto
	//Array de arquivos para deletar
	var dirFileToDeleteArray []files.FileDto

	dirFileMap := make(map[string]files.FileDto)
	for _, entry := range entries {
		var fileDto = files.FileDto{}
		if err := fileDto.ParseDirEntryToFileDto(entry); err != nil {
			fmt.Printf("Erro ao obter informações: %v\n", err)
			continue
		}
		fileDto.Path = data
		dirFileArray = append(dirFileArray, fileDto)
		dirFileMap[fileDto.Name] = fileDto
	}

	var cacheFileArray []files.FileDto
	var fileDtoPagination = utils.PaginationResponse[files.FileDto]{
		Items: cacheFileArray,
		Pagination: utils.Pagination{
			Page:     1,
			PageSize: 100,
		},
	}

	if err := service.GetFiles(files.FileFilter{Path: data}, &fileDtoPagination); err != nil {
		fmt.Printf("Erro ao obter arquivos: %v\n", err)
		return
	}

	for _, cacheEntry := range fileDtoPagination.Items {
		if _, ok := dirFileMap[cacheEntry.Name]; ok {
			delete(dirFileMap, cacheEntry.Name)
		} else {
			dirFileToDeleteArray = append(dirFileToDeleteArray, cacheEntry)
		}
	}

}
