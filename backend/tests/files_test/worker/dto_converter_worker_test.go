package worker_test

import (
	"io/fs"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/worker"
	"os"
	"sync"
	"testing"
	"time"
)

// MockFileInfo é uma implementação de os.FileInfo para testes.
type MockFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (mfi MockFileInfo) Name() string       { return mfi.name }
func (mfi MockFileInfo) Size() int64        { return mfi.size }
func (mfi MockFileInfo) Mode() fs.FileMode  { return mfi.mode }
func (mfi MockFileInfo) ModTime() time.Time { return mfi.modTime }
func (mfi MockFileInfo) IsDir() bool        { return mfi.isDir }
func (mfi MockFileInfo) Sys() interface{}   { return nil }

func TestStartDtoConverterWorker(t *testing.T) {
	// 1. Configuração dos canais e WaitGroup
	fileWalkChannel := make(chan worker.FileWalk, 5)
	fileDtoChannel := make(chan files.FileDto, 5)
	var workerGroup sync.WaitGroup

	// 2. Popula o canal de entrada com dados de teste
	testFileWalks := []worker.FileWalk{
		{
			Path: "/test/file1.txt",
			Info: MockFileInfo{
				name:    "file1.txt",
				size:    1024,
				mode:    0644,
				modTime: time.Date(2025, time.August, 10, 10, 0, 0, 0, time.UTC),
				isDir:   false,
			},
		},
		{
			Path: "/test/directory",
			Info: MockFileInfo{
				name:    "directory",
				size:    4096,
				mode:    os.ModeDir,
				modTime: time.Date(2025, time.August, 9, 9, 0, 0, 0, time.UTC),
				isDir:   true,
			},
		},
	}
	for _, fw := range testFileWalks {
		fileWalkChannel <- fw
	}
	close(fileWalkChannel)

	// 3. Executa a função em uma goroutine
	workerGroup.Add(1)
	go worker.StartDtoConverterWorker(fileWalkChannel, fileDtoChannel, &workerGroup)

	// 4. Ler do canal de saída para verificar os dados convertidos
	var receivedDtos []files.FileDto
	var wgReader sync.WaitGroup
	wgReader.Add(1)
	go func() {
		defer wgReader.Done()
		for dto := range fileDtoChannel {
			receivedDtos = append(receivedDtos, dto)
		}
	}()

	// 5. Espera a goroutine principal terminar
	workerGroup.Wait()

	// 6. Fecha o canal de saída após o worker terminar
	close(fileDtoChannel)

	// 7. Espera a goroutine de leitura terminar
	wgReader.Wait()

	// 8. Validação dos resultados
	if len(receivedDtos) != len(testFileWalks) {
		t.Errorf("Número de DTOs recebidos incorreto. Esperado %d, recebido %d", len(testFileWalks), len(receivedDtos))
	}

	// Valida os DTOs recebidos
	for i, dto := range receivedDtos {
		expectedFileWalk := testFileWalks[i]

		// Verifica o caminho
		if dto.Path != expectedFileWalk.Path {
			t.Errorf("Path incorreto. Esperado '%s', recebido '%s'", expectedFileWalk.Path, dto.Path)
		}

		// Verifica o nome do arquivo
		if dto.Name != expectedFileWalk.Info.Name() {
			t.Errorf("Nome incorreto. Esperado '%s', recebido '%s'", expectedFileWalk.Info.Name(), dto.Name)
		}

		// Verifica se a conversão de tipo está correta
		expectedType := files.File
		if expectedFileWalk.Info.IsDir() {
			expectedType = files.Directory
		}
		if dto.Type != expectedType {
			t.Errorf("Tipo incorreto. Esperado '%v', recebido '%v'", expectedType, dto.Type)
		}
	}
}
