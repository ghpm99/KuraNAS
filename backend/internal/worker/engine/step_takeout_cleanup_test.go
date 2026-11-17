package engine

import (
	"encoding/json"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/config"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type takeoutSessionFixture struct {
	entryPoint string
	sessionDir string
	zipPath    string
	imageLib   string
}

func newTakeoutSessionFixture(t *testing.T) takeoutSessionFixture {
	t.Helper()
	entryPoint := t.TempDir()
	previousEntryPoint := config.AppConfig.EntryPoint
	config.AppConfig.EntryPoint = entryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = previousEntryPoint })

	sessionDir := filepath.Join(entryPoint, ".takeout_uploads", "session-1")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("create session dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "session.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("write session meta: %v", err)
	}
	return takeoutSessionFixture{
		entryPoint: entryPoint,
		sessionDir: sessionDir,
		zipPath:    filepath.Join(sessionDir, "takeout.zip"),
		imageLib:   filepath.Join(entryPoint, "Imagens"),
	}
}

func runStepForZip(t *testing.T, zipPath string, librariesService libraries.ServiceInterface, recorder *takeoutImportRecorderMock) error {
	t.Helper()
	payload, err := json.Marshal(TakeoutStepPayload{ZipPath: zipPath, UploadID: "session-1"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	workerContext := &WorkerContext{LibrariesService: librariesService}
	if recorder != nil {
		workerContext.TakeoutImports = recorder
	}
	return executeTakeoutExtractStep(workerContext, jobs.StepModel{JobID: 4, Payload: payload})
}

func assertPathMissing(t *testing.T, targetPath string) {
	t.Helper()
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, stat error: %v", targetPath, err)
	}
}

func TestExecuteTakeoutExtractStepRemovesZipAndSessionDirOnSuccess(t *testing.T) {
	fixture := newTakeoutSessionFixture(t)
	createWorkerZip(t, fixture.zipPath)

	err := runStepForZip(t, fixture.zipPath, &librariesServiceMock{
		paths: map[libraries.LibraryCategory]string{libraries.LibraryCategoryImages: fixture.imageLib},
	}, nil)
	if err != nil {
		t.Fatalf("step returned error: %v", err)
	}

	assertPathMissing(t, fixture.zipPath)
	assertPathMissing(t, fixture.sessionDir)
	if _, statErr := os.Stat(filepath.Join(fixture.entryPoint, ".takeout_uploads")); statErr != nil {
		t.Fatalf("sessions root must be kept: %v", statErr)
	}
}

func TestExecuteTakeoutExtractStepRemovesZipAndSessionDirWhenExtractionFails(t *testing.T) {
	loadTranslationsForTest(t)
	fixture := newTakeoutSessionFixture(t)
	createWorkerZip(t, fixture.zipPath)

	if err := runStepForZip(t, fixture.zipPath, &failingLibrariesServiceMock{}, nil); err == nil {
		t.Fatalf("expected extraction error")
	}

	assertPathMissing(t, fixture.zipPath)
	assertPathMissing(t, fixture.sessionDir)
}

func TestExecuteTakeoutExtractStepRemovesSessionDirWhenZipIsInvalid(t *testing.T) {
	loadTranslationsForTest(t)
	fixture := newTakeoutSessionFixture(t)
	if err := os.WriteFile(fixture.zipPath, []byte("not a zip"), 0644); err != nil {
		t.Fatalf("write fake zip: %v", err)
	}

	if err := runStepForZip(t, fixture.zipPath, &librariesServiceMock{}, nil); err == nil {
		t.Fatalf("expected invalid zip error")
	}

	assertPathMissing(t, fixture.sessionDir)
}

func TestExecuteTakeoutExtractStepKeepsZipOutsideUploadSessionsArea(t *testing.T) {
	fixture := newTakeoutSessionFixture(t)
	outsideDir := filepath.Join(fixture.entryPoint, "Documentos")
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatalf("create outside dir: %v", err)
	}
	outsideZip := filepath.Join(outsideDir, "takeout.zip")
	createWorkerZip(t, outsideZip)
	bystander := filepath.Join(outsideDir, "keep.txt")
	if err := os.WriteFile(bystander, []byte("keep"), 0644); err != nil {
		t.Fatalf("write bystander: %v", err)
	}

	err := runStepForZip(t, outsideZip, &librariesServiceMock{
		paths: map[libraries.LibraryCategory]string{libraries.LibraryCategoryImages: fixture.imageLib},
	}, nil)
	if err != nil {
		t.Fatalf("step returned error: %v", err)
	}

	for _, preserved := range []string{outsideZip, bystander, fixture.sessionDir} {
		if _, statErr := os.Stat(preserved); statErr != nil {
			t.Fatalf("%s must be preserved: %v", preserved, statErr)
		}
	}
}

