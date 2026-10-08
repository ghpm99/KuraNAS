package files

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func newUploadTestService(t *testing.T, records []FileModel) (*Service, string) {
	t.Helper()
	entryPoint := t.TempDir()
	setEntryPointForTest(t, entryPoint)
	return newTestServiceWithFileRecords(t, entryPoint, records), entryPoint
}

func uploadSingle(t *testing.T, service *Service, name string, content string, options UploadOptions) (UploadFilesResult, error) {
	t.Helper()
	headers := buildMultipartFileHeaders(t, "files", map[string]string{name: content})
	return service.UploadFilesWithOptions(0, headers, options)
}

func writeExistingFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}

func requireFileContent(t *testing.T, path string, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != expected {
		t.Fatalf("content of %s = %q, expected %q", path, content, expected)
	}
}

func TestUploadDefaultPolicyKeepsConflictAbort(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	writeExistingFile(t, filepath.Join(entryPoint, "a.txt"), "old")

	_, err := uploadSingle(t, service, "a.txt", "new", UploadOptions{})
	requireOperationError(t, err, http.StatusConflict, "ERROR_TARGET_ALREADY_EXISTS")
	requireFileContent(t, filepath.Join(entryPoint, "a.txt"), "old")
}

func TestUploadPolicyFailExplicitAbortsOnConflict(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	writeExistingFile(t, filepath.Join(entryPoint, "a.txt"), "old")

	_, err := uploadSingle(t, service, "a.txt", "new", UploadOptions{OnConflict: UploadConflictFail})
	requireOperationError(t, err, http.StatusConflict, "ERROR_TARGET_ALREADY_EXISTS")
}

func TestUploadPolicyInvalidIsRejected(t *testing.T) {
	service, _ := newUploadTestService(t, nil)

	_, err := uploadSingle(t, service, "a.txt", "new", UploadOptions{OnConflict: "overwrite"})
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_UPLOAD_CONFLICT_POLICY_INVALID")
}

func TestUploadPolicySkipLeavesExistingFile(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	writeExistingFile(t, filepath.Join(entryPoint, "a.txt"), "old")

	result, err := uploadSingle(t, service, "a.txt", "new", UploadOptions{OnConflict: UploadConflictSkip})
	if err != nil {
		t.Fatalf("UploadFilesWithOptions returned error: %v", err)
	}
	if len(result.Files) != 1 || result.Files[0].Status != UploadStatusSkipped || result.Files[0].Name != "a.txt" {
		t.Fatalf("unexpected results %+v", result.Files)
	}
	if len(result.Uploaded) != 0 || result.JobID != 0 {
		t.Fatalf("skip must not create a processing job: %+v", result)
	}
	requireFileContent(t, filepath.Join(entryPoint, "a.txt"), "old")
}

func TestUploadPolicyReplaceOverwritesContent(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	writeExistingFile(t, filepath.Join(entryPoint, "a.txt"), "old")

	result, err := uploadSingle(t, service, "a.txt", "new", UploadOptions{OnConflict: UploadConflictReplace})
	if err != nil {
		t.Fatalf("UploadFilesWithOptions returned error: %v", err)
	}
	if result.Files[0].Status != UploadStatusReplaced || len(result.Uploaded) != 1 || result.JobID <= 0 {
		t.Fatalf("unexpected result %+v", result)
	}
	requireFileContent(t, filepath.Join(entryPoint, "a.txt"), "new")
}

