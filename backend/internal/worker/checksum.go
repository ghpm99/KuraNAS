package worker

import (
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

func UpdateCheckSumWorker(service files.ServiceInterface, data any, logService logger.LoggerServiceInterface) {

	fileId, ok := data.(int)

	if !ok {
		log.Println("Erro ao converter ID do arquivo: data não é int")
		return
	}

	fileDto, err := service.GetFileById(fileId)

	if err != nil {
		log.Printf("Erro ao obter arquivo: %v\n", err)
		return
	}

	switch fileDto.Type {
	case files.File:
		updateFileCheckSum(service, fileDto)
	case files.Directory:
		updateDirectoryCheckSum(service, fileDto)
	}
}

func updateFileCheckSum(
	service files.ServiceInterface,
	fileDto files.FileDto,
) {
	checkSumHash, err := fileDto.GetCheckSumFromFile()

	if err != nil {
		log.Printf("Erro ao calcular checksum do arquivo: %v\n", err)
		return
	}

	fileDto.CheckSum = checkSumHash
	result, err := service.UpdateFile(fileDto)

	if err != nil || !result {
		log.Printf("Erro ao atualizar arquivo: %v\n", err)
		return
	}

	log.Printf("Checksum atualizado com sucesso para o arquivo: %s\n", fileDto.Name)

}

func updateDirectoryCheckSum(
	service files.ServiceInterface,
	fileDto files.FileDto,
) {

	var page = 1
	var hasNext = true
	var checkSumFiles []string

	for hasNext {

		filesInDirectory, err := service.GetFiles(files.FileFilter{
			ParentPath: utils.Optional[string]{
				Value:    fileDto.Path,
				HasValue: true,
			},
		}, page, 1000)

		if err != nil {
			log.Printf("Erro ao obter arquivos do diretório: %v\n", err)
			return
		}

		for _, file := range filesInDirectory.Items {
			checkSumFiles = append(checkSumFiles, file.CheckSum)
		}
		hasNext = filesInDirectory.Pagination.HasNext

		if hasNext {
			page = filesInDirectory.Pagination.Page + 1

		}
	}

	resultCheckSum := fileDto.GetCheckSumFromPath(checkSumFiles)

	fileDto.CheckSum = resultCheckSum
	result, err := service.UpdateFile(fileDto)

	if err != nil || !result {
		log.Printf("Erro ao atualizar diretório: %v\n", err)
		return
	}

	log.Printf("Checksum atualizado com sucesso para o diretório: %s\n", fileDto.Name)

}
