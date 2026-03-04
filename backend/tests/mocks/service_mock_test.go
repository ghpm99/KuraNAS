package mocks

import (
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

func TestMockServiceMethods(t *testing.T) {
	scanFilesCalled := false
	scanDirCalled := false

	mock := &MockService{
		GetFilesFunc: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{{ID: 1, Name: "file"}},
			}, nil
		},
		GetFileByNameAndPathFunc: func(name string, path string) (files.FileDto, error) {
			return files.FileDto{Name: name, Path: path}, nil
		},
		CreateFileFunc: func(fileDto files.FileDto) (files.FileDto, error) {
			return fileDto, nil
		},
		UpdateFileFunc: func(file files.FileDto) (bool, error) {
			return true, nil
		},
		ScanFilesTaskFunc: func(data string) {
			scanFilesCalled = data == "scan-files"
		},
		ScanDirTaskFunc: func(data string) {
			scanDirCalled = data == "scan-dir"
		},
	}

	filesResult, err := mock.GetFiles(files.FileFilter{}, 1, 10)
	if err != nil || len(filesResult.Items) != 1 {
		t.Fatalf("GetFiles failed, len=%d err=%v", len(filesResult.Items), err)
	}

	fileByName, err := mock.GetFileByNameAndPath("a", "/tmp")
	if err != nil || fileByName.Name != "a" {
		t.Fatalf("GetFileByNameAndPath failed, got=%+v err=%v", fileByName, err)
	}

	created, err := mock.CreateFile(files.FileDto{ID: 3})
	if err != nil || created.ID != 3 {
		t.Fatalf("CreateFile failed, got=%+v err=%v", created, err)
	}

	updated, err := mock.UpdateFile(files.FileDto{ID: 4})
	if err != nil || !updated {
		t.Fatalf("UpdateFile failed, updated=%v err=%v", updated, err)
	}

	mock.ScanFilesTask("scan-files")
	mock.ScanDirTask("scan-dir")
	if !scanFilesCalled || !scanDirCalled {
		t.Fatalf("expected scan task callbacks to be called")
	}
}
