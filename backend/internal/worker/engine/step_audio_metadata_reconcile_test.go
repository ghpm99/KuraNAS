package engine

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
	musicdom "nas-go/api/internal/api/v1/music"
	"nas-go/api/internal/worker/job"
	"nas-go/api/internal/worker/scan"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type fakeAudioMetadataRepository struct {
	dbContext     *database.DbContext
	mock          sqlmock.Sqlmock
	missingByPage [][]musicdom.AudioWithoutMetadata
	missingCalls  int
	staleByPage   [][]musicdom.AudioWithStaleTags
	staleCalls    int
	listErr       error
	upsertedPaths []string
}

func newFakeAudioMetadataRepository(t *testing.T, transactionCount int) *fakeAudioMetadataRepository {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	for range transactionCount {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}
	return &fakeAudioMetadataRepository{dbContext: database.NewDbContext(db), mock: mock}
}

func (f *fakeAudioMetadataRepository) GetDbContext() *database.DbContext { return f.dbContext }

func (f *fakeAudioMetadataRepository) GetAudioMetadataByID(id int) (musicdom.AudioMetadataModel, error) {
	return musicdom.AudioMetadataModel{}, nil
}

func (f *fakeAudioMetadataRepository) UpsertAudioMetadata(tx *sql.Tx, metadata musicdom.AudioMetadataModel) (musicdom.AudioMetadataModel, error) {
	f.upsertedPaths = append(f.upsertedPaths, metadata.Path)
	return metadata, nil
}

func (f *fakeAudioMetadataRepository) DeleteAudioMetadata(id int) error { return nil }

func (f *fakeAudioMetadataRepository) ListAudioWithoutMetadata(afterFileID int, limit int) ([]musicdom.AudioWithoutMetadata, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.missingCalls >= len(f.missingByPage) {
		return nil, nil
	}
	page := f.missingByPage[f.missingCalls]
	f.missingCalls++
	return page, nil
}

func (f *fakeAudioMetadataRepository) ListAudioWithStaleTags(afterFileID int, limit int) ([]musicdom.AudioWithStaleTags, error) {
	if f.staleCalls >= len(f.staleByPage) {
		return nil, nil
	}
	page := f.staleByPage[f.staleCalls]
	f.staleCalls++
	return page, nil
}

func newAudioReconcileContext(repository *fakeAudioMetadataRepository, filesService *workerFilesServiceMock) *WorkerContext {
	return &WorkerContext{
		AudioMetadataRepository: repository,
		FilesService:            filesService,
		JobOrchestrator:         NewJobOrchestrator(newFakeJobsRepository(), nil),
	}
}

