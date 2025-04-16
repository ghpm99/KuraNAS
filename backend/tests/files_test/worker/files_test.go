package worker_test

import (
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/pkg/utils"
	"nas-go/api/tests/mocks"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func path_dir_test() string {
	currentDir, _ := os.Getwd()
	testDir := path.Join(currentDir, "testscan")
	return testDir
}

func TestScanFilesWorker(t *testing.T) {

	config.AppConfig.EntryPoint = path_dir_test()

	var expectedFiles = []files.FileDto{
		{Name: "test1.txt", Path: path.Join(path_dir_test(), "test1.txt"), Type: files.File, Format: "txt", DeletedAt: utils.Optional[time.Time]{
			Value:    time.Time{},
			HasValue: false,
		}},
		{Name: "test2.txt", Path: path.Join(path_dir_test(), "test2.pdf"), Type: files.File, Format: "pdf"},
		{Name: "test3.txt", Path: path.Join(path_dir_test(), "test3.xml"), Type: files.File, Format: "xml"},
		{Name: "test4.txt", Path: path.Join(path_dir_test(), "testescan/test4.mp3"), Type: files.File, Format: "mp3"},
		{Name: "testscan", Path: path.Join(path_dir_test(), "testscan"), Type: files.Directory, Format: ""},
		{Name: "testepasta", Path: path.Join(path_dir_test(), "testscan/testepasta"), Type: files.Directory, Format: ""},
	}

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

	assert.Len(t, filesCreated, len(expectedFiles))

}

func TestScanFilesWorker_FileAlreadyExists(t *testing.T) {

	config.AppConfig.EntryPoint = path_dir_test()

	var filesCreated = []files.FileDto{}
	var filesUpdated = []files.FileDto{}
	var expectedFiles = []files.FileDto{
		{Name: "test1.txt", Path: path.Join(path_dir_test(), "test1.txt"), Type: files.File, Format: "txt", DeletedAt: utils.Optional[time.Time]{
			Value:    time.Time{},
			HasValue: false,
		}},
		{Name: "test2.txt", Path: path.Join(path_dir_test(), "test2.pdf"), Type: files.File, Format: "pdf"},
		{Name: "test3.txt", Path: path.Join(path_dir_test(), "test3.xml"), Type: files.File, Format: "xml"},
		{Name: "test4.txt", Path: path.Join(path_dir_test(), "testescan/test4.mp3"), Type: files.File, Format: "mp3"},
		{Name: "testscan", Path: path.Join(path_dir_test(), "testscan"), Type: files.Directory, Format: ""},
		{Name: "testepasta", Path: path.Join(path_dir_test(), "testscan/testepasta"), Type: files.Directory, Format: ""},
	}

	mockService := &mocks.MockService{
		GetFileByNameAndPathFunc: func(name string, path string) (files.FileDto, error) {
			return files.FileDto{ID: 1, Name: name, Path: path}, nil
		},
		CreateFileFunc: func(file files.FileDto) (files.FileDto, error) {
			file.ID = 1
			filesCreated = append(filesCreated, file)
			return file, nil
		},
		UpdateFileFunc: func(fileDto files.FileDto) (bool, error) {
			filesUpdated = append(filesUpdated, fileDto)
			return true, nil
		},
	}

	worker.ScanFilesWorker(mockService)

	assert.Len(t, filesCreated, 0)
	assert.Len(t, filesUpdated, len(expectedFiles))
}