func TestUploadPolicyReplaceOnDirectoryFailsPerFile(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	if err := os.Mkdir(filepath.Join(entryPoint, "a.txt"), 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	result, err := uploadSingle(t, service, "a.txt", "new", UploadOptions{OnConflict: UploadConflictReplace})
	if err != nil {
		t.Fatalf("batch must not abort: %v", err)
	}
	if result.Files[0].Status != UploadStatusFailed || result.Files[0].Error == "" {
		t.Fatalf("expected failed result, got %+v", result.Files[0])
	}
}

func TestUploadPolicyReplaceOnColdFileFailsPerFile(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	coldPath := filepath.Join(entryPoint, "cold.txt")
	repo := service.Repository.(*filesRepoMock)
	repo.getFilesByNameAndPathFn = func(name string, path string, limit int) ([]FileModel, error) {
		if path != coldPath {
			return nil, nil
		}
		return []FileModel{{ID: 7, Name: name, Path: path, Type: File, PhysicalPath: sql.NullString{String: "/cold/cold.txt", Valid: true}}}, nil
	}

	result, err := uploadSingle(t, service, "cold.txt", "new", UploadOptions{OnConflict: UploadConflictReplace})
	if err != nil {
		t.Fatalf("batch must not abort: %v", err)
	}
	if result.Files[0].Status != UploadStatusFailed {
		t.Fatalf("expected failed result, got %+v", result.Files[0])
	}
	if _, statErr := os.Stat(coldPath); !os.IsNotExist(statErr) {
		t.Fatalf("cold logical path must stay untouched")
	}
}

func TestUploadPolicyRenameAddsSequentialSuffix(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	writeExistingFile(t, filepath.Join(entryPoint, "a.txt"), "old")
	writeExistingFile(t, filepath.Join(entryPoint, "a (2).txt"), "old2")

	result, err := uploadSingle(t, service, "a.txt", "new", UploadOptions{OnConflict: UploadConflictRename})
	if err != nil {
		t.Fatalf("UploadFilesWithOptions returned error: %v", err)
	}
	if result.Files[0].Status != UploadStatusRenamed || result.Files[0].Name != "a (3).txt" {
		t.Fatalf("unexpected result %+v", result.Files[0])
	}
	requireFileContent(t, filepath.Join(entryPoint, "a (3).txt"), "new")
	requireFileContent(t, filepath.Join(entryPoint, "a.txt"), "old")
}

func TestUploadPolicyRenameWithoutConflictIsPlainUpload(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)

	result, err := uploadSingle(t, service, "a.txt", "new", UploadOptions{OnConflict: UploadConflictRename})
	if err != nil {
		t.Fatalf("UploadFilesWithOptions returned error: %v", err)
	}
	if result.Files[0].Status != UploadStatusUploaded {
		t.Fatalf("unexpected result %+v", result.Files[0])
	}
	requireFileContent(t, filepath.Join(entryPoint, "a.txt"), "new")
}

func TestUploadBatchContinuesAfterConflictWithNonFailPolicy(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	writeExistingFile(t, filepath.Join(entryPoint, "a.txt"), "old")

	headers := buildMultipartFileHeaders(t, "files", map[string]string{"a.txt": "new", "b.txt": "fresh"})
	result, err := service.UploadFilesWithOptions(0, headers, UploadOptions{OnConflict: UploadConflictSkip})
	if err != nil {
		t.Fatalf("UploadFilesWithOptions returned error: %v", err)
	}
	statuses := map[string]string{}
	for _, fileResult := range result.Files {
		statuses[fileResult.Name] = fileResult.Status
	}
	if statuses["a.txt"] != UploadStatusSkipped || statuses["b.txt"] != UploadStatusUploaded {
		t.Fatalf("unexpected statuses %+v", statuses)
	}
}

func TestUploadRelativePathCreatesIntermediateFolders(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)

	result, err := uploadSingle(t, service, "a.mp3", "song", UploadOptions{RelativePaths: []string{"Album/Disc1/a.mp3"}})
	if err != nil {
		t.Fatalf("UploadFilesWithOptions returned error: %v", err)
	}
	expectedPath := filepath.Join(entryPoint, "Album", "Disc1", "a.mp3")
	requireFileContent(t, expectedPath, "song")
	if result.Files[0].Path != expectedPath || result.Files[0].Status != UploadStatusUploaded {
		t.Fatalf("unexpected result %+v", result.Files[0])
	}
}

