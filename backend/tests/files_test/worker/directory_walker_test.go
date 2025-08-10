package worker_test

import (
	"log"
	"nas-go/api/internal/worker"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

const testDirectory = "testscan"

func TestStartDirectoryWalker(t *testing.T) {
	if _, err := os.Stat(testDirectory); os.IsNotExist(err) {
		t.Fatalf("O diretório de teste '%s' não existe. Por favor, crie-o com os arquivos de teste.", testDirectory)
	}

	fileWalkChannel := make(chan worker.FileWalk, 10)
	var workerGroup sync.WaitGroup
	workerGroup.Add(1)

	go worker.StartDirectoryWalker(testDirectory, fileWalkChannel, &workerGroup)

	var receivedPaths []string

	var wgReader sync.WaitGroup
	wgReader.Add(1)
	go func() {
		defer wgReader.Done()
		for fw := range fileWalkChannel {
			receivedPaths = append(receivedPaths, fw.Path)
			log.Printf("Recebido do canal: %s", fw.Path)
		}
	}()

	workerGroup.Wait()

	wgReader.Wait()

	expectedPaths := []string{
		testDirectory,
		filepath.Join(testDirectory, "documentos"),
		filepath.Join(testDirectory, "documentos", "conteudo_teste.pdf"),
		filepath.Join(testDirectory, "documentos", "conteudo_teste_transparente.pdf"),
		filepath.Join(testDirectory, "documentos", "teste2.pdf"),
		filepath.Join(testDirectory, "image"),
		filepath.Join(testDirectory, "image", "ChatGPT Image 28 de mar. de 2025, 20_45_52.png"),
		filepath.Join(testDirectory, "image", "ai-generated-8610368_1280.png"),
		filepath.Join(testDirectory, "teste1.txt"),
		filepath.Join(testDirectory, "teste3.xml"),
		filepath.Join(testDirectory, "testepasta"),
		filepath.Join(testDirectory, "testepasta", "teste4.mp3"),
	}

	if len(receivedPaths) != len(expectedPaths) {
		t.Errorf("Quantidade de caminhos recebidos incorreta. Esperado %d, recebido %d", len(expectedPaths), len(receivedPaths))
	}

	for _, expectedPath := range expectedPaths {
		found := false
		for _, receivedPath := range receivedPaths {
			if receivedPath == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("O caminho esperado '%s' não foi encontrado", expectedPath)
		}
	}
}
