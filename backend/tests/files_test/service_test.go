package files_test

import (
	"errors"
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
	"nas-go/api/tests/mocks"
	"testing"
)

type Service struct {
	Repository files.RepositoryInterface
	Tasks      chan utils.Task
}

func TestService_GetFiles(t *testing.T) {
	tests := []struct {
		name    string
		mock    *mocks.MockRepository
		args    files.FileFilter
		wantErr bool
	}{
		{
			name: "GetFiles with FileParent equals 0",
			mock: &mocks.MockRepository{
				GetFilesFunc: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
					return utils.PaginationResponse[files.FileModel]{}, nil
				},
			},
			args: files.FileFilter{
				FileParent: utils.Optional[int]{
					HasValue: false,
				},
			},
			wantErr: false,
		},
		{
			name: "GetFiles with FileParent not equals 0 and GetPathByFileId returns error",
			mock: &mocks.MockRepository{
				GetPathByFileIdFunc: func(fileParent int) (string, error) {
					return "", errors.New("GetPathByFileId error")
				},
			},
			args: files.FileFilter{
				FileParent: utils.Optional[int]{
					HasValue: true,
					Value:    1,
				},
			},
			wantErr: true,
		},
		{
			name: "GetFiles with FileParent not equals 0 and GetFiles returns error",
			mock: &mocks.MockRepository{
				GetPathByFileIdFunc: func(fileParent int) (string, error) {
					return "/path", nil
				},
				GetFilesFunc: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
					return utils.PaginationResponse[files.FileModel]{}, errors.New("GetFiles error")
				},
			},
			args: files.FileFilter{
				FileParent: utils.Optional[int]{
					HasValue: true,
					Value:    1,
				},
			},
			wantErr: true,
		},
		{
			name: "GetFiles success",
			mock: &mocks.MockRepository{
				GetPathByFileIdFunc: func(fileParent int) (string, error) {
					return "/path", nil
				},
				GetFilesFunc: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
					return utils.PaginationResponse[files.FileModel]{
						Items: []files.FileModel{
							{ID: 1, Name: "test_file.txt", Path: "/path"},
						},
					}, nil
				},
			},
			args: files.FileFilter{
				FileParent: utils.Optional[int]{
					HasValue: true,
					Value:    1,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &files.Service{
				Repository: tt.mock,
				Tasks:      make(chan utils.Task, 1),
			}

			pagination, err := service.GetFiles(tt.args, 1, 10)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, file := range pagination.Items {
				fmt.Println(file)
			}
		})
	}
}
