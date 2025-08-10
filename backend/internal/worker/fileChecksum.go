package worker

import (
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"sync"
)

func StartChecksumWorker(
	metadataProcessedChannel <-chan files.FileDto,
	checksumCompletedChannel chan<- files.FileDto,
	getFileChecksum func(path string) (string, error),
	getDirectorysum func(dirPath string) (string, error),
	workerGroup *sync.WaitGroup,
) {
	defer workerGroup.Done()
	defer close(checksumCompletedChannel)

	for fileToProcess := range metadataProcessedChannel {
		log.Println("StartChecksumWorker, Recendo arquivo de fila", fileToProcess.Path)
		checksum, err := getCheckSum(fileToProcess, getFileChecksum, getDirectorysum)

		if err != nil {
			log.Printf("Erro ao gerar checksum: %v\n", err)
		} else {
			fileToProcess.CheckSum = checksum
		}
		log.Println("StartChecksumWorker, enviando arquivo para fila", fileToProcess.Path)
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
