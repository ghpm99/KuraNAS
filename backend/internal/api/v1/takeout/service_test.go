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
	"strings"
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

type importCreatorMock struct {
	createdJobID    int
	createdFileName string
	createdSize     int64
	createdImportID int
	createErr       error
}

func (m *importCreatorMock) CreateQueuedImport(tx *sql.Tx, jobID int, fileName string, size int64) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	m.createdJobID = jobID
	m.createdFileName = fileName
	m.createdSize = size
	return m.createdImportID, nil
}

func newTakeoutServiceForTest(dispatcher *uploadDispatcherMock) *Service {
	dispatcher.db = database.NewDbContext(nil)
	return &Service{UploadJobDispatcher: dispatcher, ImportCreator: &importCreatorMock{createdImportID: 7}}
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
	result := NewService(&uploadDispatcherMock{db: database.NewDbContext(nil)}, nil, nil, &importCreatorMock{})
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

func TestCompleteUploadCreatesQueuedImportAndReturnsImportID(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	importCreator := &importCreatorMock{createdImportID: 42}
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	service.ImportCreator = importCreator

	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	if err := service.UploadChunk(chunk, UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}

	completeResult, err := service.CompleteUpload(CompleteTakeoutUploadDto{UploadID: initResult.UploadID})
	if err != nil {
		t.Fatalf("CompleteUpload returned error: %v", err)
	}
	if completeResult.ImportID != 42 {
		t.Fatalf("expected import id 42, got %d", completeResult.ImportID)
	}
	if importCreator.createdJobID != completeResult.JobID || importCreator.createdFileName != "takeout.zip" || importCreator.createdSize != 5 {
		t.Fatalf("unexpected queued import: %+v", importCreator)
	}
}

func TestCompleteUploadFailsWhenImportRecordCannotBeCreated(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	service.ImportCreator = &importCreatorMock{createErr: errors.New("insert failed")}

	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	if err := service.UploadChunk(chunk, UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}
	if _, err := service.CompleteUpload(CompleteTakeoutUploadDto{UploadID: initResult.UploadID}); err == nil {
		t.Fatalf("expected complete upload to fail")
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
	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	session, err := service.loadTakeoutUploadSession(initResult.UploadID)
	if err != nil {
		t.Fatalf("loadTakeoutUploadSession returned error: %v", err)
	}
	session.FileName = "takeout.txt"
	if err := service.saveTakeoutUploadSession(session); err != nil {
		t.Fatalf("saveTakeoutUploadSession returned error: %v", err)
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
	generatedUploadID := strings.Repeat("ab", takeoutUploadIDLength/2)
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	session := TakeoutUploadSession{
		UploadID:      generatedUploadID,
		FileName:      "takeout.zip",
		ExpectedSize:  1,
		ReceivedSize:  1,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
	}
	if err := os.MkdirAll(service.takeoutUploadSessionDir(generatedUploadID), 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	if err := service.saveTakeoutUploadSession(session); err != nil {
		t.Fatalf("saveTakeoutUploadSession returned error: %v", err)
	}
	loaded, err := service.loadTakeoutUploadSession(generatedUploadID)
	if err != nil {
		t.Fatalf("loadTakeoutUploadSession returned error: %v", err)
	}
	if loaded.UploadID != generatedUploadID {
		t.Fatalf("expected upload id generated")
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

func TestInitUploadRejectsInvalidSizeAndFileName(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	_, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 0})
	if !errors.Is(err, ErrInvalidUploadSize) {
		t.Fatalf("expected ErrInvalidUploadSize, got %v", err)
	}
	_, err = service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.rar", Size: 5})
	if !errors.Is(err, ErrInvalidUploadFileName) {
		t.Fatalf("expected ErrInvalidUploadFileName, got %v", err)
	}
}

func TestInitAndCompleteUploadAcceptUppercaseZipExtension(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "FOTOS.ZIP", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	chunk := buildChunkHeader(t, "chunk.bin", "hello")
	if err := service.UploadChunk(chunk, UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}
	_, err = service.CompleteUpload(CompleteTakeoutUploadDto{UploadID: initResult.UploadID})
	if errors.Is(err, ErrInvalidZipFile) {
		t.Fatalf("uppercase .ZIP must pass the extension check, got %v", err)
	}
}

func TestInitUploadResumesOpenSessionWithSameNameAndSize(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	firstInit, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	if firstInit.Resumed || firstInit.ReceivedSize != 0 {
		t.Fatalf("new session must not be resumed, got %+v", firstInit)
	}
	if err := service.UploadChunk(buildChunkHeader(t, "chunk.bin", "hello"), UploadTakeoutChunkDto{UploadID: firstInit.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}

	secondInit, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})
	if err != nil {
		t.Fatalf("second InitUpload returned error: %v", err)
	}
	if !secondInit.Resumed || secondInit.UploadID != firstInit.UploadID || secondInit.ReceivedSize != 5 {
		t.Fatalf("expected resumed session with 5 bytes, got %+v", secondInit)
	}
	if secondInit.ChunkSize != takeoutUploadChunkSize {
		t.Fatalf("unexpected chunk size %d", secondInit.ChunkSize)
	}
}

func TestInitUploadWithDifferentSizeCreatesNewSession(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	firstInit, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	secondInit, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 11})
	if err != nil {
		t.Fatalf("second InitUpload returned error: %v", err)
	}
	if secondInit.Resumed || secondInit.UploadID == firstInit.UploadID || secondInit.ReceivedSize != 0 {
		t.Fatalf("expected new session, got %+v", secondInit)
	}
}

func TestInitUploadDoesNotResumeFullyReceivedSession(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	firstInit, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 5})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	if err := service.UploadChunk(buildChunkHeader(t, "chunk.bin", "hello"), UploadTakeoutChunkDto{UploadID: firstInit.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}
	secondInit, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 5})
	if err != nil {
		t.Fatalf("second InitUpload returned error: %v", err)
	}
	if secondInit.Resumed || secondInit.UploadID == firstInit.UploadID {
		t.Fatalf("fully received session must not resume, got %+v", secondInit)
	}
}

