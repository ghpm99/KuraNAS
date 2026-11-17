package files

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/internal/config"

	"github.com/gin-gonic/gin"
)

func (m *filesHandlerServiceMock) InitUpload(request UploadInitRequest) (UploadInitResult, error) {
	return UploadInitResult{}, nil
}

type uploadTestEnvironment struct {
	service    *Service
	entryPoint string
	sessions   string
}

func newUploadTestEnvironment(t *testing.T, records []FileModel) uploadTestEnvironment {
	t.Helper()
	service, entryPoint := newUploadTestService(t, records)
	sessionsRoot := t.TempDir()
	setCapturesPathForTest(t, sessionsRoot)
	service.AvailableBytesAt = func(string) (int64, error) { return 1 << 40, nil }
	return uploadTestEnvironment{service: service, entryPoint: entryPoint, sessions: sessionsRoot}
}

func setCapturesPathForTest(t *testing.T, capturesPath string) {
	t.Helper()
	previousCapturesPath := config.AppConfig.CapturesPath
	config.AppConfig.CapturesPath = capturesPath
	t.Cleanup(func() { config.AppConfig.CapturesPath = previousCapturesPath })
}

func sha256HexOf(content string) string {
	digest := sha256.Sum256([]byte(content))
	return hex.EncodeToString(digest[:])
}

func newUploadRouter(service *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.POST("/files/upload/init", handler.InitUploadHandler)
	return router
}

func postInitUpload(router *gin.Engine, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/files/upload/init", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeInitResponse(t *testing.T, recorder *httptest.ResponseRecorder) initUploadResponse {
	t.Helper()
	var response initUploadResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode init response: %v (%s)", err, recorder.Body.String())
	}
	return response
}

