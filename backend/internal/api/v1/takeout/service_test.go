package takeout

import (
	"bytes"
	"database/sql"
	"errors"
	"mime/multipart"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type uploadDispatcherMock struct {
	db           *database.DbContext
	createJobFn  func(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error)
	createStepFn func(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error)
}

func (m *uploadDispatcherMock) GetDbContext() *database.DbContext { return m.db }
func (m *uploadDispatcherMock) CreateJob(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error) {
	if m.createJobFn != nil {
		return m.createJobFn(tx, job)
	}
	job.ID = 1
	return job, nil
}
func (m *uploadDispatcherMock) CreateStep(tx *sql.Tx, step jobs.StepModel) (jobs.StepModel, error) {
	if m.createStepFn != nil {
		return m.createStepFn(tx, step)
	}
	step.ID = 10
	return step, nil
}

type notificationServiceMock struct {
	called bool
}

func (m *notificationServiceMock) GroupOrCreate(dto notifications.CreateNotificationDto) (notifications.NotificationDto, error) {
	m.called = true
	return notifications.NotificationDto{ID: 1}, nil
}

func buildChunkHeader(t *testing.T, name string, content string) *multipart.FileHeader {
	t.Helper()
	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("chunk", name)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	_ = writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1024 * 1024)
	if err != nil {
		t.Fatalf("failed to read multipart form: %v", err)
	}
	return form.File["chunk"][0]
}

func setEntryPoint(t *testing.T, path string) {
	t.Helper()
	original := config.AppConfig.EntryPoint
	config.AppConfig.EntryPoint = path
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = original
	})
}

func newTakeoutServiceForTest(dispatcher *uploadDispatcherMock) *Service {
	dispatcher.db = database.NewDbContext(nil)
	return &Service{UploadJobDispatcher: dispatcher}
}

func TestInitUploadSuccess(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	result, err := service.InitUpload(InitTakeoutUploadDto{
		FileName: "takeout.zip",
		Size:     8,
	})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	if result.UploadID == "" {
		t.Fatalf("expected upload id")
	}
}

func TestNewService(t *testing.T) {
	result := NewService(&uploadDispatcherMock{db: database.NewDbContext(nil)}, nil, nil)
	if result == nil {
		t.Fatalf("expected non-nil service")
	}
}

func TestInitUploadEmptyFileName(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	_, err := service.InitUpload(InitTakeoutUploadDto{})
	if err == nil {
		t.Fatalf("expected error for empty file name")
	}
}

func TestUploadChunkSuccess(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}

	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	if err := service.UploadChunk(chunk, UploadTakeoutChunkDto{
		UploadID: initResult.UploadID,
		Offset:   0,
	}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}
}

func TestUploadChunkOffsetMismatch(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}

	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	err = service.UploadChunk(chunk, UploadTakeoutChunkDto{
		UploadID: initResult.UploadID,
		Offset:   1,
	})
	if !errors.Is(err, ErrUploadOffsetMismatch) {
		t.Fatalf("expected ErrUploadOffsetMismatch, got %v", err)
	}
}

func TestUploadChunkSessionNotFound(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	err := service.UploadChunk(chunk, UploadTakeoutChunkDto{UploadID: "missing"})
	if !errors.Is(err, ErrUploadSessionNotFound) {
		t.Fatalf("expected ErrUploadSessionNotFound, got %v", err)
	}
}

func TestCompleteUploadSuccess(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	if err := service.UploadChunk(chunk, UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}

	result, err := service.CompleteUpload(CompleteTakeoutUploadDto{UploadID: initResult.UploadID})
	if err != nil {
		t.Fatalf("CompleteUpload returned error: %v", err)
	}
	if result.JobID <= 0 {
		t.Fatalf("expected created job id")
	}
}

func TestCompleteUploadSizeMismatch(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	if err := service.UploadChunk(chunk, UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}

	_, err = service.CompleteUpload(CompleteTakeoutUploadDto{UploadID: initResult.UploadID})
	if !errors.Is(err, ErrUploadIncomplete) {
		t.Fatalf("expected ErrUploadIncomplete, got %v", err)
	}
}

func TestTakeoutUploadPaths(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	if service.takeoutUploadRootDir() != filepath.Join(root, ".takeout_uploads") {
		t.Fatalf("unexpected takeout root path")
	}
}

func TestSanitizeTakeoutFileName(t *testing.T) {
	if got := sanitizeTakeoutFileName("a:b.zip"); got != "a_b.zip" {
		t.Fatalf("expected sanitized file name, got %s", got)
	}
	if got := sanitizeTakeoutFileName(" "); got != "" {
		t.Fatalf("expected empty name for whitespace, got %s", got)
	}
}

func TestLoadTakeoutUploadSessionNotFound(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	_, err := service.loadTakeoutUploadSession("missing")
	if !errors.Is(err, ErrUploadSessionNotFound) {
		t.Fatalf("expected ErrUploadSessionNotFound, got %v", err)
	}
}

func TestCompleteUploadInvalidZipExtension(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.txt", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	if err := service.UploadChunk(chunk, UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}
	_, err = service.CompleteUpload(CompleteTakeoutUploadDto{UploadID: initResult.UploadID})
	if !errors.Is(err, ErrInvalidZipFile) {
		t.Fatalf("expected ErrInvalidZipFile, got %v", err)
	}
}

func TestCompleteUploadCreateJobFailure(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{
		createJobFn: func(tx *sql.Tx, job jobs.JobModel) (jobs.JobModel, error) {
			return jobs.JobModel{}, errors.New("create job failed")
		},
	})

	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	if err := service.UploadChunk(chunk, UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}
	_, err = service.CompleteUpload(CompleteTakeoutUploadDto{UploadID: initResult.UploadID})
	if err == nil {
		t.Fatalf("expected complete upload to fail")
	}
}

func TestGenerateTakeoutUploadID(t *testing.T) {
	uploadID, err := generateTakeoutUploadID()
	if err != nil {
		t.Fatalf("generateTakeoutUploadID returned error: %v", err)
	}
	if len(uploadID) == 0 {
		t.Fatalf("expected non-empty upload id")
	}
}

func TestSaveAndLoadTakeoutSessionRoundTrip(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	session := TakeoutUploadSession{
		UploadID:      "abc",
		FileName:      "takeout.zip",
		ExpectedSize:  1,
		ReceivedSize:  1,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
	}
	if err := os.MkdirAll(service.takeoutUploadSessionDir("abc"), 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	if err := service.saveTakeoutUploadSession(session); err != nil {
		t.Fatalf("saveTakeoutUploadSession returned error: %v", err)
	}
	loaded, err := service.loadTakeoutUploadSession("abc")
	if err != nil {
		t.Fatalf("loadTakeoutUploadSession returned error: %v", err)
	}
	if loaded.UploadID != "abc" {
		t.Fatalf("expected upload id abc")
	}
}

func TestEmitImportStartedNotification(t *testing.T) {
	mock := &notificationServiceMock{}
	service := &Service{NotificationService: mock}
	service.emitImportStartedNotification("takeout.zip", "u1")
	if !mock.called {
		t.Fatalf("expected notification call")
	}
}
