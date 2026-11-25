package engine

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/takeout"
	"nas-go/api/pkg/i18n"
	"os"
	"path/filepath"
	"reflect"
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
	payload, err := json.Marshal(TakeoutStepPayload{ZipPath: zipPath, UploadID: "u1", ImportID: 77})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	notifier := &fakeWorkerNotifSvc{}
	stepErr := executeTakeoutExtractStep(&WorkerContext{
		LibrariesService:    librariesService,
		NotificationService: notifier,
	}, jobs.StepModel{JobID: 9, Payload: payload})
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
	failedSummary    takeout.ImportSummary
	recordedJobID    int
	closeErr         error
	progressUpdates  [][2]int
}

func (m *takeoutImportRecorderMock) MarkImportRunning(jobID int) error {
	m.events = append(m.events, "running")
	m.recordedJobID = jobID
	return nil
}

func (m *takeoutImportRecorderMock) UpdateImportProgress(jobID int, processedEntries int, totalEntries int) error {
	m.progressUpdates = append(m.progressUpdates, [2]int{processedEntries, totalEntries})
	return nil
}

func (m *takeoutImportRecorderMock) CloseImportCompleted(jobID int, summary takeout.ImportSummary) error {
	m.events = append(m.events, "completed")
	m.completedSummary = summary
	return m.closeErr
}

func (m *takeoutImportRecorderMock) CloseImportFailed(jobID int, failureMessage string, summary takeout.ImportSummary) error {
	m.events = append(m.events, "failed")
	m.failedSummary = summary
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

func assertTakeoutEventMetadata(t *testing.T, metadata any, expectedEvent string) {
	t.Helper()
	expected := map[string]any{"event": expectedEvent, "import_id": 77, "job_id": 9}
	if !reflect.DeepEqual(metadata, expected) {
		t.Fatalf("expected metadata %v, got %v", expected, metadata)
	}
}

func TestExecuteTakeoutExtractStepCompletionNotificationCarriesEventMetadata(t *testing.T) {
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
	assertTakeoutEventMetadata(t, notifier.dtos[0].Metadata, "takeout_import_completed")
}

func TestExecuteTakeoutExtractStepFailureNotificationCarriesEventMetadata(t *testing.T) {
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
	assertTakeoutEventMetadata(t, notifier.dtos[0].Metadata, "takeout_import_failed")
}

type resolverFailingOnVideosMock struct {
	librariesServiceMock
}

func (m *resolverFailingOnVideosMock) GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error) {
	if category == libraries.LibraryCategoryVideos {
		return libraries.LibraryDto{}, errors.New("videos library unavailable")
	}
	return m.librariesServiceMock.GetLibraryByCategory(category)
}

type uploadProcessJobSpy struct {
	files.ServiceInterface
	enqueuedBatches [][]string
}

func (spy *uploadProcessJobSpy) CreateUploadProcessJob(paths []string) (int, error) {
	spy.enqueuedBatches = append(spy.enqueuedBatches, paths)
	return 1, nil
}

