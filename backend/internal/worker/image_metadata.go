package worker

import (
	"encoding/json"
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/logger"
	"os/exec"
)

func RunPythonScript(imagePath string) (string, error) {
	cmd := exec.Command("scripts/.venv/bin/python", "scripts/image_metadata.py", imagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("erro ao executar script python: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

func CreateImageMetadataWorker(service files.MetadataRepositoryInterface, filePath string, logService logger.LoggerServiceInterface) {
	loggerModel, _ := logService.CreateLog(logger.LoggerModel{
		Name:        "CreateImageMetadataWorker",
		Description: "Extraindo metadados de imagem",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
	}, nil)

	if filePath == "" {
		err := fmt.Errorf("caminho do arquivo não encontrado")
		logService.CompleteWithErrorLog(loggerModel, err)
		fmt.Printf("Erro: %v\n", err)
		return
	}

	result, err := RunPythonScript(filePath)
	if err != nil {
		fmt.Println("Erro:", err)
	} else {
		fmt.Println("Resultado:", result)
	}

	var metadata files.ImageMetadataModel
	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		fmt.Println("Erro ao converter JSON:", err)
	}
	createdMetadata, err := service.CreateImageMetadata(metadata)
	if err != nil {
		logService.CompleteWithErrorLog(loggerModel, err)
		fmt.Printf("Erro ao criar metadados de imagem: %v\n", err)
		return
	}
	logService.CompleteWithSuccessLog(loggerModel)
	fmt.Printf("Metadados de imagem criados com sucesso: %+v\n", createdMetadata)

}
