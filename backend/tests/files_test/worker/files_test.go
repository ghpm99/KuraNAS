package worker_test

import (
	"database/sql"
	"fmt"
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

var existingFiles = []files.FileDto{
	{Name: "testscan", Path: path.Join(path_dir_test(), ""), Type: files.Directory, Format: "", DeletedAt: utils.Optional[time.Time]{
		HasValue: false,
	}},
	{Name: "teste1.txt", Path: path.Join(path_dir_test(), "teste1.txt"), Type: files.File, Format: ".txt", DeletedAt: utils.Optional[time.Time]{
		HasValue: false,
	}},
	{Name: "teste2.pdf", Path: path.Join(path_dir_test(), "teste2.pdf"), Type: files.File, Format: ".pdf", DeletedAt: utils.Optional[time.Time]{
		HasValue: false,
	}},
	{Name: "teste3.xml", Path: path.Join(path_dir_test(), "teste3.xml"), Type: files.File, Format: ".xml", DeletedAt: utils.Optional[time.Time]{
		HasValue: false,
	}},
	{Name: "testepasta", Path: path.Join(path_dir_test(), "testepasta"), Type: files.Directory, Format: "", DeletedAt: utils.Optional[time.Time]{
		HasValue: false,
	}},
	{Name: "teste4.mp3", Path: path.Join(path_dir_test(), "testepasta/teste4.mp3"), Type: files.File, Format: ".mp3", DeletedAt: utils.Optional[time.Time]{
		HasValue: false,
	}},
}

func TestScanFilesWorker(t *testing.T) {

	config.AppConfig.EntryPoint = path_dir_test()

	var filesCreated = []files.FileDto{}

	mockService := &mocks.MockService{
		GetFileByNameAndPathFunc: func(name string, path string) (files.FileDto, error) {
			return files.FileDto{}, sql.ErrNoRows
		},
		CreateFileFunc: func(file files.FileDto) (files.FileDto, error) {
			file.ID = 1
			filesCreated = append(filesCreated, file)
			return file, nil
		},
		GetFilesFunc: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{}, fmt.Errorf("page index out of range")
		},
	}

	worker.ScanFilesWorker(mockService)

	assert.Len(t, filesCreated, len(existingFiles))

	for index, file := range filesCreated {
		var expectedFile = existingFiles[index]
		assert.Equal(t, expectedFile.Name, file.Name)
		assert.Equal(t, expectedFile.Path, file.Path)
		assert.Equal(t, expectedFile.Type, file.Type)
		assert.Equal(t, expectedFile.Format, file.Format)

	}
}

func TestScanFilesWorker_FileAlreadyExists(t *testing.T) {

	config.AppConfig.EntryPoint = path_dir_test()

	var filesCreated = []files.FileDto{}
	var filesUpdated = []files.FileDto{}

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
		GetFilesFunc: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{}, fmt.Errorf("page index out of range")
		},
	}

	worker.ScanFilesWorker(mockService)

	assert.Len(t, filesCreated, 0)
	assert.Len(t, filesUpdated, len(existingFiles))

	for index, file := range filesUpdated {
		var expectedFile = existingFiles[index]
		assert.Equal(t, expectedFile.Name, file.Name)
		assert.Equal(t, expectedFile.Path, file.Path)
		assert.Equal(t, expectedFile.Type, file.Type)
		assert.Equal(t, expectedFile.Format, file.Format)

	}
}

func TestScanFilesWorker_DontFindFileToDelete(t *testing.T) {

	config.AppConfig.EntryPoint = path_dir_test()

	var filesCreated = []files.FileDto{}
	var filesUpdated = []files.FileDto{}

	mockService := &mocks.MockService{
		GetFileByNameAndPathFunc: func(name string, path string) (files.FileDto, error) {
			return files.FileDto{}, os.ErrNotExist
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
		GetFilesFunc: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			var pageIndex = page - 1
			fmt.Println("Página index:", pageIndex)
			fmt.Println("Página atual:", page)
			fmt.Println("length:", len(existingFiles))
			fmt.Println("hasnext:", page < len(existingFiles))

			if page > len(existingFiles) {
				return utils.PaginationResponse[files.FileDto]{}, fmt.Errorf("page index out of range")
			}
			var pagination = utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					existingFiles[pageIndex],
				},
				Pagination: utils.Pagination{
					Page:     page,
					PageSize: pageSize,
					HasNext:  page < len(existingFiles),
				},
			}
			return pagination, nil
		},
	}

	worker.ScanFilesWorker(mockService)

	assert.Len(t, filesCreated, len(existingFiles))
	assert.Len(t, filesUpdated, 0)
}

func TestScanFilesWorker_FindFileToDelete(t *testing.T) {

	config.AppConfig.EntryPoint = path_dir_test()

	var filesCreated = []files.FileDto{}
	var filesUpdated = []files.FileDto{}
	var fileToDelete = append(existingFiles, files.FileDto{Name: "teste5.mp4", Path: path.Join(path_dir_test(), "testepasta/teste5.mp4"), Type: files.File, Format: ".mp4", DeletedAt: utils.Optional[time.Time]{
		HasValue: false,
	}})

	mockService := &mocks.MockService{
		GetFileByNameAndPathFunc: func(name string, path string) (files.FileDto, error) {
			return files.FileDto{}, sql.ErrNoRows
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
		GetFilesFunc: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			var pageIndex = page - 1

			if page > len(fileToDelete) {
				return utils.PaginationResponse[files.FileDto]{}, fmt.Errorf("page index out of range")
			}
			var pagination = utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					fileToDelete[pageIndex],
				},
				Pagination: utils.Pagination{
					Page:     page,
					PageSize: pageSize,
					HasNext:  page < len(fileToDelete),
				},
			}
			return pagination, nil
		},
	}

	worker.ScanFilesWorker(mockService)

	assert.Len(t, filesCreated, len(existingFiles))
	assert.Len(t, filesUpdated, 1)
}