func TestInitUploadHandlerOpensSession(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newUploadRouter(environment.service)

	recorder := postInitUpload(router, `{"file_name":"video.mp4","size":5000000,"modified_at":"2024-05-01T10:00:00Z"}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	response := decodeInitResponse(t, recorder)
	if !isValidUploadID(response.UploadID) || response.ChunkSize != 2*1024*1024 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.ReceivedSize != 0 || response.Resumed || response.AlreadyExists {
		t.Fatalf("unexpected flags: %+v", response)
	}
	session, err := loadUploadSession(response.UploadID)
	if err != nil {
		t.Fatalf("session not persisted: %v", err)
	}
	if session.FileName != "video.mp4" || session.Size != 5000000 || session.TargetFolderPath != filepath.Clean(environment.entryPoint) || session.ModifiedAt == nil {
		t.Fatalf("unexpected session: %+v", session)
	}
}

func TestInitUploadHandlerRejectsInvalidRequests(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newUploadRouter(environment.service)
	cases := map[string]string{
		"missing size":       `{"file_name":"a.txt"}`,
		"malformed json":     `{`,
		"empty name":         `{"file_name":"  ","size":1}`,
		"slash in name":      `{"file_name":"a/b.txt","size":1}`,
		"dotdot in name":     `{"file_name":"..","size":1}`,
		"name too long":      `{"file_name":"` + string(bytes.Repeat([]byte("a"), 256)) + `","size":1}`,
		"negative size":      `{"file_name":"a.txt","size":-1}`,
		"traversal":          `{"file_name":"a.txt","size":1,"relative_path":"../x"}`,
		"absolute":           `{"file_name":"a.txt","size":1,"relative_path":"/etc"}`,
		"bad checksum":       `{"file_name":"a.txt","size":1,"checksum":"zz"}`,
		"bad conflict":       `{"file_name":"a.txt","size":1,"on_conflict":"overwrite"}`,
		"bad modified_at":    `{"file_name":"a.txt","size":1,"modified_at":"ontem"}`,
		"drive letter":       `{"file_name":"a.txt","size":1,"relative_path":"C:/x"}`,
		"nul in name":        `{"file_name":"a\u0000b","size":1}`,
		"control char name":  `{"file_name":"a\nb","size":1}`,
		"backslash relative": `{"file_name":"a.txt","size":1,"relative_path":"..\\x"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			recorder := postInitUpload(router, body)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestInitUploadHandlerUnknownFolderIsNotFound(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newUploadRouter(environment.service)

	recorder := postInitUpload(router, `{"file_name":"a.txt","size":1,"target_folder_id":99}`)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestInitUploadTargetMustBeDirectory(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	filePath := filepath.Join(environment.entryPoint, "plain.txt")
	writeExistingFile(t, filePath, "x")
	environment.service = newUploadTestServiceWithRecords(t, environment, []FileModel{
		{ID: 5, Name: "plain.txt", Path: filePath, Type: File},
	})

	_, err := environment.service.InitUpload(UploadInitRequest{FileName: "a.txt", Size: 1, TargetFolderID: 5})

	requireOperationError(t, err, http.StatusBadRequest, "ERROR_TARGET_NOT_DIRECTORY")
}

func newUploadTestServiceWithRecords(t *testing.T, environment uploadTestEnvironment, records []FileModel) *Service {
	t.Helper()
	service := newTestServiceWithFileRecords(t, environment.entryPoint, records)
	service.AvailableBytesAt = environment.service.AvailableBytesAt
	return service
}

func TestInitUploadInsufficientSpaceBeforeAnyPart(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	environment.service.AvailableBytesAt = func(string) (int64, error) { return 999, nil }
	router := newUploadRouter(environment.service)

	recorder := postInitUpload(router, `{"file_name":"big.bin","size":1000}`)

	if recorder.Code != http.StatusInsufficientStorage {
		t.Fatalf("expected 507, got %d: %s", recorder.Code, recorder.Body.String())
	}
	sessionEntries, _ := os.ReadDir(uploadSessionsRootDir())
	if len(sessionEntries) != 0 {
		t.Fatalf("no session may be created, found %d", len(sessionEntries))
	}
}

func TestInitUploadFailsOnConflictByDefault(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	writeExistingFile(t, filepath.Join(environment.entryPoint, "a.txt"), "old")
	router := newUploadRouter(environment.service)

	recorder := postInitUpload(router, `{"file_name":"a.txt","size":3}`)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestInitUploadOtherConflictPoliciesOpenSession(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	writeExistingFile(t, filepath.Join(environment.entryPoint, "a.txt"), "old")
	router := newUploadRouter(environment.service)

	recorder := postInitUpload(router, `{"file_name":"a.txt","size":3,"on_conflict":"rename"}`)

	if recorder.Code != http.StatusCreated || decodeInitResponse(t, recorder).UploadID == "" {
		t.Fatalf("expected a session, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestInitUploadReportsAlreadyExistsByChecksum(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	writeExistingFile(t, filepath.Join(environment.entryPoint, "a.txt"), "same content")
	router := newUploadRouter(environment.service)

	body := `{"file_name":"a.txt","size":12,"checksum":"` + sha256HexOf("same content") + `"}`
	recorder := postInitUpload(router, body)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	response := decodeInitResponse(t, recorder)
	if !response.AlreadyExists || response.UploadID != "" {
		t.Fatalf("expected already_exists without upload_id, got %+v", response)
	}
}

func TestInitUploadDifferentContentWithChecksumStillConflicts(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	writeExistingFile(t, filepath.Join(environment.entryPoint, "a.txt"), "same content")
	router := newUploadRouter(environment.service)

	body := `{"file_name":"a.txt","size":12,"checksum":"` + sha256HexOf("other conten") + `"}`
	recorder := postInitUpload(router, body)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestInitUploadResumesEquivalentSession(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newUploadRouter(environment.service)
	first := decodeInitResponse(t, postInitUpload(router, `{"file_name":"a.bin","size":10,"relative_path":"photos/2024"}`))
	session, err := loadUploadSession(first.UploadID)
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	session.ReceivedSize = 4
	if err := saveUploadSession(session); err != nil {
		t.Fatalf("save session: %v", err)
	}

	second := decodeInitResponse(t, postInitUpload(router, `{"file_name":"a.bin","size":10,"relative_path":"./photos//2024/"}`))

	if !second.Resumed || second.UploadID != first.UploadID || second.ReceivedSize != 4 {
		t.Fatalf("expected resume of %s at 4, got %+v", first.UploadID, second)
	}
}

func TestInitUploadDifferentSizeOpensNewSession(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newUploadRouter(environment.service)
	first := decodeInitResponse(t, postInitUpload(router, `{"file_name":"a.bin","size":10}`))

	second := decodeInitResponse(t, postInitUpload(router, `{"file_name":"a.bin","size":11}`))

	if second.Resumed || second.UploadID == first.UploadID {
		t.Fatalf("expected a new session, got %+v", second)
	}
}

func TestInitUploadDoesNotTouchDestination(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newUploadRouter(environment.service)

	postInitUpload(router, `{"file_name":"a.bin","size":10,"relative_path":"new/sub"}`)

	destinationEntries, err := os.ReadDir(environment.entryPoint)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if len(destinationEntries) != 0 {
		t.Fatalf("destination must stay untouched during upload, found %d entries", len(destinationEntries))
	}
}

func TestInitUploadFreeSpaceMeasuredAtExistingAncestor(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	var measuredPath string
	environment.service.AvailableBytesAt = func(path string) (int64, error) {
		measuredPath = path
		return 1 << 40, nil
	}

	_, err := environment.service.InitUpload(UploadInitRequest{FileName: "a.bin", Size: 1, RelativePath: "missing/deeper"})

	if err != nil {
		t.Fatalf("InitUpload failed: %v", err)
	}
	if measuredPath != filepath.Clean(environment.entryPoint) {
		t.Fatalf("free space measured at %q, expected %q", measuredPath, environment.entryPoint)
	}
}

func TestValidateItemNameTrimsSpaces(t *testing.T) {
	name, err := validateItemName("  report.pdf ")
	if err != nil || name != "report.pdf" {
		t.Fatalf("expected trimmed name, got %q (%v)", name, err)
	}
}
