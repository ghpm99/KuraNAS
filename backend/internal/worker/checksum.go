package worker

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"strconv"
)

func UpdateCheckSumWorker(service files.ServiceInterface, data string) {

	fileId, err := strconv.Atoi(data)

	if err != nil {
		fmt.Printf("Erro ao converter ID do arquivo: %v\n", err)
		return
	}

	fileDto, err := service.GetFileById(fileId)

	if err != nil {
		fmt.Printf("Erro ao obter arquivo: %v\n", err)
		return
	}

	checkSumHash, err := fileDto.GetCheckSumFromFile()

	if err != nil {
		fmt.Printf("Erro ao calcular checksum do arquivo: %v\n", err)
		return
	}

	fileDto.CheckSum = checkSumHash
	result, err := service.UpdateFile(fileDto)

	if err != nil || !result {
		fmt.Printf("Erro ao atualizar arquivo: %v\n", err)
		return
	}

	fmt.Printf("Checksum atualizado com sucesso para o arquivo: %s\n", fileDto.Name)
}