func TestGetUploadStatusReturnsSizeAndReceivedSize(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	if err := service.UploadChunk(buildChunkHeader(t, "chunk.bin", "hello"), UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}

	status, err := service.GetUploadStatus(initResult.UploadID)
	if err != nil {
		t.Fatalf("GetUploadStatus returned error: %v", err)
	}
	if status.UploadID != initResult.UploadID || status.Size != 10 || status.ReceivedSize != 5 {
		t.Fatalf("unexpected status %+v", status)
	}
}

func TestGetUploadStatusRejectsUnknownAndMalformedIDs(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	unknownButWellFormedID := strings.Repeat("a", takeoutUploadIDLength)
	for _, uploadID := range []string{unknownButWellFormedID, "..", "../x", ""} {
		if _, err := service.GetUploadStatus(uploadID); !errors.Is(err, ErrUploadSessionNotFound) {
			t.Fatalf("id %q: expected ErrUploadSessionNotFound, got %v", uploadID, err)
		}
	}
}

func TestCancelUploadRemovesSessionFolder(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	if err := service.UploadChunk(buildChunkHeader(t, "chunk.bin", "hello"), UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}

	if err := service.CancelUpload(initResult.UploadID); err != nil {
		t.Fatalf("CancelUpload returned error: %v", err)
	}

	if _, err := os.Stat(service.takeoutUploadSessionDir(initResult.UploadID)); !os.IsNotExist(err) {
		t.Fatalf("expected session folder removed, stat err: %v", err)
	}
	if err := service.CancelUpload(initResult.UploadID); !errors.Is(err, ErrUploadSessionNotFound) {
		t.Fatalf("expected ErrUploadSessionNotFound on second cancel, got %v", err)
	}
}

