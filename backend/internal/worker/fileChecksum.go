package worker

import (
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"sync"
)

func StartChecksumWorker(metadataProcessedChannel <-chan files.FileDto, checksumCompletedChannel chan<- files.FileDto, workerGroup *sync.WaitGroup) {
	defer workerGroup.Done()
	defer close(checksumCompletedChannel)

	for fileToProcess := range metadataProcessedChannel {
		log.Println("StartChecksumWorker, Recendo arquivo de fila", fileToProcess.Path)
		checksum, err := getCheckSum(fileToProcess)

		if err != nil {
			log.Printf("Erro ao gerar checksum: %v\n", err)
		} else {
			fileToProcess.CheckSum = checksum
		}
		log.Println("StartChecksumWorker, enviando arquivo para fila", fileToProcess.Path)
		checksumCompletedChannel <- fileToProcess
	}
}

func getCheckSum(fileDto files.FileDto) (string, error) {

	switch fileDto.Type {
	case files.File:
		return getFileCheckSum(fileDto)
	case files.Directory:
		return getDirectoryCheckSum(fileDto)
	default:
		return "", fmt.Errorf("file type not found")
	}

}

func getFileCheckSum(
	fileDto files.FileDto,
) (string, error) {
	return fileDto.GetCheckSumFromFile()
}

func getDirectoryCheckSum(fileDto files.FileDto) (string, error) {

	var checkSumFiles []string

	// TODO: buscar todos arquivos na pasta e adicionar checksum no array

	resultCheckSum := fileDto.GetCheckSumFromPath(checkSumFiles)

	return resultCheckSum, nil
}
