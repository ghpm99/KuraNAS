package files

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/utils"
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

func TestResolvePathNormalizesBackslashes(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	resolvedPath, err := resolvePathInEntryPoint(`\nested\file.txt`)
	if err != nil {
		t.Fatalf("resolvePathInEntryPoint with backslashes returned error: %v", err)
	}
	expected := filepath.Join(entryPoint, "nested", "file.txt")
	if resolvedPath != expected {
		t.Fatalf("expected %q, got %q", expected, resolvedPath)
	}

	resolvedPath, err = resolvePathInEntryPoint(`/\Documentos\Trabalho`)
	if err != nil {
		t.Fatalf("resolvePathInEntryPoint with mixed separators returned error: %v", err)
	}
	expected = filepath.Join(entryPoint, "Documentos", "Trabalho")
	if resolvedPath != expected {
		t.Fatalf("expected %q, got %q", expected, resolvedPath)
	}
}

func TestResolvePathInEntryPointFromOperationsTest(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

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
}

func newTestServiceWithFileRecords(t *testing.T, entryPoint string, records []FileModel) *Service {
	t.Helper()
	repo := &filesRepoMock{
		getFilesFn: func(filter FileFilter, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
			if filter.ID.HasValue {
				for _, r := range records {
					if r.ID == filter.ID.Value {
						return utils.PaginationResponse[FileModel]{Items: []FileModel{r}}, nil
					}
				}
				return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
			}
			return utils.PaginationResponse[FileModel]{Items: []FileModel{}}, nil
		},
	}
	service := newFilesServiceForTest(t, repo, &metadataRepoMock{})
	service.JobsRepository = newFilesJobsRepoMockForTest(t)
	return service
}

