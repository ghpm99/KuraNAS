package worker

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"os"
)

func ScanDirHandler(service *files.Service, data string) {
	fmt.Println("🔍 Escaneando diretorio...")

	entries, err := os.ReadDir(data)
	if err != nil {
		fmt.Printf("❌ Erro ao ler diretório %s: %v\n", data, err)
		return
	}

	var dirFileArray []files.FileDto

	for _, entry := range entries {
		var fileDto = files.FileDto{}
		if err := fileDto.ParseDirEntryToFileDto(entry); err != nil {
			fmt.Printf("Erro ao obter informações: %v\n", err)
			continue
		}
		dirFileArray = append(dirFileArray, fileDto)
	}

}
