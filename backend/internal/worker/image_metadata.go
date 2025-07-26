package worker

import (
	"encoding/json"
	"fmt"
	"log"
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

// TODO: trocar filePath por data que vai receber um json com fileId e filePath
// salvar o fileId no metadatos
func CreateImageMetadataWorker(
	service files.MetadataRepositoryInterface,
	data any,
	logService logger.LoggerServiceInterface,
) {

	newRegister := false
	fileDto, ok := data.(files.FileDto)

	if !ok {
		err := fmt.Errorf("data não é do tipo FileDto")
		log.Printf("Erro: %v\n", err)
		return
	}

	metadata, err := service.GetImageMetadataByID(fileDto.ID)

	if err != nil {
		metadata = files.ImageMetadataModel{
			FileId: fileDto.ID,
			Path:   fileDto.Path,
		}
		newRegister = true
	}

	result, err := RunPythonScript(fileDto.Path)
	if err != nil {
		log.Println("Erro:", err)
	} else {
		log.Println("Resultado:", result)
	}

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		log.Println("Erro ao converter JSON:", err)
	}

	if newRegister {
		metadata, err = service.CreateImageMetadata(metadata)
	} else {
		metadata, err = service.UpdateImageMetadata(metadata)
	}

	if err != nil {
		log.Printf("Erro ao criar metadados de imagem: %v\n", err)
		return
	}

	log.Printf("Metadados de imagem criados com sucesso: %+v\n", metadata)

}
