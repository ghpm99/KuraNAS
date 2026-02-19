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
	fileWalkChannel := make(chan worker.FileWalk, 5)
	fileDtoChannel := make(chan files.FileDto, 5)
	var workerGroup sync.WaitGroup

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

	workerGroup.Add(1)
	go worker.StartDtoConverterWorker(fileWalkChannel, fileDtoChannel, &workerGroup)

	var receivedDtos []files.FileDto
	var wgReader sync.WaitGroup
	wgReader.Add(1)
	go func() {
		defer wgReader.Done()
		for dto := range fileDtoChannel {
			receivedDtos = append(receivedDtos, dto)
		}
	}()

	workerGroup.Wait()

	close(fileDtoChannel)

	wgReader.Wait()

	if len(receivedDtos) != len(testFileWalks) {
		t.Errorf("Número de DTOs recebidos incorreto. Esperado %d, recebido %d", len(testFileWalks), len(receivedDtos))
	}

	for i, dto := range receivedDtos {
		expectedFileWalk := testFileWalks[i]

		if dto.Path != expectedFileWalk.Path {
			t.Errorf("Path incorreto. Esperado '%s', recebido '%s'", expectedFileWalk.Path, dto.Path)
		}

		if dto.Name != expectedFileWalk.Info.Name() {
			t.Errorf("Nome incorreto. Esperado '%s', recebido '%s'", expectedFileWalk.Info.Name(), dto.Name)
		}

		expectedType := files.File
		if expectedFileWalk.Info.IsDir() {
			expectedType = files.Directory
		}
		if dto.Type != expectedType {
			t.Errorf("Tipo incorreto. Esperado '%v', recebido '%v'", expectedType, dto.Type)
		}
	}
}