func writeTakeoutZipWithEntries(t *testing.T, zipPath string, entryNames []string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	for _, entryName := range entryNames {
		entry, createErr := writer.Create(entryName)
		if createErr != nil {
			t.Fatalf("failed to create zip entry: %v", createErr)
		}
		_, _ = entry.Write([]byte(takeoutZipEntryContent(entryName)))
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
}

func takeoutZipEntryContent(entryName string) string {
	if !strings.HasSuffix(entryName, ".json") {
		return "content"
	}
	photoTitle := strings.TrimSuffix(filepath.Base(entryName), ".json")
	return `{"title":"` + photoTitle + `","description":"Praia","geoData":{"latitude":-23.5,"longitude":-46.6}}`
}

type abortedTakeoutRun struct {
	stepErr        error
	sidecars       *sidecarRecorderSpy
	indexing       *uploadProcessJobSpy
	imports        *takeoutImportRecorderMock
	notifier       *fakeWorkerNotifSvc
	imageLibrary   string
	expectedFailed string
}

func runTakeoutStepAbortedByVideoLibrary(t *testing.T, entryNames []string) abortedTakeoutRun {
	t.Helper()
	loadTranslationsForTest(t)
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "upload", "takeout.zip")
	writeTakeoutZipWithEntries(t, zipPath, entryNames)
	imageLibrary := filepath.Join(tempDir, "Imagens")
	payload, _ := json.Marshal(TakeoutStepPayload{ZipPath: zipPath, UploadID: "u1", ImportID: 5})

	run := abortedTakeoutRun{
		sidecars:       &sidecarRecorderSpy{},
		indexing:       &uploadProcessJobSpy{},
		imports:        &takeoutImportRecorderMock{},
		notifier:       &fakeWorkerNotifSvc{},
		imageLibrary:   imageLibrary,
		expectedFailed: i18n.GetMessage("NOTIFICATION_TAKEOUT_IMPORT_FAILED_MESSAGE"),
	}
	run.stepErr = executeTakeoutExtractStep(&WorkerContext{
		LibrariesService:       &resolverFailingOnVideosMock{librariesServiceMock{paths: map[libraries.LibraryCategory]string{libraries.LibraryCategoryImages: imageLibrary}}},
		TakeoutSidecarRecorder: run.sidecars,
		FilesService:           run.indexing,
		TakeoutImports:         run.imports,
		NotificationService:    run.notifier,
	}, jobs.StepModel{JobID: 9, Payload: payload})
	return run
}

func TestExecuteTakeoutExtractStepAbortedExtractionStillIndexesAlreadyWrittenFiles(t *testing.T) {
	run := runTakeoutStepAbortedByVideoLibrary(t, []string{
		"Takeout/Google Fotos/IMG_01.jpg",
		"Takeout/Google Fotos/IMG_01.jpg.json",
		"Takeout/Google Fotos/IMG_02.jpg",
		"Takeout/Google Fotos/IMG_02.jpg.json",
		"Takeout/Google Fotos/clip.mp4",
	})

	if run.stepErr == nil {
		t.Fatalf("expected extraction error")
	}
	expectedPaths := []string{
		filepath.Join(run.imageLibrary, "takeout", "IMG_01.jpg"),
		filepath.Join(run.imageLibrary, "takeout", "IMG_02.jpg"),
	}
	if len(run.indexing.enqueuedBatches) != 1 || !reflect.DeepEqual(run.indexing.enqueuedBatches[0], expectedPaths) {
		t.Fatalf("expected one indexing job for %v, got %v", expectedPaths, run.indexing.enqueuedBatches)
	}
	if len(run.sidecars.saved) != 2 {
		t.Fatalf("expected sidecars for the 2 written files, got %+v", run.sidecars.saved)
	}
	if strings.Join(run.imports.events, ",") != "running,failed" || run.imports.failureMessage != run.expectedFailed {
		t.Fatalf("expected import closed as failed with translated message, got %v %q", run.imports.events, run.imports.failureMessage)
	}
	if len(run.notifier.dtos) != 1 || run.notifier.dtos[0].Message != run.expectedFailed {
		t.Fatalf("expected one failure notification, got %+v", run.notifier.dtos)
	}
}

func TestExecuteTakeoutExtractStepAbortedExtractionWithoutWrittenFilesEnqueuesNoIndexing(t *testing.T) {
	run := runTakeoutStepAbortedByVideoLibrary(t, []string{"Takeout/Google Fotos/clip.mp4"})

	if run.stepErr == nil {
		t.Fatalf("expected extraction error")
	}
	if len(run.indexing.enqueuedBatches) != 0 {
		t.Fatalf("expected no indexing job, got %v", run.indexing.enqueuedBatches)
	}
	if strings.Join(run.imports.events, ",") != "running,failed" {
		t.Fatalf("unexpected events %v", run.imports.events)
	}
}
