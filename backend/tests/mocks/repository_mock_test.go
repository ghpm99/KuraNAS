package mocks

import (
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

func TestMockRepositoryMethods(t *testing.T) {
	expected := utils.PaginationResponse[files.FileModel]{
		Items: []files.FileModel{{ID: 1, Name: "a"}},
	}
	mock := &MockRepository{
		GetFilesFunc: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
			return expected, nil
		},
	}

	if db := mock.GetDbContext(); db != nil {
		t.Fatalf("expected nil db context")
	}

	result, err := mock.GetFiles(files.FileFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("GetFiles returned error: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].ID != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	created, err := mock.CreateFile(nil, files.FileModel{ID: 2})
	if err != nil {
		t.Fatalf("CreateFile returned error: %v", err)
	}
	if created.ID != 0 {
		t.Fatalf("expected zero-value CreateFile response")
	}

	updated, err := mock.UpdateFile(nil, files.FileModel{ID: 3})
	if err != nil {
		t.Fatalf("UpdateFile returned error: %v", err)
	}
	if updated {
		t.Fatalf("expected false from UpdateFile")
	}
}
