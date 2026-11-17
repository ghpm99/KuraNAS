package engine

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/takeout"
	"nas-go/api/pkg/i18n"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type librariesServiceMock struct {
	paths map[libraries.LibraryCategory]string
}

func (m *librariesServiceMock) GetLibraries() ([]libraries.LibraryDto, error) {
	return nil, nil
}

func (m *librariesServiceMock) GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error) {
	return libraries.LibraryDto{Category: string(category), Path: m.paths[category]}, nil
}

func (m *librariesServiceMock) UpdateLibrary(category libraries.LibraryCategory, dto libraries.UpdateLibraryDto) (libraries.LibraryDto, error) {
	return libraries.LibraryDto{}, nil
}

func (m *librariesServiceMock) ResolveLibraries() error {
	return nil
}

func createWorkerZip(t *testing.T, zipPath string) {
	t.Helper()
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	entry, err := writer.Create("Google Fotos/IMG_01.jpg")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	_, _ = entry.Write([]byte("img"))
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
}

func TestExecuteTakeoutExtractStepSuccess(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "upload", "takeout.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	createWorkerZip(t, zipPath)

	imageLib := filepath.Join(tempDir, "Imagens")
	if err := os.MkdirAll(imageLib, 0755); err != nil {
		t.Fatalf("failed to create image lib: %v", err)
	}

	payload, err := json.Marshal(TakeoutStepPayload{
		ZipPath:  zipPath,
		UploadID: "u1",
	})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	err = executeTakeoutExtractStep(&WorkerContext{
		LibrariesService: &librariesServiceMock{
			paths: map[libraries.LibraryCategory]string{
				libraries.LibraryCategoryImages: imageLib,
				libraries.LibraryCategoryVideos: filepath.Join(tempDir, "Videos"),
			},
		},
	}, jobs.StepModel{Payload: payload})
	if err != nil {
		t.Fatalf("executeTakeoutExtractStep returned error: %v", err)
	}
}

func TestExecuteTakeoutExtractStepInvalidPayload(t *testing.T) {
	err := executeTakeoutExtractStep(&WorkerContext{
		LibrariesService: &librariesServiceMock{
			paths: map[libraries.LibraryCategory]string{},
		},
	}, jobs.StepModel{Payload: []byte(`{`)})
	if err == nil {
		t.Fatalf("expected payload decode error")
	}
}

type failingLibrariesServiceMock struct {
	librariesServiceMock
}

func (m *failingLibrariesServiceMock) GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error) {
	return libraries.LibraryDto{}, errors.New("secret internal failure at /srv/internal/path")
}

func runTakeoutStepWithNotifier(t *testing.T, zipPath string, librariesService libraries.ServiceInterface) (*fakeWorkerNotifSvc, error) {
	t.Helper()
	payload, err := json.Marshal(TakeoutStepPayload{ZipPath: zipPath, UploadID: "u1"})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	notifier := &fakeWorkerNotifSvc{}
	stepErr := executeTakeoutExtractStep(&WorkerContext{
		LibrariesService:    librariesService,
		NotificationService: notifier,
	}, jobs.StepModel{Payload: payload})
	return notifier, stepErr
}

func loadTranslationsForTest(t *testing.T) {
	t.Helper()
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("failed to load translations: %v", err)
	}
}

func TestExecuteTakeoutExtractStepCompletionNotificationSummarizesCounts(t *testing.T) {
	loadTranslationsForTest(t)
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "upload", "takeout.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	createWorkerZip(t, zipPath)
	imageLib := filepath.Join(tempDir, "Imagens")
	if err := os.MkdirAll(imageLib, 0755); err != nil {
		t.Fatalf("failed to create image lib: %v", err)
	}

	notifier, err := runTakeoutStepWithNotifier(t, zipPath, &librariesServiceMock{
		paths: map[libraries.LibraryCategory]string{libraries.LibraryCategoryImages: imageLib},
	})
	if err != nil {
		t.Fatalf("executeTakeoutExtractStep returned error: %v", err)
	}
	if len(notifier.dtos) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.dtos))
	}
	expectedMessage := i18n.Translate("NOTIFICATION_TAKEOUT_IMPORT_COMPLETED_MESSAGE", 1, 0, 0, 0)
	if notifier.dtos[0].Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, notifier.dtos[0].Message)
	}
	for _, count := range []string{"1 ", "0 "} {
		if !strings.Contains(notifier.dtos[0].Message, count) {
			t.Fatalf("expected counts in message, got %q", notifier.dtos[0].Message)
		}
	}
}

