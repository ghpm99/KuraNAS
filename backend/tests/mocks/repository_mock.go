package mocks

import (
	"database/sql"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

type MockRepository struct {
	GetDbContextFunc         *sql.DB
	GetFilesFunc             func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	GetFilesByPathFunc       func(path string) ([]files.FileModel, error)
	GetFileByNameAndPathFunc func(name string, path string) (files.FileModel, error)
	CreateFileFunc           func(transaction *sql.Tx, file files.FileModel) (files.FileModel, error)
	UpdateFileFunc           func(transaction *sql.Tx, file files.FileModel) (bool, error)
	GetPathByFileIdFunc      func(fileParent int) (string, error)
}

func (m *MockRepository) GetDbContext() *sql.DB {
	return nil
}

func (m *MockRepository) GetFiles(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
	return m.GetFilesFunc(filter, page, pageSize)
}

func (m *MockRepository) GetFilesByPath(path string) ([]files.FileModel, error) {
	return nil, nil
}

func (m *MockRepository) GetFileByNameAndPath(name string, path string) (files.FileModel, error) {
	return files.FileModel{}, nil
}

func (m *MockRepository) CreateFile(transaction *sql.Tx, file files.FileModel) (files.FileModel, error) {
	return files.FileModel{}, nil
}

func (m *MockRepository) UpdateFile(transaction *sql.Tx, file files.FileModel) (bool, error) {
	return false, nil
}

func (m *MockRepository) GetPathByFileId(fileParent int) (string, error) {
	return m.GetPathByFileIdFunc(fileParent)
}
