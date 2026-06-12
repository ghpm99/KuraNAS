package files

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"nas-go/api/internal/config"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolvePathInEntryPoint(t *testing.T) {
	tempDir := t.TempDir()
	config.AppConfig.EntryPoint = tempDir

	path, err := resolvePathInRoots("")
	if err != nil {
		t.Fatalf("expected entry point path, got error: %v", err)
	}
	if path != filepath.Clean(tempDir) {
		t.Fatalf("expected %s, got %s", filepath.Clean(tempDir), path)
	}

	relativePath := "docs/file.txt"
	resolvedRelative, err := resolvePathInRoots(relativePath)
	if err != nil {
		t.Fatalf("expected valid relative path, got error: %v", err)
	}
	expectedRelative := filepath.Join(filepath.Clean(tempDir), filepath.FromSlash(relativePath))
	if resolvedRelative != expectedRelative {
		t.Fatalf("expected %s, got %s", expectedRelative, resolvedRelative)
	}

	if _, err := resolvePathInRoots("../outside"); err == nil {
		t.Fatalf("expected error for path outside entry point")
	}
}

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

	resolvedPath, err := resolvePathInRoots(`\nested\file.txt`)
	if err != nil {
		t.Fatalf("resolvePathInRoots with backslashes returned error: %v", err)
	}
	expected := filepath.Join(entryPoint, "nested", "file.txt")
	if resolvedPath != expected {
		t.Fatalf("expected %q, got %q", expected, resolvedPath)
	}

	resolvedPath, err = resolvePathInRoots(`/\Documentos\Trabalho`)
	if err != nil {
		t.Fatalf("resolvePathInRoots with mixed separators returned error: %v", err)
	}
	expected = filepath.Join(entryPoint, "Documentos", "Trabalho")
	if resolvedPath != expected {
		t.Fatalf("expected %q, got %q", expected, resolvedPath)
	}
}

func TestResolvePathInEntryPointFromOperationsTest(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	resolvedPath, err := resolvePathInRoots("nested/file.txt")
	if err != nil {
		t.Fatalf("resolvePathInRoots returned error: %v", err)
	}
	if resolvedPath != filepath.Join(entryPoint, "nested", "file.txt") {
		t.Fatalf("resolvePathInRoots returned %q", resolvedPath)
	}

	if _, err := resolvePathInRoots("../outside"); err == nil {
		t.Fatalf("expected resolvePathInRoots outside-entrypoint error")
	}
}

func newTestServiceWithFileRecords(t *testing.T, entryPoint string, records []FileModel) *Service {
	t.Helper()
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			for _, r := range records {
				if r.ID == id {
					return r, true, nil
				}
			}
			return FileModel{}, false, nil
		},
	}
	service := newFilesServiceForTest(t, repo)
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

// trashBinRecorder fakes the trash domain: it records the call and moves the
// file aside, like the real bin, so assertions can check both sides.
type trashBinRecorder struct {
	calls []string
	sizes []int64
	dest  string
	err   error
}

func (b *trashBinRecorder) MoveToTrash(originalPath string, size int64) error {
	if b.err != nil {
		return b.err
	}
	b.calls = append(b.calls, originalPath)
	b.sizes = append(b.sizes, size)
	if b.dest != "" {
		return os.Rename(originalPath, b.dest)
	}
	return nil
}

