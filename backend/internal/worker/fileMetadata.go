package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"sync"
)

type ScriptRunner func(scriptType utils.ScriptType, filePath string) (string, error)

func StartMetadataWorker(
	fileDtoChannel <-chan files.FileDto,
	metadataProcessedChannel chan<- files.FileDto,
	runner ScriptRunner,
	workerGroup *sync.WaitGroup,
) {
	defer workerGroup.Done()

	for unprocessedFile := range fileDtoChannel {
		log.Println("StartMetadataWorker, Recendo arquivo de fila", unprocessedFile.Path)
		metadata, err := getMetadata(unprocessedFile, runner)

		if err != nil {
			log.Println(err)
		} else {
			unprocessedFile.Metadata = metadata
		}
		log.Println("StartMetadataWorker, Enviando arquivo para fila", unprocessedFile.Path)

		metadataProcessedChannel <- unprocessedFile
	}
}

func getMetadata(fileDto files.FileDto, runner ScriptRunner) (any, error) {
	formatType := utils.GetFormatTypeByExtension(fileDto.Format)

	switch formatType.Type {
	case utils.FormatTypeImage:
		return getImageMetadata(fileDto, runner)
	case utils.FormatTypeAudio:
		return getAudioMetadata(fileDto, runner)
	case utils.FormatTypeVideo:
		return getVideoMetadata(fileDto, runner)
	default:
		return nil, fmt.Errorf("sem metadata")
	}
}

func getImageMetadata(fileDto files.FileDto, runner ScriptRunner) (files.ImageMetadataModel, error) {
	metadata := files.ImageMetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	result, err := runner(utils.ImageMetadata, fileDto.Path)
	if err != nil {
		return metadata, err
	}
	log.Println("StartMetadataWorker, Resultado do script de imagem:", result)
	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}

func getAudioMetadata(fileDto files.FileDto, runner ScriptRunner) (files.AudioMetadataModel, error) {
	metadata := files.AudioMetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	result, err := runner(utils.AudioMetadata, fileDto.Path)
	if err != nil {
		return metadata, err
	}
	log.Println("StartMetadataWorker, Resultado do script de áudio:", result)

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}

func getVideoMetadata(fileDto files.FileDto, runner ScriptRunner) (files.VideoMetadataModel, error) {
	metadata := files.VideoMetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	result, err := runner(utils.VideoMetadata, fileDto.Path)
	if err != nil {
		return metadata, err
	}
	log.Println("StartMetadataWorker, Resultado do script de vídeo:", result)

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}
