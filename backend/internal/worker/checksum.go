package worker

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"strconv"
)

func UpdateCheckSumWorker(service files.ServiceInterface, data string, logService logger.LoggerServiceInterface) {

	loggerModel, _ := logService.CreateLog(logger.LoggerModel{
		Name:        "UpdateCheckSumWorker",
		Description: i18n.GetMessage("SCAN_FILES_START"),
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
	}, nil)

	fileId, err := strconv.Atoi(data)

	if err != nil {
		logService.CompleteWithErrorLog(loggerModel, err)
		fmt.Printf("Erro ao converter ID do arquivo: %v\n", err)
		return
	}

	fileDto, err := service.GetFileById(fileId)

	if err != nil {
		logService.CompleteWithErrorLog(loggerModel, err)
		fmt.Printf("Erro ao obter arquivo: %v\n", err)
		return
	}

	if fileDto.Type == files.File {
		updateFileCheckSum(service, fileDto, logService, loggerModel)
	} else if fileDto.Type == files.Directory {
		updateDirectoryCheckSum(service, fileDto, logService, loggerModel)
	}
}

func updateFileCheckSum(
	service files.ServiceInterface,
	fileDto files.FileDto,
	logService logger.LoggerServiceInterface,
	loggerModel logger.LoggerModel,
) {
	checkSumHash, err := fileDto.GetCheckSumFromFile()

	if err != nil {
		logService.CompleteWithErrorLog(loggerModel, err)
		fmt.Printf("Erro ao calcular checksum do arquivo: %v\n", err)
		return
	}

	fileDto.CheckSum = checkSumHash
	result, err := service.UpdateFile(fileDto)

	if err != nil || !result {
		logService.CompleteWithErrorLog(loggerModel, err)
		fmt.Printf("Erro ao atualizar arquivo: %v\n", err)
		return
	}

	fmt.Printf("Checksum atualizado com sucesso para o arquivo: %s\n", fileDto.Name)
	logService.CompleteWithSuccessLog(loggerModel)
}

func updateDirectoryCheckSum(
	service files.ServiceInterface,
	fileDto files.FileDto,
	logService logger.LoggerServiceInterface,
	loggerModel logger.LoggerModel,
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
			logService.CompleteWithErrorLog(loggerModel, err)
			fmt.Printf("Erro ao obter arquivos do diretório: %v\n", err)
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
		logService.CompleteWithErrorLog(loggerModel, err)
		fmt.Printf("Erro ao atualizar diretório: %v\n", err)
		return
	}

	fmt.Printf("Checksum atualizado com sucesso para o diretório: %s\n", fileDto.Name)
	logService.CompleteWithSuccessLog(loggerModel)
}
