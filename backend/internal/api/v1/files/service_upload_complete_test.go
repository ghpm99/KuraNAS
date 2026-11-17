package files

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func (m *filesHandlerServiceMock) CompleteUpload(uploadID string) (UploadCompleteResult, error) {
	return UploadCompleteResult{}, nil
}

func newCompleteRouter(service *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.POST("/files/upload/init", handler.InitUploadHandler)
	router.POST("/files/upload/chunk", handler.UploadChunkHandler)
	router.POST("/files/upload/complete", handler.CompleteUploadHandler)
	return router
}

func postComplete(router *gin.Engine, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/files/upload/complete", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func completeBody(uploadID string) string {
	return `{"upload_id":"` + uploadID + `"}`
}

func initBody(fileName string, content string, extraFields string) string {
	return `{"file_name":"` + fileName + `","size":` + strconv.Itoa(len(content)) + extraFields + `}`
}

func uploadAllChunks(t *testing.T, router *gin.Engine, initRequestBody string, content string) string {
	t.Helper()
	uploadID := decodeInitResponse(t, postInitUpload(router, initRequestBody)).UploadID
	if recorder := postChunk(t, router, uploadID, "0", content); recorder.Code != http.StatusOK {
		t.Fatalf("chunk failed: %d %s", recorder.Code, recorder.Body.String())
	}
	return uploadID
}

type completeResponseBody struct {
	File  UploadFileResult `json:"file"`
	JobID int              `json:"job_id"`
}

func decodeCompleteResponse(t *testing.T, recorder *httptest.ResponseRecorder) completeResponseBody {
	t.Helper()
	var response completeResponseBody
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode complete response: %v (%s)", err, recorder.Body.String())
	}
	return response
}

func requireSessionDiscarded(t *testing.T, uploadID string) {
	t.Helper()
	if _, err := os.Stat(uploadSessionDir(uploadID)); !os.IsNotExist(err) {
		t.Fatalf("session %s must be discarded (stat err %v)", uploadID, err)
	}
}

func requireNothingPublished(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read %s: %v", directory, err)
	}
	if len(entries) != 0 {
		t.Fatalf("nothing may be published in %s, found %d entries", directory, len(entries))
	}
}

func TestCompleteUploadResumedAfterInterruptionPublishesIdenticalContent(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newCompleteRouter(environment.service)
	content := "first-half|second-half"
	body := initBody("movie.bin", content, `,"checksum":"`+sha256HexOf(content)+`"`)

	interruptedID := decodeInitResponse(t, postInitUpload(router, body)).UploadID
	postChunk(t, router, interruptedID, "0", "first-half|")
	resumed := decodeInitResponse(t, postInitUpload(router, body))
	if !resumed.Resumed || resumed.UploadID != interruptedID || resumed.ReceivedSize != 11 {
		t.Fatalf("unexpected resume: %+v", resumed)
	}
	postChunk(t, router, interruptedID, strconv.FormatInt(resumed.ReceivedSize, 10), "second-half")

	recorder := postComplete(router, completeBody(interruptedID))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	response := decodeCompleteResponse(t, recorder)
	publishedPath := filepath.Join(environment.entryPoint, "movie.bin")
	if response.File.Name != "movie.bin" || response.File.Path != publishedPath || response.File.Status != UploadStatusUploaded || response.JobID <= 0 {
		t.Fatalf("unexpected response: %+v", response)
	}
	publishedChecksum, err := sha256HexOfFile(publishedPath)
	if err != nil || publishedChecksum != sha256HexOf(content) {
		t.Fatalf("published checksum %q (err %v) differs from original", publishedChecksum, err)
	}
	requireSessionDiscarded(t, interruptedID)
}

func TestCompleteUploadRejectsDivergingSizeOrChecksum(t *testing.T) {
	cases := map[string]struct {
		initFields   string
		sentContent  string
		declaredName string
	}{
		"checksum differs": {initFields: `,"checksum":"` + sha256HexOf("abcdef") + `"`, sentContent: "abcdeX"},
		"size is short":    {sentContent: "abcde"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			environment := newUploadTestEnvironment(t, nil)
			router := newCompleteRouter(environment.service)
			declaredBody := `{"file_name":"a.bin","size":6` + testCase.initFields + `}`
			uploadID := uploadAllChunks(t, router, declaredBody, testCase.sentContent)

			recorder := postComplete(router, completeBody(uploadID))

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
			}
			requireNothingPublished(t, environment.entryPoint)
			requireSessionDiscarded(t, uploadID)
		})
	}
}

func TestCompleteUploadUnknownAndMalformedSession(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newCompleteRouter(environment.service)

	if recorder := postComplete(router, completeBody("0123456789abcdef0123456789abcdef")); recorder.Code != http.StatusNotFound {
		t.Fatalf("unknown session: expected 404, got %d", recorder.Code)
	}
	if recorder := postComplete(router, completeBody("../x")); recorder.Code != http.StatusBadRequest {
		t.Fatalf("malformed id: expected 400, got %d", recorder.Code)
	}
	if recorder := postComplete(router, `{`); recorder.Code != http.StatusBadRequest {
		t.Fatalf("malformed body: expected 400, got %d", recorder.Code)
	}
}