func TestExecuteTakeoutExtractStepKeepsNestedZipInsideSessionsArea(t *testing.T) {
	fixture := newTakeoutSessionFixture(t)
	nestedDir := filepath.Join(fixture.sessionDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}
	nestedZip := filepath.Join(nestedDir, "takeout.zip")
	createWorkerZip(t, nestedZip)

	err := runStepForZip(t, nestedZip, &librariesServiceMock{
		paths: map[libraries.LibraryCategory]string{libraries.LibraryCategoryImages: fixture.imageLib},
	}, nil)
	if err != nil {
		t.Fatalf("step returned error: %v", err)
	}

	if _, statErr := os.Stat(fixture.sessionDir); statErr != nil {
		t.Fatalf("session dir must be preserved: %v", statErr)
	}
}

func TestExecuteTakeoutExtractStepRecordsProgressEndingAtTotal(t *testing.T) {
	fixture := newTakeoutSessionFixture(t)
	createWorkerZip(t, fixture.zipPath)
	recorder := &takeoutImportRecorderMock{}

	err := runStepForZip(t, fixture.zipPath, &librariesServiceMock{
		paths: map[libraries.LibraryCategory]string{libraries.LibraryCategoryImages: fixture.imageLib},
	}, recorder)
	if err != nil {
		t.Fatalf("step returned error: %v", err)
	}

	if len(recorder.progressUpdates) != 1 || recorder.progressUpdates[0] != [2]int{1, 1} {
		t.Fatalf("expected a single progress update [1 1], got %v", recorder.progressUpdates)
	}
}

func TestTakeoutProgressReporterThrottlesButAlwaysReportsFirstAndLast(t *testing.T) {
	recorder := &takeoutImportRecorderMock{}
	reporter := newTakeoutProgressReporter(&WorkerContext{TakeoutImports: recorder}, jobs.StepModel{JobID: 4})
	clock := time.Unix(1000, 0)
	reporter.now = func() time.Time { return clock }

	reporter.report(1, 200)
	reporter.report(2, 200)
	reporter.report(30, 200)
	reporter.report(51, 200)
	clock = clock.Add(2 * time.Second)
	reporter.report(60, 200)
	reporter.report(61, 200)
	reporter.report(200, 200)

	expected := [][2]int{{1, 200}, {51, 200}, {60, 200}, {200, 200}}
	if len(recorder.progressUpdates) != len(expected) {
		t.Fatalf("expected updates %v, got %v", expected, recorder.progressUpdates)
	}
	for index, update := range expected {
		if recorder.progressUpdates[index] != update {
			t.Fatalf("expected updates %v, got %v", expected, recorder.progressUpdates)
		}
	}
}

func TestTakeoutProgressPercent(t *testing.T) {
	cases := []struct{ processed, total, expected int }{{0, 0, 100}, {1, 4, 25}, {4, 4, 100}}
	for _, testCase := range cases {
		if got := takeoutProgressPercent(testCase.processed, testCase.total); got != testCase.expected {
			t.Fatalf("percent(%d,%d) = %d, want %d", testCase.processed, testCase.total, got, testCase.expected)
		}
	}
}
