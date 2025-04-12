package files_test

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"

	"reflect"
	"testing"
)

func TestRepository_GetFiles(t *testing.T) {
	type args struct {
		filter     files.FileFilter
		pagination utils.Pagination
	}
	tests := []struct {
		name    string
		args    args
		mock    func(sqlmock.Sqlmock)
		want    utils.PaginationResponse[files.FileModel]
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				filter: files.FileFilter{
					Path: "/test",
				},
				pagination: utils.Pagination{
					PageSize: 10,
					Page:     1,
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "path", "format", "size", "updated_at", "created_at", "last_interaction", "last_backup", "type", "check_sum", "deleted_at"}).
					AddRow(1, "test.txt", "/test", "txt", 1024, "2023-10-26 00:00:00", "2023-10-26 00:00:00", "2023-10-26 00:00:00", "2023-10-26 00:00:00", "file", "checksum", nil)

				mock.ExpectQuery("SELECT id, name, path, format, size, updated_at, created_at, last_interaction, last_backup, type, check_sum, deleted_at FROM files WHERE path = \\? LIMIT \\? OFFSET \\?").
					WithArgs("/test", 11, 1).
					WillReturnRows(rows)
			},
			want: utils.PaginationResponse[files.FileModel]{
				Items: []files.FileModel{
					{
						ID:              1,
						Name:            "test.txt",
						Path:            "/test",
						Format:          "txt",
						Size:            1024,
						UpdatedAt:       "2023-10-26 00:00:00",
						CreatedAt:       "2023-10-26 00:00:00",
						LastInteraction: "2023-10-26 00:00:00",
						LastBackup:      "2023-10-26 00:00:00",
						Type:            "file",
						CheckSum:        "checksum",
						DeletedAt:       nil,
					},
				},
				Pagination: utils.Pagination{
					PageSize: 10,
					Page:     1,
				},
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				filter: files.FileFilter{
					Path: "/test",
				},
				pagination: utils.Pagination{
					PageSize: 10,
					Page:     1,
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, name, path, format, size, updated_at, created_at, last_interaction, last_backup, type, check_sum, deleted_at FROM files WHERE path = \\? LIMIT \\? OFFSET \\?").
					WithArgs("/test", 11, 1).
					WillReturnError(fmt.Errorf("db error"))
			},
			want: utils.PaginationResponse[files.FileModel]{
				Items: []files.FileModel{},
				Pagination: utils.Pagination{
					PageSize: 10,
					Page:     1,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.mock(mock)

			r := &files.Repository{
				dbContext: db,
			}

			got, err := r.GetFiles(tt.args.filter, tt.args.pagination)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repository.GetFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Items, tt.want.Items) {
				t.Errorf("Repository.GetFiles() = %v, want %v", got.Items, tt.want.Items)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
