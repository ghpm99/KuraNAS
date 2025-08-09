package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"sync"
)

func StartMetadataWorker(fileDtoChannel <-chan files.FileDto, metadataProcessedChannel chan<- files.FileDto, workerGroup *sync.WaitGroup) {
	defer workerGroup.Done()

	for unprocessedFile := range fileDtoChannel {
		log.Println("StartMetadataWorker, Recendo arquivo de fila", unprocessedFile.Path)
		metadata, err := getMetadata(unprocessedFile)

		if err != nil {
			log.Println(err)
		} else {
			unprocessedFile.Metadata = metadata
		}

		log.Println("StartMetadataWorker, Enviando arquivo para fila", unprocessedFile.Path)

		metadataProcessedChannel <- unprocessedFile
	}
}

func getMetadata(fileDto files.FileDto) (any, error) {
	formatType := utils.GetFormatTypeByExtension(fileDto.Format)

	switch formatType.Type {
	case utils.FormatTypeImage:
		return getImageMetadata(fileDto)
	case utils.FormatTypeAudio:
		return getAudioMetadata(fileDto)
	case utils.FormatTypeVideo:
		return getVideoMetadata(fileDto)
	default:
		return nil, fmt.Errorf("sem metadata")
	}
}

func getImageMetadata(fileDto files.FileDto) (files.ImageMetadataModel, error) {
	metadata := files.ImageMetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	result, err := utils.RunPythonScript(utils.ImageMetadata, fileDto.Path)
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}

func getAudioMetadata(fileDto files.FileDto) (files.AudioMetadataModel, error) {
	metadata := files.AudioMetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	result, err := utils.RunPythonScript(utils.AudioMetadata, fileDto.Path)
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}

func getVideoMetadata(fileDto files.FileDto) (files.VideoMetadataModel, error) {
	metadata := files.VideoMetadataModel{
		FileId: fileDto.ID,
		Path:   fileDto.Path,
	}

	result, err := utils.RunPythonScript(utils.VideoMetadata, fileDto.Path)
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}