func TestUploadFilesWithFolderID(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	subDir := filepath.Join(entryPoint, "albums")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	records := []FileModel{
		{ID: 10, Name: "albums", Path: subDir, Type: Directory},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	// Upload to root (ID 0)
	headers := buildMultipartFileHeaders(t, "files", map[string]string{"root.jpg": "binary"})
	result, err := service.UploadFiles(0, headers)
	if err != nil {
		t.Fatalf("UploadFiles to root returned error: %v", err)
	}
	if result.JobID <= 0 || len(result.Uploaded) != 1 {
		t.Fatalf("UploadFiles to root returned %+v", result)
	}

	// Upload to folder ID 10
	headers = buildMultipartFileHeaders(t, "files", map[string]string{"photo.jpg": "binary"})
	result, err = service.UploadFiles(10, headers)
	if err != nil {
		t.Fatalf("UploadFiles to folder ID returned error: %v", err)
	}
	if _, err := os.Stat(result.Uploaded[0]); err != nil {
		t.Fatalf("uploaded file missing: %v", err)
	}

	// Upload empty payload
	_, err = service.UploadFiles(0, nil)
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_NO_FILES_UPLOADED")

	// Upload duplicate
	_, err = service.UploadFiles(10, buildMultipartFileHeaders(t, "files", map[string]string{"photo.jpg": "again"}))
	requireOperationError(t, err, http.StatusConflict, "ERROR_TARGET_ALREADY_EXISTS")

	// Upload to non-existent folder ID
	_, err = service.UploadFiles(999, buildMultipartFileHeaders(t, "files", map[string]string{"x.txt": "data"}))
	requireOperationError(t, err, http.StatusNotFound, "ERROR_FOLDER_NOT_FOUND")
}

func TestCreateFolderWithParentID(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	subDir := filepath.Join(entryPoint, "docs")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	records := []FileModel{
		{ID: 5, Name: "docs", Path: subDir, Type: Directory},
		{ID: 6, Name: "file.txt", Path: filepath.Join(subDir, "file.txt"), Type: File},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	// Create folder at root (nil parent)
	createdPath, err := service.CreateFolder(nil, "albums")
	if err != nil {
		t.Fatalf("CreateFolder at root returned error: %v", err)
	}
	if info, err := os.Stat(createdPath); err != nil || !info.IsDir() {
		t.Fatalf("created folder missing: %v", err)
	}

	// Create folder under parent ID 5
	parentID := 5
	createdPath, err = service.CreateFolder(&parentID, "sub")
	if err != nil {
		t.Fatalf("CreateFolder under parent ID returned error: %v", err)
	}
	if info, err := os.Stat(createdPath); err != nil || !info.IsDir() {
		t.Fatalf("created subfolder missing: %v", err)
	}

	// Invalid folder name
	_, err = service.CreateFolder(nil, "../invalid")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_FOLDER_NAME_INVALID")

	// Parent ID pointing to a file instead of directory
	fileID := 6
	_, err = service.CreateFolder(&fileID, "nope")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_TARGET_NOT_DIRECTORY")
}

func TestMoveFileByID(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	sourceFile := filepath.Join(entryPoint, "source.txt")
	if err := os.WriteFile(sourceFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	destDir := filepath.Join(entryPoint, "target")
	if err := os.Mkdir(destDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "source.txt", Path: sourceFile, Type: File},
		{ID: 2, Name: "target", Path: destDir, Type: Directory},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	// Move to existing folder by ID
	destFolderID := 2
	movedPath, err := service.MoveFile(1, &destFolderID, "")
	if err != nil {
		t.Fatalf("MoveFile returned error: %v", err)
	}
	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}

	// Invalid source ID
	_, err = service.MoveFile(0, nil, "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_INVALID_ID")

	// Non-existent source ID
	_, err = service.MoveFile(999, nil, "")
	requireOperationError(t, err, http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND")
}

func TestMoveFileToNewPath(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	sourceFile := filepath.Join(entryPoint, "data.txt")
	if err := os.WriteFile(sourceFile, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "data.txt", Path: sourceFile, Type: File},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	// Move to new path (creates folder chain)
	movedPath, err := service.MoveFile(1, nil, "new-folder")
	if err != nil {
		t.Fatalf("MoveFile to new path returned error: %v", err)
	}
	if _, err := os.Stat(movedPath); err != nil {
		t.Fatalf("moved file missing at new path: %v", err)
	}
}

func TestMoveDirectoryIntoItself(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	nestedDir := filepath.Join(entryPoint, "library")
	childDir := filepath.Join(nestedDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "library", Path: nestedDir, Type: Directory},
		{ID: 2, Name: "child", Path: childDir, Type: Directory},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	destID := 2
	_, err := service.MoveFile(1, &destID, "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_CANNOT_MOVE_INTO_ITSELF")
}

func TestRenameFileByID(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	sourceFile := filepath.Join(entryPoint, "old.txt")
	if err := os.WriteFile(sourceFile, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "old.txt", Path: sourceFile, Type: File},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	renamedPath, err := service.RenameFile(1, "new.txt")
	if err != nil {
		t.Fatalf("RenameFile returned error: %v", err)
	}
	if filepath.Base(renamedPath) != "new.txt" {
		t.Fatalf("RenameFile returned %q", renamedPath)
	}

	// Invalid ID
	_, err = service.RenameFile(0, "x.txt")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_INVALID_ID")

	// Empty name
	_, err = service.RenameFile(1, "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_RENAME_NAME_REQUIRED")

	// Invalid name with path separator
	_, err = service.RenameFile(1, "../bad")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_RENAME_NAME_INVALID")

	// Non-existent ID
	_, err = service.RenameFile(999, "whatever.txt")
	requireOperationError(t, err, http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND")
}

func TestCopyFileByID(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	sourceFile := filepath.Join(entryPoint, "original.txt")
	if err := os.WriteFile(sourceFile, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	destDir := filepath.Join(entryPoint, "copies")
	if err := os.Mkdir(destDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "original.txt", Path: sourceFile, Type: File},
		{ID: 2, Name: "copies", Path: destDir, Type: Directory},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	// Copy to existing folder by ID (default name)
	destFolderID := 2
	copiedPath, err := service.CopyFile(1, &destFolderID, "", "")
	if err != nil {
		t.Fatalf("CopyFile returned error: %v", err)
	}
	if _, err := os.Stat(copiedPath); err != nil {
		t.Fatalf("copied file missing: %v", err)
	}

	// Copy with custom name
	copiedPath, err = service.CopyFile(1, &destFolderID, "", "renamed_copy.txt")
	if err != nil {
		t.Fatalf("CopyFile with custom name returned error: %v", err)
	}
	if filepath.Base(copiedPath) != "renamed_copy.txt" {
		t.Fatalf("CopyFile custom name returned %q", copiedPath)
	}

	// Copy to new folder path
	copiedPath, err = service.CopyFile(1, nil, "new-copies", "")
	if err != nil {
		t.Fatalf("CopyFile to new path returned error: %v", err)
	}
	if _, err := os.Stat(copiedPath); err != nil {
		t.Fatalf("copied file at new path missing: %v", err)
	}

	// Invalid source ID
	_, err = service.CopyFile(0, nil, "", "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_INVALID_ID")

	// Non-existent source ID
	_, err = service.CopyFile(999, nil, "", "")
	requireOperationError(t, err, http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND")
}

func TestCopyDirectoryIntoItself(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	nestedDir := filepath.Join(entryPoint, "library")
	childDir := filepath.Join(nestedDir, "child")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "library", Path: nestedDir, Type: Directory},
		{ID: 2, Name: "child", Path: childDir, Type: Directory},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	destID := 2
	_, err := service.CopyFile(1, &destID, "", "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_CANNOT_COPY_INTO_ITSELF")
}

func TestDeleteFileFromDiskByID(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	fileToDelete := filepath.Join(entryPoint, "delete-me.txt")
	if err := os.WriteFile(fileToDelete, []byte("bye"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "delete-me.txt", Path: fileToDelete, Type: File},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	if err := service.DeleteFileFromDisk(1); err != nil {
		t.Fatalf("DeleteFileFromDisk returned error: %v", err)
	}
	if _, err := os.Stat(fileToDelete); !os.IsNotExist(err) {
		t.Fatalf("deleted file still exists")
	}

	// Invalid ID
	err := service.DeleteFileFromDisk(0)
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_INVALID_ID")

	// Non-existent ID
	err = service.DeleteFileFromDisk(999)
	requireOperationError(t, err, http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND")
}

func TestDeleteFileFromDiskForbidsEntryPoint(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	records := []FileModel{
		{ID: 1, Name: filepath.Base(entryPoint), Path: entryPoint, Type: Directory},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	err := service.DeleteFileFromDisk(1)
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_DELETE_ENTRYPOINT_FORBIDDEN")
}

func TestResolveTargetFolderRoot(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	service := newTestServiceWithFileRecords(t, entryPoint, nil)

	// nil ID + empty path = root
	path, err := service.resolveTargetFolder(nil, "")
	if err != nil {
		t.Fatalf("resolveTargetFolder root returned error: %v", err)
	}
	if path != entryPoint {
		t.Fatalf("expected %q, got %q", entryPoint, path)
	}
}

func TestResolveTargetFolderCreatesPath(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	service := newTestServiceWithFileRecords(t, entryPoint, nil)

	// nil ID + relative path = creates and returns
	path, err := service.resolveTargetFolder(nil, "a/b/c")
	if err != nil {
		t.Fatalf("resolveTargetFolder with path returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("created folder chain missing: %v", err)
	}
}

func TestResolveTargetFolderNonExistentID(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	service := newTestServiceWithFileRecords(t, entryPoint, nil)

	id := 999
	_, err := service.resolveTargetFolder(&id, "")
	requireOperationError(t, err, http.StatusNotFound, "ERROR_FOLDER_NOT_FOUND")
}

func TestResolveTargetFolderFileNotDirectory(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	records := []FileModel{
		{ID: 1, Name: "file.txt", Path: filepath.Join(entryPoint, "file.txt"), Type: File},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	id := 1
	_, err := service.resolveTargetFolder(&id, "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_TARGET_NOT_DIRECTORY")
}

func TestUploadFilesInvalidFileName(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	service := newTestServiceWithFileRecords(t, entryPoint, nil)

	// Build a header with a path-separator-only filename
	headers := buildMultipartFileHeaders(t, "files", map[string]string{".": "data"})
	_, err := service.UploadFiles(0, headers)
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_FILE_NAME_INVALID")
}

func TestRenameFileConflict(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	file1 := filepath.Join(entryPoint, "a.txt")
	file2 := filepath.Join(entryPoint, "b.txt")
	if err := os.WriteFile(file1, []byte("a"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.WriteFile(file2, []byte("b"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "a.txt", Path: file1, Type: File},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	_, err := service.RenameFile(1, "b.txt")
	requireOperationError(t, err, http.StatusConflict, "ERROR_TARGET_ALREADY_EXISTS")
}

func TestMoveFileConflict(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	sourceFile := filepath.Join(entryPoint, "src.txt")
	if err := os.WriteFile(sourceFile, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Create dest dir with existing file of same name
	destDir := filepath.Join(entryPoint, "dest")
	if err := os.Mkdir(destDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "src.txt"), []byte("existing"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "src.txt", Path: sourceFile, Type: File},
		{ID: 2, Name: "dest", Path: destDir, Type: Directory},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	destID := 2
	_, err := service.MoveFile(1, &destID, "")
	requireOperationError(t, err, http.StatusConflict, "ERROR_TARGET_ALREADY_EXISTS")
}

func TestCopyFileConflict(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	sourceFile := filepath.Join(entryPoint, "src.txt")
	if err := os.WriteFile(sourceFile, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	destDir := filepath.Join(entryPoint, "dest")
	if err := os.Mkdir(destDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "src.txt"), []byte("existing"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "src.txt", Path: sourceFile, Type: File},
		{ID: 2, Name: "dest", Path: destDir, Type: Directory},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	destID := 2
	_, err := service.CopyFile(1, &destID, "", "")
	requireOperationError(t, err, http.StatusConflict, "ERROR_TARGET_ALREADY_EXISTS")
}

func TestCreateFolderAlreadyExists(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	if err := os.Mkdir(filepath.Join(entryPoint, "existing"), 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	service := newTestServiceWithFileRecords(t, entryPoint, nil)

	_, err := service.CreateFolder(nil, "existing")
	requireOperationError(t, err, http.StatusConflict, "ERROR_FOLDER_ALREADY_EXISTS")
}

func TestCreateFolderEmptyName(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	service := newTestServiceWithFileRecords(t, entryPoint, nil)

	_, err := service.CreateFolder(nil, "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_FOLDER_NAME_REQUIRED")
}
