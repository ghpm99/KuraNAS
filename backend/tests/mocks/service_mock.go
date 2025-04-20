package mocks

import (
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

type MockService struct {
	GetFilesFunc             func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	GetFilesByPathFunc       func(path string) (utils.PaginationResponse[files.FileDto], error)
	GetFileByNameAndPathFunc func(name string, path string) (files.FileDto, error)
	CreateFileFunc           func(fileDto files.FileDto) (files.FileDto, error)
	UpdateFileFunc           func(file files.FileDto) (bool, error)
	ScanFilesTaskFunc        func(data string)
	ScanDirTaskFunc          func(data string)
}

func (m *MockService) GetFiles(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return m.GetFilesFunc(filter, page, pageSize)
}

func (m *MockService) GetFilesByPath(path string) (utils.PaginationResponse[files.FileDto], error) {
	return m.GetFilesByPathFunc(path)
}

func (m *MockService) GetFileByNameAndPath(name string, path string) (files.FileDto, error) {
	return m.GetFileByNameAndPathFunc(name, path)
}

func (m *MockService) CreateFile(fileDto files.FileDto) (files.FileDto, error) {
	return m.CreateFileFunc(fileDto)
}

func (m *MockService) UpdateFile(file files.FileDto) (bool, error) {
	return m.UpdateFileFunc(file)
}

func (m *MockService) ScanFilesTask(data string) {
	m.ScanFilesTaskFunc(data)
}

func (m *MockService) ScanDirTask(data string) {
	m.ScanDirTaskFunc(data)
}
