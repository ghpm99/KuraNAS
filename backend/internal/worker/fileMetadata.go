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

type MetadataStepInput struct {
	File files.FileDto
}

type MetadataStepOutput struct {
	File    files.FileDto
	Skipped bool
}

type MetadataStepExecutor struct {
	runner ScriptRunner
}

func NewMetadataStepExecutor(runner ScriptRunner) *MetadataStepExecutor {
	if runner == nil {
		runner = pythonScriptRunner
	}
	return &MetadataStepExecutor{runner: runner}
}

func (e *MetadataStepExecutor) Execute(input MetadataStepInput) (MetadataStepOutput, error) {
	file := input.File
	if file.Path == "" {
		return MetadataStepOutput{File: file}, fmt.Errorf("metadata step: file path is required")
	}

	if shouldSkipMetadataStep(file) {
		return MetadataStepOutput{File: file, Skipped: true}, newStepSkipped("metadata up-to-date")
	}

	metadata, err := getMetadata(file, e.runner)
	if err != nil {
		return MetadataStepOutput{File: file}, err
	}

	file.Metadata = metadata
	return MetadataStepOutput{File: file}, nil
}

func shouldSkipMetadataStep(file files.FileDto) bool {
	formatType := utils.GetFormatTypeByExtension(file.Format)
	if formatType.Type != utils.FormatTypeImage &&
		formatType.Type != utils.FormatTypeAudio &&
		formatType.Type != utils.FormatTypeVideo {
		return true
	}

	return file.Metadata != nil
}

func StartMetadataWorker(
	fileDtoChannel <-chan files.FileDto,
	metadataProcessedChannel chan<- files.FileDto,
	runner ScriptRunner,
	monitorChannel chan<- ResultWorkerData,
	workerGroup *sync.WaitGroup,
) {
	defer workerGroup.Done()
	executor := NewMetadataStepExecutor(runner)

	for unprocessedFile := range fileDtoChannel {
		output, err := executor.Execute(MetadataStepInput{File: unprocessedFile})

		if err != nil {
			if isStepSkipped(err) {
				metadataProcessedChannel <- output.File
				continue
			}

			log.Println(err)
			monitorChannel <- ResultWorkerData{
				Path:    unprocessedFile.Path,
				Success: false,
				Error:   err.Error(),
			}
		} else {
			unprocessedFile = output.File
		}

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
		return nil, nil
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

	err = json.Unmarshal([]byte(result), &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}
