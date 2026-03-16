package files

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"nas-go/api/internal/config"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setEntryPointForTest(t *testing.T, entryPoint string) {
	t.Helper()
	previous := config.AppConfig.EntryPoint
	config.AppConfig.EntryPoint = entryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = previous
	})
}

func requireOperationError(t *testing.T, err error, statusCode int, messageKey string) {
	t.Helper()
	var operationErr *FileOperationError
	if !errors.As(err, &operationErr) {
		t.Fatalf("expected FileOperationError, got %v", err)
	}
	if operationErr.StatusCode != statusCode || operationErr.MessageKey != messageKey {
		t.Fatalf("unexpected operation error %+v", operationErr)
	}
}

func buildMultipartFileHeaders(t *testing.T, fieldName string, files map[string]string) []*multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for name, content := range files {
		part, err := writer.CreateFormFile(fieldName, name)
		if err != nil {
			t.Fatalf("CreateFormFile failed: %v", err)
		}
		if _, err := io.WriteString(part, content); err != nil {
			t.Fatalf("WriteString failed: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("ParseMultipartForm failed: %v", err)
	}

	return request.MultipartForm.File[fieldName]
}

func TestFileOperationsResolveUploadAndCreateFolder(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	service := newFilesServiceForTest(t, &filesRepoMock{}, &metadataRepoMock{})
	service.JobsRepository = newFilesJobsRepoMockForTest(t)

	resolvedPath, err := resolvePathInEntryPoint("nested/file.txt")
	if err != nil {
		t.Fatalf("resolvePathInEntryPoint returned error: %v", err)
	}
	if resolvedPath != filepath.Join(entryPoint, "nested", "file.txt") {
		t.Fatalf("resolvePathInEntryPoint returned %q", resolvedPath)
	}

	if _, err := resolvePathInEntryPoint("../outside"); err == nil {
		t.Fatalf("expected resolvePathInEntryPoint outside-entrypoint error")
	}

	headers := buildMultipartFileHeaders(t, "files", map[string]string{"photo.jpg": "binary"})
	result, err := service.UploadFiles(entryPoint, headers)
	if err != nil {
		t.Fatalf("UploadFiles returned error: %v", err)
	}
	if result.JobID <= 0 || len(result.Uploaded) != 1 {
		t.Fatalf("UploadFiles returned %+v", result)
	}
	if _, err := os.Stat(result.Uploaded[0]); err != nil {
		t.Fatalf("uploaded file missing: %v", err)
	}

	_, err = service.UploadFiles(entryPoint, nil)
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_NO_FILES_UPLOADED")

	_, err = service.UploadFiles(entryPoint, buildMultipartFileHeaders(t, "files", map[string]string{"photo.jpg": "again"}))
	requireOperationError(t, err, http.StatusConflict, "ERROR_TARGET_ALREADY_EXISTS")

	createdPath, err := service.CreateFolder(entryPoint, "albums")
	if err != nil {
		t.Fatalf("CreateFolder returned error: %v", err)
	}
	if info, err := os.Stat(createdPath); err != nil || !info.IsDir() {
		t.Fatalf("created folder missing: %v", err)
	}

	_, err = service.CreateFolder(entryPoint, "../invalid")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_FOLDER_NAME_INVALID")
}

func TestFileOperationsMoveRenameCopyAndDelete(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	service := newFilesServiceForTest(t, &filesRepoMock{}, &metadataRepoMock{})

	sourceFile := filepath.Join(entryPoint, "source.txt")
	if err := os.WriteFile(sourceFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile source failed: %v", err)
	}

	movedPath, err := service.MovePath(sourceFile, filepath.Join(entryPoint, "moved.txt"))
	if err != nil {
		t.Fatalf("MovePath returned error: %v", err)
	}
	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}

	renamedPath, err := service.RenamePath(movedPath, "renamed.txt")
	if err != nil {
		t.Fatalf("RenamePath returned error: %v", err)
	}
	if filepath.Base(renamedPath) != "renamed.txt" {
		t.Fatalf("RenamePath returned %q", renamedPath)
	}

	copiedPath, err := service.CopyPath(renamedPath, filepath.Join(entryPoint, "copied.txt"))
	if err != nil {
		t.Fatalf("CopyPath returned error: %v", err)
	}
	if _, err := os.Stat(copiedPath); err != nil {
		t.Fatalf("copied file missing: %v", err)
	}

	nestedDir := filepath.Join(entryPoint, "library")
	if err := os.MkdirAll(filepath.Join(nestedDir, "child"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	_, err = service.MovePath(nestedDir, filepath.Join(nestedDir, "child", "library"))
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_CANNOT_MOVE_INTO_ITSELF")

	_, err = service.CopyPath(nestedDir, filepath.Join(nestedDir, "child", "library-copy"))
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_CANNOT_COPY_INTO_ITSELF")

	_, err = service.RenamePath("", "noop")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_RENAME_SOURCE_REQUIRED")

	if err := service.DeletePath(""); err == nil {
		t.Fatalf("expected DeletePath empty-path error")
	} else {
		requireOperationError(t, err, http.StatusBadRequest, "ERROR_DELETE_PATH_REQUIRED")
	}

	if err := service.DeletePath(entryPoint); err == nil {
		t.Fatalf("expected DeletePath entrypoint error")
	} else {
		requireOperationError(t, err, http.StatusBadRequest, "ERROR_DELETE_ENTRYPOINT_FORBIDDEN")
	}

	if err := service.DeletePath(copiedPath); err != nil {
		t.Fatalf("DeletePath returned error: %v", err)
	}
	if _, err := os.Stat(copiedPath); !os.IsNotExist(err) {
		t.Fatalf("copied path still exists: %v", err)
	}
}
