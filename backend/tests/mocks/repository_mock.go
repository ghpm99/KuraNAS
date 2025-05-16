package mocks

import (
	"database/sql"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

type MockRepository struct {
	GetDbContextFunc *sql.DB
	GetFilesFunc     func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error)
	CreateFileFunc   func(transaction *sql.Tx, file files.FileModel) (files.FileModel, error)
	UpdateFileFunc   func(transaction *sql.Tx, file files.FileModel) (bool, error)
}

func (m *MockRepository) GetDbContext() *sql.DB {
	return nil
}

func (m *MockRepository) GetFiles(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
	return m.GetFilesFunc(filter, page, pageSize)
}

func (m *MockRepository) CreateFile(transaction *sql.Tx, file files.FileModel) (files.FileModel, error) {
	return files.FileModel{}, nil
}

func (m *MockRepository) UpdateFile(transaction *sql.Tx, file files.FileModel) (bool, error) {
	return false, nil
}
