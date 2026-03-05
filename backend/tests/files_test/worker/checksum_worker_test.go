package worker_test

import (
	"crypto/sha256"
	"encoding/hex"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/worker"
	"sync"
	"testing"
)

func MockChecksum(path string) (string, error) {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:]), nil
}

func TestStartChecksumWorker(t *testing.T) {
	metadataProcessedChannel := make(chan files.FileDto, 5)
	checksumCompletedChannel := make(chan files.FileDto, 5)
	monitorChannel := make(chan worker.ResultWorkerData, 5)
	var workerGroup sync.WaitGroup

	testFiles := []files.FileDto{
		{
			Name:       "file1.txt",
			ParentPath: "/test/",
			Path:       "/test/file1.txt",
			Type:       files.File,
			Format:     ".txt",
		},
		{
			Name: "directory",
			Path: "/test/directory",
			Type: files.Directory,
		},
		{
			Name: "unknown",
			Path: "/test/unknown",
		},
	}
	for _, f := range testFiles {
		metadataProcessedChannel <- f
	}
	close(metadataProcessedChannel)

	workerGroup.Add(1)
	go worker.StartChecksumWorker(
		metadataProcessedChannel,
		checksumCompletedChannel,
		MockChecksum,
		MockChecksum,
		monitorChannel,
		&workerGroup,
	)

	var receivedFiles []files.FileDto
	var wgReader sync.WaitGroup
	wgReader.Add(1)
	go func() {
		defer wgReader.Done()
		for file := range checksumCompletedChannel {
			receivedFiles = append(receivedFiles, file)
		}
	}()

	workerGroup.Wait()
	close(checksumCompletedChannel)

	wgReader.Wait()

	if len(receivedFiles) != len(testFiles) {
		t.Errorf("Número de arquivos recebidos incorreto. Esperado %d, recebido %d", len(testFiles), len(receivedFiles))
	}

	for _, receivedFile := range receivedFiles {
		switch receivedFile.Type {
		case files.File:
			expectedChecksum := sha256.Sum256([]byte(receivedFile.Path))
			if receivedFile.CheckSum != hex.EncodeToString(expectedChecksum[:]) {
				t.Errorf("Checksum de arquivo incorreto. Esperado '%s', recebido '%s'", hex.EncodeToString(expectedChecksum[:]), receivedFile.CheckSum)
			}
		case files.Directory:
			expectedChecksum := sha256.Sum256([]byte(receivedFile.Path))
			if receivedFile.CheckSum != hex.EncodeToString(expectedChecksum[:]) {
				t.Errorf("Checksum de diretório incorreto. Esperado '%s', recebido '%s'", hex.EncodeToString(expectedChecksum[:]), receivedFile.CheckSum)
			}
		}

	}
}
