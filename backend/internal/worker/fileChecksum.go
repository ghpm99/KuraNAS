package worker

import (
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"sync"
)

type ChecksumStepInput struct {
	File files.FileDto
}

type ChecksumStepOutput struct {
	File    files.FileDto
	Skipped bool
}

type ChecksumStepExecutor struct {
	getFileChecksum      func(path string) (string, error)
	getDirectoryChecksum func(dirPath string) (string, error)
}

func NewChecksumStepExecutor(
	getFileChecksum func(path string) (string, error),
	getDirectoryChecksum func(dirPath string) (string, error),
) *ChecksumStepExecutor {
	return &ChecksumStepExecutor{
		getFileChecksum:      getFileChecksum,
		getDirectoryChecksum: getDirectoryChecksum,
	}
}

func (e *ChecksumStepExecutor) Execute(input ChecksumStepInput) (ChecksumStepOutput, error) {
	checksum, err := getCheckSum(input.File, e.getFileChecksum, e.getDirectoryChecksum)
	if err != nil {
		return ChecksumStepOutput{File: input.File}, err
	}

	if input.File.CheckSum == checksum && checksum != "" {
		return ChecksumStepOutput{File: input.File, Skipped: true}, newStepSkipped("checksum up-to-date")
	}

	file := input.File
	file.CheckSum = checksum
	return ChecksumStepOutput{File: file}, nil
}

func StartChecksumWorker(
	metadataProcessedChannel <-chan files.FileDto,
	checksumCompletedChannel chan<- files.FileDto,
	getFileChecksum func(path string) (string, error),
	getDirectorysum func(dirPath string) (string, error),
	monitorChannel chan<- ResultWorkerData,
	workerGroup *sync.WaitGroup,
) {
	defer workerGroup.Done()
	executor := NewChecksumStepExecutor(getFileChecksum, getDirectorysum)

	for fileToProcess := range metadataProcessedChannel {
		output, err := executor.Execute(ChecksumStepInput{File: fileToProcess})

		if err != nil {
			if isStepSkipped(err) {
				checksumCompletedChannel <- output.File
				continue
			}

			log.Printf("Erro ao gerar checksum: %v\n", err)
			monitorChannel <- ResultWorkerData{
				Path:    fileToProcess.Path,
				Success: false,
				Error:   err.Error(),
			}
		} else {
			fileToProcess = output.File
		}
		checksumCompletedChannel <- fileToProcess
	}
}

func getCheckSum(fileDto files.FileDto,
	getFileChecksum func(path string) (string, error),
	getDirectoryChecksum func(dirPath string) (string, error),
) (string, error) {

	switch fileDto.Type {
	case files.File:
		return getFileChecksum(fileDto.Path)
	case files.Directory:
		return getDirectoryChecksum(fileDto.Path)
	default:
		return "", fmt.Errorf("file type not found")
	}

}