func TestDeleteFileFromDiskMovesToTrashByDefault(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	fileToDelete := filepath.Join(entryPoint, "delete-me.txt")
	if err := os.WriteFile(fileToDelete, []byte("bye"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "delete-me.txt", Path: fileToDelete, Type: File, Size: 3},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)
	bin := &trashBinRecorder{dest: filepath.Join(t.TempDir(), "delete-me.txt.1")}
	service.SetTrashBin(bin)

	if err := service.DeleteFileFromDisk(1, false); err != nil {
		t.Fatalf("DeleteFileFromDisk returned error: %v", err)
	}
	if len(bin.calls) != 1 || bin.calls[0] != fileToDelete || bin.sizes[0] != 3 {
		t.Fatalf("expected one MoveToTrash(%q, 3), got %v %v", fileToDelete, bin.calls, bin.sizes)
	}
	if _, err := os.Stat(fileToDelete); !os.IsNotExist(err) {
		t.Fatalf("file must be gone from the original path")
	}
	if data, err := os.ReadFile(bin.dest); err != nil || string(data) != "bye" {
		t.Fatalf("bytes must survive in the bin, got %q err=%v", data, err)
	}

	// Invalid ID
	err := service.DeleteFileFromDisk(0, false)
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_INVALID_ID")

	// Non-existent ID
	err = service.DeleteFileFromDisk(999, false)
	requireOperationError(t, err, http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND")
}

func TestDeleteFileFromDiskPermanentRemovesBytes(t *testing.T) {
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
	bin := &trashBinRecorder{}
	service.SetTrashBin(bin)

	if err := service.DeleteFileFromDisk(1, true); err != nil {
		t.Fatalf("DeleteFileFromDisk returned error: %v", err)
	}
	if _, err := os.Stat(fileToDelete); !os.IsNotExist(err) {
		t.Fatalf("permanently deleted file still exists")
	}
	if len(bin.calls) != 0 {
		t.Fatalf("permanent delete must not touch the trash bin, got %v", bin.calls)
	}
}

func TestDeleteFileFromDiskWithoutTrashBinRefusesDefaultDelete(t *testing.T) {
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

	err := service.DeleteFileFromDisk(1, false)
	requireOperationError(t, err, http.StatusInternalServerError, "ERROR_DELETE_FAILED")
	if _, statErr := os.Stat(fileToDelete); statErr != nil {
		t.Fatalf("file must be untouched when the trash bin is missing: %v", statErr)
	}
}

func TestDeleteFileFromDiskForbidsEntryPoint(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	records := []FileModel{
		{ID: 1, Name: filepath.Base(entryPoint), Path: entryPoint, Type: Directory},
	}
	service := newTestServiceWithFileRecords(t, entryPoint, records)

	err := service.DeleteFileFromDisk(1, false)
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

func TestMoveDirectorySyncsRowAndDescendantsInOneTransaction(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	sourceDir := filepath.Join(entryPoint, "library")
	if err := os.MkdirAll(filepath.Join(sourceDir, "child"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	destDir := filepath.Join(entryPoint, "archive")
	if err := os.Mkdir(destDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "library", Path: sourceDir, ParentPath: entryPoint, Type: Directory},
		{ID: 2, Name: "archive", Path: destDir, ParentPath: entryPoint, Type: Directory},
	}

	var updated []FileModel
	var prefixSwaps [][2]string
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			for _, r := range records {
				if r.ID == id {
					return r, true, nil
				}
			}
			return FileModel{}, false, nil
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			updated = append(updated, file)
			return true, nil
		},
		updateDescendantPathsFn: func(transaction *sql.Tx, oldPath string, newPath string) (int64, error) {
			prefixSwaps = append(prefixSwaps, [2]string{oldPath, newPath})
			return 1, nil
		},
	}
	service := newFilesServiceForTest(t, repo)

	destID := 2
	movedPath, err := service.MoveFile(1, &destID, "")
	if err != nil {
		t.Fatalf("MoveFile returned error: %v", err)
	}

	if len(updated) != 1 {
		t.Fatalf("expected 1 updated row, got %d", len(updated))
	}
	row := updated[0]
	if row.ID != 1 || row.Path != movedPath || row.ParentPath != destDir || row.Name != "library" {
		t.Fatalf("moved row not synced: %+v", row)
	}
	if len(prefixSwaps) != 1 || prefixSwaps[0][0] != sourceDir || prefixSwaps[0][1] != movedPath {
		t.Fatalf("descendant prefix swap not applied: %+v", prefixSwaps)
	}
}

func TestRenameFileSyncsRowSynchronously(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	sourceFile := filepath.Join(entryPoint, "old.txt")
	if err := os.WriteFile(sourceFile, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "old.txt", Path: sourceFile, ParentPath: entryPoint, Type: File, Format: ".txt"},
	}

	var updated []FileModel
	var prefixSwaps int
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			if id == 1 {
				return records[0], true, nil
			}
			return FileModel{}, false, nil
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			updated = append(updated, file)
			return true, nil
		},
		updateDescendantPathsFn: func(transaction *sql.Tx, oldPath string, newPath string) (int64, error) {
			prefixSwaps++
			return 0, nil
		},
	}
	service := newFilesServiceForTest(t, repo)

	renamedPath, err := service.RenameFile(1, "new.md")
	if err != nil {
		t.Fatalf("RenameFile returned error: %v", err)
	}

	if len(updated) != 1 {
		t.Fatalf("expected 1 updated row, got %d", len(updated))
	}
	row := updated[0]
	if row.Name != "new.md" || row.Path != renamedPath || row.Format != ".md" {
		t.Fatalf("renamed row not synced: %+v", row)
	}
	if prefixSwaps != 0 {
		t.Fatalf("plain file rename must not touch descendants, got %d prefix swaps", prefixSwaps)
	}
}

