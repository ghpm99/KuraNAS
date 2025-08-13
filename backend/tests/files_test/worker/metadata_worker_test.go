package worker_test

import (
	"encoding/json"
	"errors"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/utils"
	"sync"
	"testing"
)

// mockScriptRunner é uma implementação mock de ScriptRunner.
func mockScriptRunner(scriptType utils.ScriptType, filePath string) (string, error) {
	switch scriptType {
	case utils.ImageMetadata:
		if filePath == "/test/image_with_error.jpg" {
			return "", errors.New("erro de script")
		}
		// Retorna um JSON simulado de metadados de imagem
		imgMetadata := files.ImageMetadataModel{
			Format: "JPEG",
			Width:  800,
			Height: 600,
		}
		jsonBytes, _ := json.Marshal(imgMetadata)
		return string(jsonBytes), nil
	case utils.VideoMetadata:
		// Retorna um JSON simulado de metadados de vídeo
		videoMetadata := files.VideoMetadataModel{
			FormatName: "mp4",
			Duration:   "120",
			Width:      1920,
			Height:     1080,
		}
		jsonBytes, _ := json.Marshal(videoMetadata)
		return string(jsonBytes), nil
	case utils.AudioMetadata:
		// Retorna um JSON simulado de metadados de áudio
		audioMetadata := files.AudioMetadataModel{
			Mime:   "mp3",
			Length: 180,
		}
		jsonBytes, _ := json.Marshal(audioMetadata)
		return string(jsonBytes), nil
	default:
		return "", errors.New("tipo de script desconhecido")
	}
}

func TestStartMetadataWorker(t *testing.T) {
	// 1. Configuração dos canais e WaitGroup
	fileDtoChannel := make(chan files.FileDto, 5)
	metadataProcessedChannel := make(chan files.FileDto, 5)
	monitorChannel := make(chan worker.ResultWorkerData, 5)
	var workerGroup sync.WaitGroup

	// 2. Popula o canal de entrada com dados de teste
	testFiles := []files.FileDto{
		{ID: 1, Name: "image.jpg", Path: "/test/image.jpg", Format: ".jpg", Type: files.File},
		{ID: 2, Name: "video.mp4", Path: "/test/video.mp4", Format: ".mp4", Type: files.File},
		{ID: 3, Name: "audio.mp3", Path: "/test/audio.mp3", Format: ".mp3", Type: files.File},
		{ID: 4, Name: "document.pdf", Path: "/test/document.pdf", Format: ".pdf", Type: files.File},
		{ID: 5, Name: "error.jpg", Path: "/test/image_with_error.jpg", Format: ".jpg", Type: files.File},
	}
	for _, f := range testFiles {
		fileDtoChannel <- f
	}
	close(fileDtoChannel)

	// 3. Executa a função em uma goroutine, passando o mock
	workerGroup.Add(1)
	go worker.StartMetadataWorker(fileDtoChannel, metadataProcessedChannel, mockScriptRunner, monitorChannel, &workerGroup)

	// 4. Ler do canal de saída para verificar os dados
	var receivedFiles []files.FileDto
	var wgReader sync.WaitGroup
	wgReader.Add(1)
	go func() {
		defer wgReader.Done()
		for file := range metadataProcessedChannel {
			receivedFiles = append(receivedFiles, file)
		}
	}()

	// 5. Espera a goroutine principal e a de leitura terminarem
	workerGroup.Wait()
	close(metadataProcessedChannel)
	wgReader.Wait()

	// 6. Validação dos resultados
	if len(receivedFiles) != len(testFiles) {
		t.Errorf("Número de arquivos recebidos incorreto. Esperado %d, recebido %d", len(testFiles), len(receivedFiles))
	}

	for _, file := range receivedFiles {
		switch file.ID {
		case 1:
			imgMetadata, ok := file.Metadata.(files.ImageMetadataModel)
			if !ok || imgMetadata.Format != "JPEG" {
				t.Errorf("Metadados de imagem incorretos para o arquivo %s", file.Name)
			}
		case 2:
			videoMetadata, ok := file.Metadata.(files.VideoMetadataModel)
			if !ok || videoMetadata.FormatName != "mp4" {
				t.Errorf("Metadados de vídeo incorretos para o arquivo %s", file.Name)
			}
		case 3:
			audioMetadata, ok := file.Metadata.(files.AudioMetadataModel)
			if !ok || audioMetadata.Mime != "mp3" {
				t.Errorf("Metadados de áudio incorretos para o arquivo %s", file.Name)
			}
		case 4:
			if file.Metadata != nil {
				t.Errorf("Metadados inesperados para o arquivo %s", file.Name)
			}
		case 5:
			if file.Metadata != nil {
				t.Errorf("Metadados inesperados para o arquivo com erro: %s", file.Name)
			}
		}
	}
}