func TestExecuteTakeoutExtractStepInvalidZipNotificationHidesErrorAndPath(t *testing.T) {
	loadTranslationsForTest(t)
	zipPath := filepath.Join(t.TempDir(), "upload", "takeout.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	if err := os.WriteFile(zipPath, []byte("not a zip"), 0644); err != nil {
		t.Fatalf("failed to write fake zip: %v", err)
	}

	notifier, err := runTakeoutStepWithNotifier(t, zipPath, &librariesServiceMock{})
	if !errors.Is(err, takeout.ErrInvalidZipFile) {
		t.Fatalf("expected ErrInvalidZipFile, got %v", err)
	}
	assertFailureNotificationIsSafe(t, notifier, "NOTIFICATION_TAKEOUT_IMPORT_FAILED_INVALID_ZIP_MESSAGE", zipPath)
}

func TestExecuteTakeoutExtractStepGenericFailureNotificationHidesErrorAndPath(t *testing.T) {
	loadTranslationsForTest(t)
	zipPath := filepath.Join(t.TempDir(), "upload", "takeout.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	createWorkerZip(t, zipPath)

	notifier, err := runTakeoutStepWithNotifier(t, zipPath, &failingLibrariesServiceMock{})
	if err == nil {
		t.Fatalf("expected extraction error")
	}
	assertFailureNotificationIsSafe(t, notifier, "NOTIFICATION_TAKEOUT_IMPORT_FAILED_MESSAGE", zipPath)
}

func assertFailureNotificationIsSafe(t *testing.T, notifier *fakeWorkerNotifSvc, expectedMessageKey string, zipPath string) {
	t.Helper()
	if len(notifier.dtos) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.dtos))
	}
	message := notifier.dtos[0].Message
	if message != i18n.GetMessage(expectedMessageKey) || message == expectedMessageKey {
		t.Fatalf("expected translated %s, got %q", expectedMessageKey, message)
	}
	for _, leaked := range []string{zipPath, filepath.Dir(zipPath), "secret internal failure", "/srv/internal/path", "invalid zip file"} {
		if strings.Contains(message, leaked) {
			t.Fatalf("notification leaked %q: %q", leaked, message)
		}
	}
}

type takeoutImportRecorderMock struct {
	events           []string
	completedSummary takeout.ImportSummary
	failureMessage   string
	recordedJobID    int
	closeErr         error
}

func (m *takeoutImportRecorderMock) MarkImportRunning(jobID int) error {
	m.events = append(m.events, "running")
	m.recordedJobID = jobID
	return nil
}

func (m *takeoutImportRecorderMock) CloseImportCompleted(jobID int, summary takeout.ImportSummary) error {
	m.events = append(m.events, "completed")
	m.completedSummary = summary
	return m.closeErr
}

func (m *takeoutImportRecorderMock) CloseImportFailed(jobID int, failureMessage string) error {
	m.events = append(m.events, "failed")
	m.failureMessage = failureMessage
	return m.closeErr
}

func runTakeoutStepWithRecorder(t *testing.T, zipPath string, librariesService libraries.ServiceInterface, recorder *takeoutImportRecorderMock) error {
	t.Helper()
	payload, err := json.Marshal(TakeoutStepPayload{ZipPath: zipPath, UploadID: "u1"})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	return executeTakeoutExtractStep(&WorkerContext{
		LibrariesService: librariesService,
		TakeoutImports:   recorder,
	}, jobs.StepModel{JobID: 9, Payload: payload})
}

func TestExecuteTakeoutExtractStepRecordsRunningThenCompletedWithCounts(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "upload", "takeout.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	createWorkerZip(t, zipPath)
	imageLib := filepath.Join(tempDir, "Imagens")
	if err := os.MkdirAll(imageLib, 0755); err != nil {
		t.Fatalf("failed to create image lib: %v", err)
	}

	recorder := &takeoutImportRecorderMock{}
	err := runTakeoutStepWithRecorder(t, zipPath, &librariesServiceMock{
		paths: map[libraries.LibraryCategory]string{libraries.LibraryCategoryImages: imageLib},
	}, recorder)
	if err != nil {
		t.Fatalf("executeTakeoutExtractStep returned error: %v", err)
	}
	if strings.Join(recorder.events, ",") != "running,completed" || recorder.recordedJobID != 9 {
		t.Fatalf("unexpected events %v for job %d", recorder.events, recorder.recordedJobID)
	}
	expectedSummary := takeout.ImportSummary{TotalEntries: 1, Imported: 1}
	if recorder.completedSummary != expectedSummary {
		t.Fatalf("expected summary %+v, got %+v", expectedSummary, recorder.completedSummary)
	}
}

