package files_test

import (
	"database/sql"
	"errors"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"testing"
)

type Service struct {
	Repository files.RepositoryInterface
	Tasks      chan utils.Task
}

type mockRepository struct {
	getPathByFileId func(fileParent int) (string, error)
	getFiles        func(filter files.FileFilter, pagination utils.Pagination) (utils.PaginationResponse[files.FileModel], error)
}

func (m *mockRepository) GetDbContext() *sql.DB {
	return nil
}

func (m *mockRepository) GetFiles(filter files.FileFilter, pagination utils.Pagination) (utils.PaginationResponse[files.FileModel], error) {
	return m.getFiles(filter, pagination)
}

func (m *mockRepository) GetFilesByPath(path string) ([]files.FileModel, error) {
	return nil, nil
}

func (m *mockRepository) GetFileByNameAndPath(name string, path string) (files.FileModel, error) {
	return files.FileModel{}, nil
}

func (m *mockRepository) CreateFile(transaction *sql.Tx, file files.FileModel) (files.FileModel, error) {
	return files.FileModel{}, nil
}

func (m *mockRepository) UpdateFile(transaction *sql.Tx, file files.FileModel) (bool, error) {
	return false, nil
}

func (m *mockRepository) GetPathByFileId(fileParent int) (string, error) {
	return m.getPathByFileId(fileParent)
}

func TestService_GetFiles(t *testing.T) {
	tests := []struct {
		name    string
		mock    *mockRepository
		args    files.FileFilter
		wantErr bool
	}{
		{
			name: "GetFiles with FileParent equals 0",
			mock: &mockRepository{
				getFiles: func(filter files.FileFilter, pagination utils.Pagination) (utils.PaginationResponse[files.FileModel], error) {
					return utils.PaginationResponse[files.FileModel]{}, nil
				},
			},
			args: files.FileFilter{
				FileParent: 0,
			},
			wantErr: false,
		},
		{
			name: "GetFiles with FileParent not equals 0 and GetPathByFileId returns error",
			mock: &mockRepository{
				getPathByFileId: func(fileParent int) (string, error) {
					return "", errors.New("GetPathByFileId error")
				},
			},
			args: files.FileFilter{
				FileParent: 1,
			},
			wantErr: true,
		},
		{
			name: "GetFiles with FileParent not equals 0 and GetFiles returns error",
			mock: &mockRepository{
				getPathByFileId: func(fileParent int) (string, error) {
					return "/path", nil
				},
				getFiles: func(filter files.FileFilter, pagination utils.Pagination) (utils.PaginationResponse[files.FileModel], error) {
					return utils.PaginationResponse[files.FileModel]{}, errors.New("GetFiles error")
				},
			},
			args: files.FileFilter{
				FileParent: 1,
			},
			wantErr: true,
		},
		{
			name: "GetFiles success",
			mock: &mockRepository{
				getPathByFileId: func(fileParent int) (string, error) {
					return "/path", nil
				},
				getFiles: func(filter files.FileFilter, pagination utils.Pagination) (utils.PaginationResponse[files.FileModel], error) {
					return utils.PaginationResponse[files.FileModel]{
						Items: []files.FileModel{
							{ID: 1, Name: "test_file.txt", Path: "/path"},
						},
					}, nil
				},
			},
			args: files.FileFilter{
				FileParent: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &files.Service{
				Repository: tt.mock, // Agora funciona porque Repository é uma interface
				Tasks:      make(chan utils.Task, 1),
			}

			fileDtoList := &utils.PaginationResponse[files.FileDto]{}
			err := service.GetFiles(tt.args, fileDtoList)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
