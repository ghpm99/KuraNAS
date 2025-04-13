package worker_test

import (
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/tests/mocks"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func path_dir_test() string {
	currentDir, _ := os.Getwd()
	testDir := path.Join(currentDir, "testscan")
	return testDir
}

func TestScanFilesWorker(t *testing.T) {

	config.AppConfig.EntryPoint = path_dir_test()

	var filesCreated = []files.FileDto{}

	mockService := &mocks.MockService{
		GetFileByNameAndPathFunc: func(name string, path string) (files.FileDto, error) {
			return files.FileDto{}, os.ErrNotExist
		},
		CreateFileFunc: func(file files.FileDto) (files.FileDto, error) {
			file.ID = 1
			filesCreated = append(filesCreated, file)
			return file, nil
		},
	}

	worker.ScanFilesWorker(mockService)

	assert.Len(t, filesCreated, 4)

}

func TestScanFilesWorker_FileAlreadyExists(t *testing.T) {

	config.AppConfig.EntryPoint = path_dir_test()

	var filesCreated = []files.FileDto{}

	mockService := &mocks.MockService{
		GetFileByNameAndPathFunc: func(name string, path string) (files.FileDto, error) {
			return files.FileDto{ID: 1, Name: name, Path: path}, nil
		},
		CreateFileFunc: func(file files.FileDto) (files.FileDto, error) {
			file.ID = 1
			filesCreated = append(filesCreated, file)
			return file, nil
		},
	}

	worker.ScanFilesWorker(mockService)

	assert.Len(t, filesCreated, 0)
}