func TestMoveFileSucceedsWhenDatabaseSyncFails(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	sourceFile := filepath.Join(entryPoint, "data.txt")
	if err := os.WriteFile(sourceFile, []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	destDir := filepath.Join(entryPoint, "dest")
	if err := os.Mkdir(destDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "data.txt", Path: sourceFile, ParentPath: entryPoint, Type: File},
		{ID: 2, Name: "dest", Path: destDir, ParentPath: entryPoint, Type: Directory},
	}
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			for _, r := range records {
				if r.ID == id {
					return r, true, nil
				}
			}
			return FileModel{}, false, nil
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			return false, errors.New("db down")
		},
	}
	service := newFilesServiceForTest(t, repo)

	destID := 2
	movedPath, err := service.MoveFile(1, &destID, "")
	if err != nil {
		t.Fatalf("MoveFile must succeed when only the db sync fails, got: %v", err)
	}
	if _, statErr := os.Stat(movedPath); statErr != nil {
		t.Fatalf("moved file missing on disk: %v", statErr)
	}
	if _, statErr := os.Stat(sourceFile); !os.IsNotExist(statErr) {
		t.Fatalf("source file still present, disk operation corrupted")
	}
}

func TestDeleteFileMarksSubtreeDeletedSynchronously(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	dirToDelete := filepath.Join(entryPoint, "library")
	if err := os.MkdirAll(filepath.Join(dirToDelete, "child"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "library", Path: dirToDelete, ParentPath: entryPoint, Type: Directory},
	}

	var markedPaths []string
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			if id == 1 {
				return records[0], true, nil
			}
			return FileModel{}, false, nil
		},
		markDeletedSubtreeFn: func(transaction *sql.Tx, path string, deletedAt time.Time) (int64, error) {
			markedPaths = append(markedPaths, path)
			if deletedAt.IsZero() {
				t.Fatalf("expected non-zero deleted_at timestamp")
			}
			return 2, nil
		},
	}
	service := newFilesServiceForTest(t, repo)

	if err := service.DeleteFileFromDisk(1, true); err != nil {
		t.Fatalf("DeleteFileFromDisk returned error: %v", err)
	}

	if len(markedPaths) != 1 || markedPaths[0] != dirToDelete {
		t.Fatalf("expected subtree soft-delete for %q, got %v", dirToDelete, markedPaths)
	}
}

