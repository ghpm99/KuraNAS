package worker

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"strconv"
)

func UpdateCheckSumWorker(service files.ServiceInterface, data string, logService logger.LoggerServiceInterface) {

	logger, _ := logService.CreateLog(logger.LoggerModel{
		Name:        "ScanFilesWorker",
		Description: i18n.GetMessage("SCAN_FILES_START"),
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
	}, nil)

	fileId, err := strconv.Atoi(data)

	if err != nil {
		logService.CompleteWithErrorLog(logger, err)
		fmt.Printf("Erro ao converter ID do arquivo: %v\n", err)
		return
	}

	fileDto, err := service.GetFileById(fileId)

	if err != nil {
		logService.CompleteWithErrorLog(logger, err)
		fmt.Printf("Erro ao obter arquivo: %v\n", err)
		return
	}

	checkSumHash, err := fileDto.GetCheckSumFromFile()

	if err != nil {
		logService.CompleteWithErrorLog(logger, err)
		fmt.Printf("Erro ao calcular checksum do arquivo: %v\n", err)
		return
	}

	fileDto.CheckSum = checkSumHash
	result, err := service.UpdateFile(fileDto)

	if err != nil || !result {
		logService.CompleteWithErrorLog(logger, err)
		fmt.Printf("Erro ao atualizar arquivo: %v\n", err)
		return
	}

	fmt.Printf("Checksum atualizado com sucesso para o arquivo: %s\n", fileDto.Name)
	logService.CompleteWithSuccessLog(logger)
}
