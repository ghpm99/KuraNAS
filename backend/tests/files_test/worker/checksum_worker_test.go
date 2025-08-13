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
	// 1. Configuração dos canais e WaitGroup
	metadataProcessedChannel := make(chan files.FileDto, 5)
	checksumCompletedChannel := make(chan files.FileDto, 5)
	monitorChannel := make(chan worker.ResultWorkerData, 5)
	var workerGroup sync.WaitGroup

	// 2. Popula o canal de entrada com dados de teste
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

	// 3. Executa a função em uma goroutine
	workerGroup.Add(1)
	go worker.StartChecksumWorker(
		metadataProcessedChannel,
		checksumCompletedChannel,
		MockChecksum,
		MockChecksum,
		monitorChannel,
		&workerGroup,
	)

	// 4. Ler do canal de saída para verificar os dados processados
	var receivedFiles []files.FileDto
	var wgReader sync.WaitGroup
	wgReader.Add(1)
	go func() {
		defer wgReader.Done()
		for file := range checksumCompletedChannel {
			receivedFiles = append(receivedFiles, file)
		}
	}()

	// 5. Espera a goroutine principal terminar
	workerGroup.Wait()
	close(checksumCompletedChannel)

	// 6. Espera a goroutine de leitura terminar
	wgReader.Wait()

	// 7. Validação dos resultados
	if len(receivedFiles) != len(testFiles) {
		t.Errorf("Número de arquivos recebidos incorreto. Esperado %d, recebido %d", len(testFiles), len(receivedFiles))
	}

	for _, receivedFile := range receivedFiles {
		// Valida se o checksum foi gerado para arquivos e diretórios
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