func TestDeleteFileSucceedsWhenDatabaseSyncFails(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	fileToDelete := filepath.Join(entryPoint, "doomed.txt")
	if err := os.WriteFile(fileToDelete, []byte("bye"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	records := []FileModel{
		{ID: 1, Name: "doomed.txt", Path: fileToDelete, ParentPath: entryPoint, Type: File},
	}
	repo := &filesRepoMock{
		getFileByIDFn: func(id int) (FileModel, bool, error) {
			if id == 1 {
				return records[0], true, nil
			}
			return FileModel{}, false, nil
		},
		markDeletedSubtreeFn: func(transaction *sql.Tx, path string, deletedAt time.Time) (int64, error) {
			return 0, errors.New("db down")
		},
	}
	service := newFilesServiceForTest(t, repo)

	if err := service.DeleteFileFromDisk(1, true); err != nil {
		t.Fatalf("DeleteFileFromDisk must succeed when only the db sync fails, got: %v", err)
	}
	if _, statErr := os.Stat(fileToDelete); !os.IsNotExist(statErr) {
		t.Fatalf("deleted file still on disk")
	}
}

func TestCreateFolderInsertsDirectoryRowSynchronously(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	var created []FileModel
	repo := &filesRepoMock{
		createFileFn: func(transaction *sql.Tx, file FileModel) (FileModel, error) {
			file.ID = len(created) + 1
			created = append(created, file)
			return file, nil
		},
	}
	service := newFilesServiceForTest(t, repo)

	createdPath, err := service.CreateFolder(nil, "albums")
	if err != nil {
		t.Fatalf("CreateFolder returned error: %v", err)
	}

	if len(created) != 1 {
		t.Fatalf("expected 1 directory row inserted, got %d", len(created))
	}
	row := created[0]
	if row.Type != Directory || row.Name != "albums" || row.Path != createdPath || row.ParentPath != entryPoint {
		t.Fatalf("unexpected directory row: %+v", row)
	}
}

func TestCreateFolderRevivesSoftDeletedRow(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	deletedRow := FileModel{
		ID:         7,
		Name:       "albums",
		Path:       filepath.Join(entryPoint, "albums"),
		ParentPath: entryPoint,
		Type:       Directory,
		DeletedAt:  sql.NullTime{Valid: true, Time: time.Now()},
	}

	var updated []FileModel
	repo := &filesRepoMock{
		getFilesByNameAndPathFn: func(name string, path string, limit int) ([]FileModel, error) {
			if name == "albums" {
				return []FileModel{deletedRow}, nil
			}
			return nil, nil
		},
		createFileFn: func(transaction *sql.Tx, file FileModel) (FileModel, error) {
			t.Fatalf("expected revive via UpdateFile, got CreateFile for %+v", file)
			return file, nil
		},
		updateFileFn: func(transaction *sql.Tx, file FileModel) (bool, error) {
			updated = append(updated, file)
			return true, nil
		},
	}
	service := newFilesServiceForTest(t, repo)

	if _, err := service.CreateFolder(nil, "albums"); err != nil {
		t.Fatalf("CreateFolder returned error: %v", err)
	}

	if len(updated) != 1 {
		t.Fatalf("expected 1 updated row, got %d", len(updated))
	}
	if updated[0].ID != 7 || updated[0].DeletedAt.Valid {
		t.Fatalf("expected row 7 revived (deleted_at cleared), got %+v", updated[0])
	}
}

func TestCreateFolderSucceedsWhenDatabaseSyncFails(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	repo := &filesRepoMock{
		createFileFn: func(transaction *sql.Tx, file FileModel) (FileModel, error) {
			return file, errors.New("db down")
		},
	}
	service := newFilesServiceForTest(t, repo)

	createdPath, err := service.CreateFolder(nil, "albums")
	if err != nil {
		t.Fatalf("CreateFolder must succeed when only the db sync fails, got: %v", err)
	}
	if info, statErr := os.Stat(createdPath); statErr != nil || !info.IsDir() {
		t.Fatalf("created folder missing on disk: %v", statErr)
	}
}

func TestUploadAndCopyInsertBasicRowsSynchronously(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	var created []FileModel
	repo := &filesRepoMock{
		createFileFn: func(transaction *sql.Tx, file FileModel) (FileModel, error) {
			file.ID = len(created) + 1
			created = append(created, file)
			return file, nil
		},
	}
	service := newFilesServiceForTest(t, repo)
	service.JobsRepository = newFilesJobsRepoMockForTest(t)

	headers := buildMultipartFileHeaders(t, "files", map[string]string{"photo.jpg": "binary"})
	result, err := service.UploadFiles(0, headers)
	if err != nil {
		t.Fatalf("UploadFiles returned error: %v", err)
	}
	if len(created) != 1 || created[0].Path != result.Uploaded[0] || created[0].Type != File {
		t.Fatalf("expected uploaded file row inserted, got %+v", created)
	}

	sourceRecord := FileModel{ID: 50, Name: "photo.jpg", Path: result.Uploaded[0], ParentPath: entryPoint, Type: File, Format: ".jpg"}
	repo.getFileByIDFn = func(id int) (FileModel, bool, error) {
		if id == 50 {
			return sourceRecord, true, nil
		}
		return FileModel{}, false, nil
	}

	copiedPath, err := service.CopyFile(50, nil, "", "copy.jpg")
	if err != nil {
		t.Fatalf("CopyFile returned error: %v", err)
	}
	if len(created) != 2 || created[1].Path != copiedPath || created[1].Name != "copy.jpg" {
		t.Fatalf("expected copied file row inserted, got %+v", created)
	}
}

func TestCreateFolderEmptyName(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)

	service := newTestServiceWithFileRecords(t, entryPoint, nil)

	_, err := service.CreateFolder(nil, "")
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_FOLDER_NAME_REQUIRED")
}
