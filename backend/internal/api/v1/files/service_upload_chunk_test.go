package files

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func (m *filesHandlerServiceMock) UploadChunk(uploadID string, offset int64, chunk *multipart.FileHeader) (int64, error) {
	return 0, nil
}

func newChunkRouter(service *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{})
	router := gin.New()
	router.POST("/files/upload/init", handler.InitUploadHandler)
	router.POST("/files/upload/chunk", handler.UploadChunkHandler)
	return router
}

func openUploadSessionFor(t *testing.T, router *gin.Engine, fileName string, size int) string {
	t.Helper()
	body := `{"file_name":"` + fileName + `","size":` + strconv.Itoa(size) + `}`
	return decodeInitResponse(t, postInitUpload(router, body)).UploadID
}

func postChunk(t *testing.T, router *gin.Engine, uploadID string, offset string, content string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if content != "" {
		part, err := writer.CreateFormFile("chunk", "chunk.bin")
		if err != nil {
			t.Fatalf("CreateFormFile failed: %v", err)
		}
		part.Write([]byte(content))
	}
	writer.WriteField("upload_id", uploadID)
	writer.WriteField("offset", offset)
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/files/upload/chunk", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeReceivedSize(t *testing.T, recorder *httptest.ResponseRecorder) int64 {
	t.Helper()
	var response struct {
		ReceivedSize int64  `json:"received_size"`
		Error        string `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode chunk response: %v (%s)", err, recorder.Body.String())
	}
	return response.ReceivedSize
}

func readPartialFile(t *testing.T, uploadID string) string {
	t.Helper()
	content, err := os.ReadFile(uploadSessionPartialPath(uploadID))
	if err != nil {
		t.Fatalf("read partial file: %v", err)
	}
	return string(content)
}

func TestUploadChunkAccumulatesInOrder(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newChunkRouter(environment.service)
	uploadID := openUploadSessionFor(t, router, "a.bin", 10)

	first := postChunk(t, router, uploadID, "0", "hello")
	second := postChunk(t, router, uploadID, "5", "world")

	if first.Code != http.StatusOK || decodeReceivedSize(t, first) != 5 {
		t.Fatalf("first chunk: %d %s", first.Code, first.Body.String())
	}
	if second.Code != http.StatusOK || decodeReceivedSize(t, second) != 10 {
		t.Fatalf("second chunk: %d %s", second.Code, second.Body.String())
	}
	if readPartialFile(t, uploadID) != "helloworld" {
		t.Fatalf("partial file = %q", readPartialFile(t, uploadID))
	}
	session, _ := loadUploadSession(uploadID)
	if session.ReceivedSize != 10 || session.LastActivityAt.Before(session.CreatedAt) {
		t.Fatalf("session not updated: %+v", session)
	}
}

func TestUploadChunkWrongOffsetConflictsWithoutWriting(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newChunkRouter(environment.service)
	uploadID := openUploadSessionFor(t, router, "a.bin", 10)
	postChunk(t, router, uploadID, "0", "hello")

	for _, wrongOffset := range []string{"0", "3", "9"} {
		recorder := postChunk(t, router, uploadID, wrongOffset, "XXXXX")
		if recorder.Code != http.StatusConflict {
			t.Fatalf("offset %s: expected 409, got %d", wrongOffset, recorder.Code)
		}
		if decodeReceivedSize(t, recorder) != 5 {
			t.Fatalf("offset %s: received_size must be 5, body %s", wrongOffset, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), `"error"`) {
			t.Fatalf("409 body must carry error: %s", recorder.Body.String())
		}
	}
	if readPartialFile(t, uploadID) != "hello" {
		t.Fatalf("partial file changed: %q", readPartialFile(t, uploadID))
	}
}

func TestUploadChunkBeyondDeclaredSizeIsBadRequest(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newChunkRouter(environment.service)
	uploadID := openUploadSessionFor(t, router, "a.bin", 4)

	recorder := postChunk(t, router, uploadID, "0", "hello")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if readPartialFile(t, uploadID) != "" {
		t.Fatalf("nothing may be written, got %q", readPartialFile(t, uploadID))
	}
}

func TestUploadChunkUnknownSessionIsNotFound(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newChunkRouter(environment.service)

	recorder := postChunk(t, router, strings.Repeat("a", 32), "0", "x")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestUploadChunkRejectsMalformedRequests(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newChunkRouter(environment.service)
	uploadID := openUploadSessionFor(t, router, "a.bin", 4)
	cases := map[string]*httptest.ResponseRecorder{
		"malformed upload id": postChunk(t, router, "../x", "0", "x"),
		"empty upload id":     postChunk(t, router, "", "0", "x"),
		"missing chunk":       postChunk(t, router, uploadID, "0", ""),
		"non numeric offset":  postChunk(t, router, uploadID, "abc", "x"),
		"negative offset":     postChunk(t, router, uploadID, "-1", "x"),
	}
	for name, recorder := range cases {
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d: %s", name, recorder.Code, recorder.Body.String())
		}
	}
}

func TestUploadChunkRewritesLeftoverBytesBeyondReceivedSize(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newChunkRouter(environment.service)
	uploadID := openUploadSessionFor(t, router, "a.bin", 6)
	postChunk(t, router, uploadID, "0", "abc")
	if err := os.WriteFile(uploadSessionPartialPath(uploadID), []byte("abcZZZZZZ"), 0644); err != nil {
		t.Fatalf("simulate crash leftovers: %v", err)
	}

	postChunk(t, router, uploadID, "3", "def")

	if readPartialFile(t, uploadID) != "abcdef" {
		t.Fatalf("partial file = %q", readPartialFile(t, uploadID))
	}
}

func TestUploadNeverShowsPartialFileAtDestination(t *testing.T) {
	environment := newUploadTestEnvironment(t, nil)
	router := newChunkRouter(environment.service)
	uploadID := openUploadSessionFor(t, router, "movie.mp4", 10)
	postChunk(t, router, uploadID, "0", "hello")

	destinationEntries, err := os.ReadDir(environment.entryPoint)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}
	if len(destinationEntries) != 0 {
		t.Fatalf("destination must stay empty during upload, got %d entries", len(destinationEntries))
	}
	if strings.HasPrefix(uploadSessionsRootDir(), filepath.Clean(environment.entryPoint)) {
		t.Fatalf("sessions area %q must be outside the indexed root", uploadSessionsRootDir())
	}
}