func TestExecuteAudioMetadataReconcileStep_RequiresDependencies(t *testing.T) {
	if err := executeAudioMetadataReconcileStep(nil, jobs.StepModel{}); err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestExecuteAudioMetadataReconcileStep_NothingMissingIsSkipped(t *testing.T) {
	ctx := newAudioReconcileContext(newFakeAudioMetadataRepository(t, 0), &workerFilesServiceMock{})
	if err := executeAudioMetadataReconcileStep(ctx, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestExecuteAudioMetadataReconcileStep_ListErrorFails(t *testing.T) {
	repository := newFakeAudioMetadataRepository(t, 0)
	repository.listErr = errors.New("boom")
	ctx := newAudioReconcileContext(repository, &workerFilesServiceMock{})
	if err := executeAudioMetadataReconcileStep(ctx, jobs.StepModel{}); err == nil {
		t.Fatal("expected error when listing fails")
	}
}

func TestExecuteAudioMetadataReconcileStep_RunsMetadataForEachMissingAudioAcrossPages(t *testing.T) {
	scan.SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		payload, _ := json.Marshal(musicdom.AudioMetadataModel{Title: "T"})
		return string(payload), nil
	})
	t.Cleanup(func() { scan.SetPythonScriptRunnerForTesting(nil) })

	fullPage := make([]musicdom.AudioWithoutMetadata, 0, audioMetadataReconcilePageSize)
	for fileID := 1; fileID <= audioMetadataReconcilePageSize; fileID++ {
		fullPage = append(fullPage, musicdom.AudioWithoutMetadata{FileID: fileID, Path: "/a.m4a"})
	}
	repository := newFakeAudioMetadataRepository(t, audioMetadataReconcilePageSize+1)
	repository.missingByPage = [][]musicdom.AudioWithoutMetadata{fullPage, {{FileID: 9999, Path: "/last.opus"}}}
	filesService := &workerFilesServiceMock{getFileByIDFn: func(id int) (files.FileDto, error) {
		return files.FileDto{ID: id, Name: "a.m4a", Path: "/a.m4a", Format: ".m4a"}, nil
	}}

	if err := executeAudioMetadataReconcileStep(newAudioReconcileContext(repository, filesService), jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.upsertedPaths) != audioMetadataReconcilePageSize+1 {
		t.Fatalf("expected %d audio metadata upserts, got %d", audioMetadataReconcilePageSize+1, len(repository.upsertedPaths))
	}
}

func TestExecuteAudioMetadataReconcileStep_ReprocessesAudioWithStaleTags(t *testing.T) {
	scan.SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		payload, _ := json.Marshal(musicdom.AudioMetadataModel{Title: "T", Artist: "A"})
		return string(payload), nil
	})
	t.Cleanup(func() { scan.SetPythonScriptRunnerForTesting(nil) })

	repository := newFakeAudioMetadataRepository(t, 2)
	repository.staleByPage = [][]musicdom.AudioWithStaleTags{{{FileID: 5, Path: "/stale.flac"}, {FileID: 6, Path: "/stale.ogg"}}}
	filesService := &workerFilesServiceMock{getFileByIDFn: func(id int) (files.FileDto, error) {
		return files.FileDto{ID: id, Name: "stale.flac", Path: "/stale.flac", Format: ".flac"}, nil
	}}

	if err := executeAudioMetadataReconcileStep(newAudioReconcileContext(repository, filesService), jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.upsertedPaths) != 2 {
		t.Fatalf("expected 2 re-extractions for stale tags, got %d", len(repository.upsertedPaths))
	}
}

func TestExecuteAudioMetadataReconcileStep_FailuresAreRecordedWithoutFailingTheStep(t *testing.T) {
	repository := newFakeAudioMetadataRepository(t, 0)
	repository.missingByPage = [][]musicdom.AudioWithoutMetadata{{{FileID: 1, Path: "/a.m4a"}, {FileID: 2, Path: "/b.ogg"}}}
	filesService := &workerFilesServiceMock{getFileByIDFn: func(id int) (files.FileDto, error) {
		return files.FileDto{}, errors.New("db down")
	}}

	if err := executeAudioMetadataReconcileStep(newAudioReconcileContext(repository, filesService), jobs.StepModel{}); err != nil {
		t.Fatalf("per-file failures must not fail the step, got %v", err)
	}
}

func TestBuildAudioMetadataReconcilePlan(t *testing.T) {
	plan := buildAudioMetadataReconcilePlan()
	if err := plan.Validate(); err != nil {
		t.Fatalf("plan must be valid: %v", err)
	}
	if plan.Type != job.JobTypeAudioMetadataReconcile || plan.Priority != job.JobPriorityLow {
		t.Fatalf("unexpected plan type/priority: %+v", plan)
	}
	if len(plan.Steps) != 1 || plan.Steps[0].Type != job.StepTypeAudioMetadataReconcile {
		t.Fatalf("expected single reconcile step, got %+v", plan.Steps)
	}
}

func TestEnqueueAudioMetadataReconcileJob_SkipsWithoutDependenciesAndEnqueuesWithThem(t *testing.T) {
	if err := enqueueAudioMetadataReconcileJob(nil); err != nil {
		t.Fatalf("nil context must be a no-op, got %v", err)
	}
	ctx := newAudioReconcileContext(newFakeAudioMetadataRepository(t, 0), &workerFilesServiceMock{})
	if err := enqueueAudioMetadataReconcileJob(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