func TestCancelUploadMalformedIDNeverDeletesOutsideSession(t *testing.T) {
	entryPoint := t.TempDir()
	setEntryPoint(t, entryPoint)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	outsideFile := filepath.Join(entryPoint, "keep.txt")
	if err := os.WriteFile(outsideFile, []byte("keep"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := os.MkdirAll(UploadSessionsRootDir(), 0755); err != nil {
		t.Fatalf("mkdir sessions root: %v", err)
	}

	for _, uploadID := range []string{"..", "../..", "../keep.txt", ".", "", "abc"} {
		if err := service.CancelUpload(uploadID); !errors.Is(err, ErrUploadSessionNotFound) {
			t.Fatalf("id %q: expected ErrUploadSessionNotFound, got %v", uploadID, err)
		}
	}

	if _, err := os.Stat(outsideFile); err != nil {
		t.Fatalf("outside file must survive: %v", err)
	}
	if _, err := os.Stat(UploadSessionsRootDir()); err != nil {
		t.Fatalf("sessions root must survive: %v", err)
	}
}

func fixedFreeSpace(availableBytes int64) func(string) (int64, error) {
	return func(string) (int64, error) { return availableBytes, nil }
}

func TestInitUploadRejectsSizeLargerThanFreeSpace(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	service.AvailableBytesAt = fixedFreeSpace(9)

	_, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})

	if !errors.Is(err, ErrInsufficientStorage) {
		t.Fatalf("expected ErrInsufficientStorage, got %v", err)
	}
	if entries, _ := os.ReadDir(service.takeoutUploadRootDir()); len(entries) != 0 {
		t.Fatalf("no session must be created, found %d entries", len(entries))
	}
}

func TestInitUploadAcceptsSizeEqualToFreeSpace(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	service.AvailableBytesAt = fixedFreeSpace(10)

	if _, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10}); err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
}

func TestInitUploadResumedSessionNeedsOnlyRemainingBytes(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	firstInit, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	if err := service.UploadChunk(buildChunkHeader(t, "chunk.bin", "hello"), UploadTakeoutChunkDto{UploadID: firstInit.UploadID, Offset: 0}); err != nil {
		t.Fatalf("UploadChunk returned error: %v", err)
	}

	service.AvailableBytesAt = fixedFreeSpace(5)
	resumed, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})
	if err != nil || !resumed.Resumed {
		t.Fatalf("expected resume with 5 bytes free, got %+v err=%v", resumed, err)
	}

	service.AvailableBytesAt = fixedFreeSpace(4)
	if _, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10}); !errors.Is(err, ErrInsufficientStorage) {
		t.Fatalf("expected ErrInsufficientStorage with 4 bytes free, got %v", err)
	}
}

func TestInitUploadReturnsErrorWhenFreeSpaceCannotBeRead(t *testing.T) {
	setEntryPoint(t, t.TempDir())
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})
	service.AvailableBytesAt = func(string) (int64, error) { return 0, errors.New("statfs failed") }

	_, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 10})

	if err == nil || errors.Is(err, ErrInsufficientStorage) {
		t.Fatalf("expected generic error, got %v", err)
	}
}

func TestUploadChunkRejectsChunkBeyondDeclaredSize(t *testing.T) {
	root := t.TempDir()
	setEntryPoint(t, root)
	service := newTakeoutServiceForTest(&uploadDispatcherMock{})

	initResult, err := service.InitUpload(InitTakeoutUploadDto{FileName: "takeout.zip", Size: 8})
	if err != nil {
		t.Fatalf("InitUpload returned error: %v", err)
	}
	firstChunk := buildChunkHeader(t, "chunk.bin", "hello")
	if err := service.UploadChunk(firstChunk, UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 0}); err != nil {
		t.Fatalf("first chunk returned error: %v", err)
	}

	oversizedChunk := buildChunkHeader(t, "chunk.bin", "world!")
	err = service.UploadChunk(oversizedChunk, UploadTakeoutChunkDto{UploadID: initResult.UploadID, Offset: 5})
	if !errors.Is(err, ErrUploadChunkExceedsSize) {
		t.Fatalf("expected ErrUploadChunkExceedsSize, got %v", err)
	}

	session, err := service.loadTakeoutUploadSession(initResult.UploadID)
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if session.ReceivedSize != 5 {
		t.Fatalf("expected received size 5, got %d", session.ReceivedSize)
	}
	info, err := os.Stat(service.takeoutUploadTempFilePath(initResult.UploadID))
	if err != nil {
		t.Fatalf("stat temp file: %v", err)
	}
	if info.Size() != 5 {
		t.Fatalf("expected temp file size 5, got %d", info.Size())
	}
}