func TestCompleteUploadAppliesModifiedAt(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newCompleteRouter(environment.service)
	uploadID := uploadAllChunks(t, router, initBody("photo.jpg", "pixels", `,"modified_at":"2020-03-04T05:06:07Z"`), "pixels")

	recorder := postComplete(router, completeBody(uploadID))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	info, err := os.Stat(filepath.Join(environment.entryPoint, "photo.jpg"))
	if err != nil {
		t.Fatalf("stat published: %v", err)
	}
	if !info.ModTime().Equal(time.Date(2020, 3, 4, 5, 6, 7, 0, time.UTC)) {
		t.Fatalf("mtime = %v", info.ModTime())
	}
}

func TestCompleteUploadCreatesRelativeDirectories(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newCompleteRouter(environment.service)
	uploadID := uploadAllChunks(t, router, initBody("a.txt", "hello", `,"relative_path":"trip/day1"`), "hello")

	recorder := postComplete(router, completeBody(uploadID))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	requireFileContent(t, filepath.Join(environment.entryPoint, "trip", "day1", "a.txt"), "hello")
}

func TestCompleteUploadConflictPolicies(t *testing.T) {
	cases := map[string]struct {
		policy         string
		expectedCode   int
		expectedStatus string
		expectedName   string
		expectedOldTxt string
	}{
		"replace": {policy: "replace", expectedCode: http.StatusCreated, expectedStatus: UploadStatusReplaced, expectedName: "a.txt", expectedOldTxt: "new"},
		"rename":  {policy: "rename", expectedCode: http.StatusCreated, expectedStatus: UploadStatusRenamed, expectedName: "a (2).txt", expectedOldTxt: "old"},
		"fail":    {policy: "fail", expectedCode: http.StatusConflict, expectedOldTxt: "old"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			environment := newUploadTestEnvironment(t, nil)
			router := newCompleteRouter(environment.service)
			existingPath := filepath.Join(environment.entryPoint, "a.txt")
			uploadID := uploadAllChunks(t, router, initBody("a.txt", "new", `,"on_conflict":"`+testCase.policy+`"`), "new")
			writeExistingFile(t, existingPath, "old")

			recorder := postComplete(router, completeBody(uploadID))

			if recorder.Code != testCase.expectedCode {
				t.Fatalf("expected %d, got %d: %s", testCase.expectedCode, recorder.Code, recorder.Body.String())
			}
			requireFileContent(t, existingPath, testCase.expectedOldTxt)
			if testCase.expectedStatus == "" {
				return
			}
			response := decodeCompleteResponse(t, recorder)
			if response.File.Status != testCase.expectedStatus || response.File.Name != testCase.expectedName {
				t.Fatalf("unexpected response: %+v", response)
			}
			requireFileContent(t, filepath.Join(environment.entryPoint, testCase.expectedName), "new")
		})
	}
}

func TestCompleteUploadSkipDiscardsSessionWithoutJob(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newCompleteRouter(environment.service)
	existingPath := filepath.Join(environment.entryPoint, "a.txt")
	uploadID := uploadAllChunks(t, router, initBody("a.txt", "new", `,"on_conflict":"skip"`), "new")
	writeExistingFile(t, existingPath, "old")

	recorder := postComplete(router, completeBody(uploadID))

	response := decodeCompleteResponse(t, recorder)
	if recorder.Code != http.StatusCreated || response.File.Status != UploadStatusSkipped || response.JobID != 0 {
		t.Fatalf("unexpected response %d: %s", recorder.Code, recorder.Body.String())
	}
	requireFileContent(t, existingPath, "old")
	requireSessionDiscarded(t, uploadID)
}

func TestCompleteUploadAcrossVolumesPublishesThroughHiddenTemporaryFile(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	var remainingNames []string
	environment.service.RenamePartialFile = func(partialPath string, destinationPath string) error {
		return &os.LinkError{Op: "rename", Old: partialPath, New: destinationPath, Err: syscall.EXDEV}
	}
	router := newCompleteRouter(environment.service)
	uploadID := uploadAllChunks(t, router, initBody("big.bin", "payload", ""), "payload")

	recorder := postComplete(router, completeBody(uploadID))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	requireFileContent(t, filepath.Join(environment.entryPoint, "big.bin"), "payload")
	entries, _ := os.ReadDir(environment.entryPoint)
	for _, entry := range entries {
		remainingNames = append(remainingNames, entry.Name())
	}
	if len(remainingNames) != 1 {
		t.Fatalf("only the published file may remain, found %v", remainingNames)
	}
	requireSessionDiscarded(t, uploadID)
}

func TestUploadTemporaryFileNaming(t *testing.T) {
	temporaryPath := uploadTemporaryPathFor(filepath.Join("root", "dir", "movie.mp4"))

	if temporaryPath != filepath.Join("root", "dir", ".movie.mp4.kuranas-part") {
		t.Fatalf("unexpected temporary path %q", temporaryPath)
	}
	if !IsUploadTemporaryFile(temporaryPath) || IsUploadTemporaryFile(filepath.Join("root", "movie.mp4")) {
		t.Fatalf("temporary file detection is wrong")
	}
}

func TestCompleteUploadPropagatesPublishFailure(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	environment.service.RenamePartialFile = func(string, string) error { return os.ErrPermission }
	router := newCompleteRouter(environment.service)
	uploadID := uploadAllChunks(t, router, initBody("a.bin", "abc", ""), "abc")

	recorder := postComplete(router, completeBody(uploadID))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