func TestUploadRelativePathReusesExistingFolders(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	if err := os.MkdirAll(filepath.Join(entryPoint, "Album"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	_, err := uploadSingle(t, service, "a.mp3", "song", UploadOptions{RelativePaths: []string{"Album\\a.mp3"}})
	if err != nil {
		t.Fatalf("UploadFilesWithOptions returned error: %v", err)
	}
	requireFileContent(t, filepath.Join(entryPoint, "Album", "a.mp3"), "song")
}

func TestUploadRelativePathFolderBlockedByFileFailsPerFile(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)
	writeExistingFile(t, filepath.Join(entryPoint, "Album"), "i am a file")

	result, err := uploadSingle(t, service, "a.mp3", "song", UploadOptions{
		OnConflict:    UploadConflictRename,
		RelativePaths: []string{"Album/a.mp3"},
	})
	if err != nil {
		t.Fatalf("batch must not abort: %v", err)
	}
	if result.Files[0].Status != UploadStatusFailed {
		t.Fatalf("unexpected result %+v", result.Files[0])
	}
}

func TestUploadRelativePathRejectsTraversalAndAbsolute(t *testing.T) {
	invalidPaths := []string{"../evil/a.txt", "a/../../b.txt", "/etc/a.txt", `\\server\a.txt`, "C:/a/b.txt", "a/..\\b.txt"}
	for _, invalidPath := range invalidPaths {
		service, entryPoint := newUploadTestService(t, nil)
		_, err := uploadSingle(t, service, "a.txt", "x", UploadOptions{RelativePaths: []string{invalidPath}})
		requireOperationError(t, err, http.StatusBadRequest, "ERROR_UPLOAD_RELATIVE_PATH_INVALID")
		entries, _ := os.ReadDir(entryPoint)
		if len(entries) != 0 {
			t.Fatalf("nothing may be written for %q, found %d entries", invalidPath, len(entries))
		}
	}
}

func TestUploadRelativePathsCountMismatch(t *testing.T) {
	service, _ := newUploadTestService(t, nil)

	_, err := uploadSingle(t, service, "a.txt", "x", UploadOptions{RelativePaths: []string{"a/a.txt", "b/b.txt"}})
	requireOperationError(t, err, http.StatusBadRequest, "ERROR_UPLOAD_RELATIVE_PATHS_MISMATCH")
}

func TestUploadRelativePathEmptyEntryUploadsToTarget(t *testing.T) {
	service, entryPoint := newUploadTestService(t, nil)

	_, err := uploadSingle(t, service, "a.txt", "x", UploadOptions{RelativePaths: []string{""}})
	if err != nil {
		t.Fatalf("UploadFilesWithOptions returned error: %v", err)
	}
	requireFileContent(t, filepath.Join(entryPoint, "a.txt"), "x")
}

func TestParseUploadDirectorySegments(t *testing.T) {
	segments, err := parseUploadDirectorySegments("./A//B/c.txt")
	if err != nil || len(segments) != 2 || segments[0] != "A" || segments[1] != "B" {
		t.Fatalf("unexpected segments %v, err %v", segments, err)
	}
	segments, err = parseUploadDirectorySegments("c.txt")
	if err != nil || len(segments) != 0 {
		t.Fatalf("unexpected segments %v, err %v", segments, err)
	}
}

type uploadCapturingService struct {
	filesHandlerServiceMock
	capturedFolderID int
	capturedHeaders  []*multipart.FileHeader
	capturedOptions  UploadOptions
}

func (m *uploadCapturingService) UploadFilesWithOptions(targetFolderID int, files []*multipart.FileHeader, options UploadOptions) (UploadFilesResult, error) {
	m.capturedFolderID = targetFolderID
	m.capturedHeaders = files
	m.capturedOptions = options
	return UploadFilesResult{
		Uploaded: []string{"/data/a.txt"},
		JobID:    5,
		Files:    []UploadFileResult{{Name: "a.txt", Path: "/data/a.txt", Status: UploadStatusUploaded}},
	}, nil
}

func postUploadForm(t *testing.T, router *gin.Engine, fields map[string][]string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, values := range fields {
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				t.Fatalf("WriteField failed: %v", err)
			}
		}
	}
	part, err := writer.CreateFormFile("files", "a.txt")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write([]byte("content")); err != nil {
		t.Fatalf("part write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close failed: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/files/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func newUploadHandlerRouter(service ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/files/upload", NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{}).UploadFilesHandler)
	return router
}

func TestUploadHandlerDecodesConflictAndRelativePathFields(t *testing.T) {
	service := &uploadCapturingService{}
	router := newUploadHandlerRouter(service)

	recorder := postUploadForm(t, router, map[string][]string{
		"target_folder_id": {"12"},
		"on_conflict":      {"replace"},
		"relative_paths":   {"Album/Disc1/a.txt"},
	})

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body %s", recorder.Code, recorder.Body.String())
	}
	if service.capturedFolderID != 12 || service.capturedOptions.OnConflict != UploadConflictReplace {
		t.Fatalf("unexpected capture %+v folder %d", service.capturedOptions, service.capturedFolderID)
	}
	if len(service.capturedOptions.RelativePaths) != 1 || service.capturedOptions.RelativePaths[0] != "Album/Disc1/a.txt" {
		t.Fatalf("unexpected relative paths %v", service.capturedOptions.RelativePaths)
	}
	if len(service.capturedHeaders) != 1 || service.capturedHeaders[0].Filename != "a.txt" {
		t.Fatalf("unexpected headers %+v", service.capturedHeaders)
	}
}

func TestUploadHandlerWithoutNewFieldsUsesDefaults(t *testing.T) {
	service := &uploadCapturingService{}
	router := newUploadHandlerRouter(service)

	recorder := postUploadForm(t, router, map[string][]string{"target_folder_id": {"3"}})

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d", recorder.Code)
	}
	if service.capturedOptions.OnConflict != "" || len(service.capturedOptions.RelativePaths) != 0 {
		t.Fatalf("expected empty options, got %+v", service.capturedOptions)
	}
}

func TestUploadHandlerResponseKeepsLegacyFieldsAndAddsFiles(t *testing.T) {
	router := newUploadHandlerRouter(&uploadCapturingService{})

	recorder := postUploadForm(t, router, nil)

	var body struct {
		Message  string             `json:"message"`
		Uploaded []string           `json:"uploaded"`
		JobID    int                `json:"job_id"`
		Files    []UploadFileResult `json:"files"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(body.Uploaded) != 1 || body.JobID != 5 || len(body.Files) != 1 || body.Files[0].Status != UploadStatusUploaded {
		t.Fatalf("unexpected body %s", recorder.Body.String())
	}
}