func TestExecuteTakeoutExtractStepRecordsFailedWithTranslatedMessage(t *testing.T) {
	loadTranslationsForTest(t)
	zipPath := filepath.Join(t.TempDir(), "upload", "takeout.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	createWorkerZip(t, zipPath)

	recorder := &takeoutImportRecorderMock{}
	if err := runTakeoutStepWithRecorder(t, zipPath, &failingLibrariesServiceMock{}, recorder); err == nil {
		t.Fatalf("expected extraction error")
	}
	if strings.Join(recorder.events, ",") != "running,failed" {
		t.Fatalf("unexpected events %v", recorder.events)
	}
	expectedMessage := i18n.GetMessage("NOTIFICATION_TAKEOUT_IMPORT_FAILED_MESSAGE")
	if recorder.failureMessage != expectedMessage {
		t.Fatalf("expected %q, got %q", expectedMessage, recorder.failureMessage)
	}
	if strings.Contains(recorder.failureMessage, "secret") || strings.Contains(recorder.failureMessage, zipPath) {
		t.Fatalf("failure message leaks internals: %q", recorder.failureMessage)
	}
}

func TestExecuteTakeoutExtractStepIgnoresRecordFailures(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "upload", "takeout.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	createWorkerZip(t, zipPath)
	imageLib := filepath.Join(tempDir, "Imagens")
	if err := os.MkdirAll(imageLib, 0755); err != nil {
		t.Fatalf("failed to create image lib: %v", err)
	}

	recorder := &takeoutImportRecorderMock{closeErr: errors.New("db down")}
	err := runTakeoutStepWithRecorder(t, zipPath, &librariesServiceMock{
		paths: map[libraries.LibraryCategory]string{libraries.LibraryCategoryImages: imageLib},
	}, recorder)
	if err != nil {
		t.Fatalf("record failure must not fail the import, got %v", err)
	}
}

type sidecarRecorderSpy struct {
	saved []takeout.PhotoSidecar
}

func (spy *sidecarRecorderSpy) SaveSidecars(sidecars []takeout.PhotoSidecar) error {
	spy.saved = append(spy.saved, sidecars...)
	return nil
}

func TestExecuteTakeoutExtractStepRecordsSidecarOfExtractedPhoto(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "upload", "takeout.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	writer := zip.NewWriter(zipFile)
	for name, content := range map[string]string{
		"Takeout/Google Fotos/IMG_01.jpg":      "img",
		"Takeout/Google Fotos/IMG_01.jpg.json": `{"title":"IMG_01.jpg","description":"Praia","geoData":{"latitude":-23.5,"longitude":-46.6}}`,
		"Takeout/Google Fotos/IMG_02.jpg":      "img2",
	} {
		entry, createErr := writer.Create(name)
		if createErr != nil {
			t.Fatalf("failed to create zip entry: %v", createErr)
		}
		_, _ = entry.Write([]byte(content))
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	zipFile.Close()

	imageLib := filepath.Join(tempDir, "Imagens")
	payload, _ := json.Marshal(TakeoutStepPayload{ZipPath: zipPath, UploadID: "u1"})
	spy := &sidecarRecorderSpy{}

	stepErr := executeTakeoutExtractStep(&WorkerContext{
		LibrariesService:       &librariesServiceMock{paths: map[libraries.LibraryCategory]string{libraries.LibraryCategoryImages: imageLib}},
		TakeoutSidecarRecorder: spy,
	}, jobs.StepModel{Payload: payload})
	if stepErr != nil {
		t.Fatalf("executeTakeoutExtractStep returned error: %v", stepErr)
	}

	expectedPath := filepath.Join(imageLib, "takeout", "IMG_01.jpg")
	if len(spy.saved) != 1 || spy.saved[0].Path != expectedPath || spy.saved[0].Description != "Praia" || !spy.saved[0].HasLocation() {
		t.Fatalf("expected exactly the sidecar of IMG_01.jpg at %s, got %+v", expectedPath, spy.saved)
	}
}
