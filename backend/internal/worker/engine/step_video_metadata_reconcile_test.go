package engine

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
	videodom "nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/worker/job"
	"nas-go/api/internal/worker/scan"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type fakeVideoMetadataRepository struct {
	videodom.VideoMetadataRepositoryInterface
	dbContext        *database.DbContext
	missingByPage    [][]videodom.VideoWithoutMetadata
	missingCalls     int
	listErr          error
	upsertedPaths    []string
	upsertedMetadata []videodom.VideoMetadataModel

	pendingByPage       [][]videodom.VideoPendingClassification
	pendingCalls        int
	pendingErr          error
	classificationByID  map[int]string
	updateClassifyError error
}

func newFakeVideoMetadataRepository(t *testing.T, transactionCount int) *fakeVideoMetadataRepository {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	for range transactionCount {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}
	return &fakeVideoMetadataRepository{dbContext: database.NewDbContext(db)}
}

func (f *fakeVideoMetadataRepository) GetDbContext() *database.DbContext { return f.dbContext }

func (f *fakeVideoMetadataRepository) UpsertVideoMetadata(tx *sql.Tx, metadata videodom.VideoMetadataModel) (videodom.VideoMetadataModel, error) {
	f.upsertedPaths = append(f.upsertedPaths, metadata.Path)
	f.upsertedMetadata = append(f.upsertedMetadata, metadata)
	return metadata, nil
}

func (f *fakeVideoMetadataRepository) ListVideosWithoutMetadata(afterFileID int, limit int) ([]videodom.VideoWithoutMetadata, error) {
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

func (f *fakeVideoMetadataRepository) ListVideosPendingClassification(afterMetadataID int, limit int) ([]videodom.VideoPendingClassification, error) {
	if f.pendingErr != nil {
		return nil, f.pendingErr
	}
	if f.pendingCalls >= len(f.pendingByPage) {
		return nil, nil
	}
	page := f.pendingByPage[f.pendingCalls]
	f.pendingCalls++
	return page, nil
}

func (f *fakeVideoMetadataRepository) UpdateVideoClassification(metadataID int, classification string) error {
	if f.updateClassifyError != nil {
		return f.updateClassifyError
	}
	if f.classificationByID == nil {
		f.classificationByID = map[int]string{}
	}
	f.classificationByID[metadataID] = classification
	return nil
}

type videoReconcileObservations struct {
	thumbnailFileIDs []int
	playlistRebuilds int
}

func newVideoReconcileContext(repository *fakeVideoMetadataRepository, filesService *workerFilesServiceMock, observations *videoReconcileObservations) *WorkerContext {
	return &WorkerContext{
		VideoMetadataRepository: repository,
		FilesService:            filesService,
		VideoService: &workerVideoServiceMock{
			rebuildFn: func() error {
				observations.playlistRebuilds++
				return nil
			},
			getVideoThumbFn: func(fileDto files.FileDto, width, height int) ([]byte, error) {
				observations.thumbnailFileIDs = append(observations.thumbnailFileIDs, fileDto.ID)
				return []byte("thumb"), nil
			},
		},
		JobOrchestrator: NewJobOrchestrator(newFakeJobsRepository(), nil),
	}
}

func stubVideoMetadataScript(t *testing.T) {
	t.Helper()
	scan.SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		payload, _ := json.Marshal(videodom.VideoMetadataModel{Path: filePath, Width: 10, Height: 10})
		return string(payload), nil
	})
	t.Cleanup(func() { scan.SetPythonScriptRunnerForTesting(nil) })
}

func videoFilesService(failingFileID int) *workerFilesServiceMock {
	return &workerFilesServiceMock{getFileByIDFn: func(id int) (files.FileDto, error) {
		if id == failingFileID {
			return files.FileDto{}, errors.New("db down")
		}
		return files.FileDto{ID: id, Name: "v.mkv", Path: "/v.mkv", Format: ".mkv", Type: files.File}, nil
	}}
}

func TestExecuteVideoMetadataReconcileStep_RequiresDependencies(t *testing.T) {
	if err := executeVideoMetadataReconcileStep(nil, jobs.StepModel{}); err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestExecuteVideoMetadataReconcileStep_NothingMissingIsSkippedWithoutRebuild(t *testing.T) {
	observations := &videoReconcileObservations{}
	ctx := newVideoReconcileContext(newFakeVideoMetadataRepository(t, 0), &workerFilesServiceMock{}, observations)
	if err := executeVideoMetadataReconcileStep(ctx, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
	if observations.playlistRebuilds != 0 {
		t.Fatalf("expected no playlist rebuild, got %d", observations.playlistRebuilds)
	}
}

func TestExecuteVideoMetadataReconcileStep_ListErrorFails(t *testing.T) {
	repository := newFakeVideoMetadataRepository(t, 0)
	repository.listErr = errors.New("boom")
	ctx := newVideoReconcileContext(repository, &workerFilesServiceMock{}, &videoReconcileObservations{})
	if err := executeVideoMetadataReconcileStep(ctx, jobs.StepModel{}); err == nil {
		t.Fatal("expected error when listing fails")
	}
}

func TestExecuteVideoMetadataReconcileStep_ProcessesAcrossPagesAndRebuildsPlaylistsOnce(t *testing.T) {
	stubVideoMetadataScript(t)

	fullPage := make([]videodom.VideoWithoutMetadata, 0, videoMetadataReconcilePageSize)
	for fileID := 1; fileID <= videoMetadataReconcilePageSize; fileID++ {
		fullPage = append(fullPage, videodom.VideoWithoutMetadata{FileID: fileID, Path: "/v.mkv"})
	}
	repository := newFakeVideoMetadataRepository(t, videoMetadataReconcilePageSize+1)
	repository.missingByPage = [][]videodom.VideoWithoutMetadata{fullPage, {{FileID: 9999, Path: "/last.mkv"}}}
	observations := &videoReconcileObservations{}
	ctx := newVideoReconcileContext(repository, videoFilesService(-1), observations)

	if err := executeVideoMetadataReconcileStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repository.upsertedPaths) != videoMetadataReconcilePageSize+1 {
		t.Fatalf("expected %d metadata upserts, got %d", videoMetadataReconcilePageSize+1, len(repository.upsertedPaths))
	}
	for _, upserted := range repository.upsertedMetadata {
		if upserted.Classification == "" || upserted.ClassificationVersion != videodom.CurrentVideoClassificationVersion {
			t.Fatalf("metadata must be persisted already classified, got %+v", upserted)
		}
	}
	if len(observations.thumbnailFileIDs) != videoMetadataReconcilePageSize+1 {
		t.Fatalf("expected %d thumbnails, got %d", videoMetadataReconcilePageSize+1, len(observations.thumbnailFileIDs))
	}
	if observations.playlistRebuilds != 1 {
		t.Fatalf("expected exactly one playlist rebuild, got %d", observations.playlistRebuilds)
	}
}

func TestExecuteVideoMetadataReconcileStep_FailureOfOneFileDoesNotAbortTheOthers(t *testing.T) {
	stubVideoMetadataScript(t)

	repository := newFakeVideoMetadataRepository(t, 2)
	repository.missingByPage = [][]videodom.VideoWithoutMetadata{{{FileID: 1, Path: "/a.mkv"}, {FileID: 2, Path: "/b.mkv"}, {FileID: 3, Path: "/c.mkv"}}}
	observations := &videoReconcileObservations{}
	ctx := newVideoReconcileContext(repository, videoFilesService(2), observations)

	if err := executeVideoMetadataReconcileStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("per-file failures must not fail the step, got %v", err)
	}
	if len(repository.upsertedPaths) != 2 || len(observations.thumbnailFileIDs) != 2 {
		t.Fatalf("expected the two healthy files processed, got upserts=%d thumbnails=%d", len(repository.upsertedPaths), len(observations.thumbnailFileIDs))
	}
	if observations.playlistRebuilds != 1 {
		t.Fatalf("expected one playlist rebuild, got %d", observations.playlistRebuilds)
	}
}

func TestExecuteVideoMetadataReconcileStep_AllFailuresCompletesWithoutRebuild(t *testing.T) {
	stubVideoMetadataScript(t)

	repository := newFakeVideoMetadataRepository(t, 0)
	repository.missingByPage = [][]videodom.VideoWithoutMetadata{{{FileID: 1, Path: "/a.mkv"}}}
	observations := &videoReconcileObservations{}
	ctx := newVideoReconcileContext(repository, videoFilesService(1), observations)

	if err := executeVideoMetadataReconcileStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("per-file failures must not fail the step, got %v", err)
	}
	if len(observations.thumbnailFileIDs) != 0 {
		t.Fatalf("expected no thumbnails, got %d", len(observations.thumbnailFileIDs))
	}
	if observations.playlistRebuilds != 0 {
		t.Fatalf("expected no playlist rebuild when nothing was reconciled, got %d", observations.playlistRebuilds)
	}
}

func TestExecuteVideoMetadataReconcileStep_BackfillsClassificationFromStoredMetadataAndRebuilds(t *testing.T) {
	repository := newFakeVideoMetadataRepository(t, 0)
	repository.pendingByPage = [][]videodom.VideoPendingClassification{{
		{MetadataID: 7, Name: "Show S01E01.mkv", Path: "/series/show/Show S01E01.mkv", ParentPath: "/series/show", Duration: "1500.0", Height: 480},
		{MetadataID: 8, Name: "film.mkv", Path: "/x/film.mkv", ParentPath: "/x", Duration: "7200.0", Height: 1080},
	}}
	observations := &videoReconcileObservations{}
	ctx := newVideoReconcileContext(repository, &workerFilesServiceMock{}, observations)

	if err := executeVideoMetadataReconcileStep(ctx, jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.classificationByID[7] != "series" || repository.classificationByID[8] != "movie" {
		t.Fatalf("unexpected backfilled classifications: %+v", repository.classificationByID)
	}
	if len(observations.thumbnailFileIDs) != 0 {
		t.Fatalf("backfill must not reprobe or regenerate thumbnails, got %d", len(observations.thumbnailFileIDs))
	}
	if observations.playlistRebuilds != 1 {
		t.Fatalf("expected one playlist rebuild after backfill, got %d", observations.playlistRebuilds)
	}
}

func TestExecuteVideoMetadataReconcileStep_BackfillListErrorFails(t *testing.T) {
	repository := newFakeVideoMetadataRepository(t, 0)
	repository.pendingErr = errors.New("boom")
	ctx := newVideoReconcileContext(repository, &workerFilesServiceMock{}, &videoReconcileObservations{})
	if err := executeVideoMetadataReconcileStep(ctx, jobs.StepModel{}); err == nil {
		t.Fatal("expected error when listing pending classification fails")
	}
}

func TestExecuteVideoMetadataReconcileStep_BackfillUpdateFailureDoesNotAbortOrRebuild(t *testing.T) {
	repository := newFakeVideoMetadataRepository(t, 0)
	repository.pendingByPage = [][]videodom.VideoPendingClassification{{{MetadataID: 1, Name: "a.mkv", Path: "/a.mkv", ParentPath: "/"}}}
	repository.updateClassifyError = errors.New("db down")
	observations := &videoReconcileObservations{}
	ctx := newVideoReconcileContext(repository, &workerFilesServiceMock{}, observations)

	if err := executeVideoMetadataReconcileStep(ctx, jobs.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped when nothing was classified, got %v", err)
	}
	if observations.playlistRebuilds != 0 {
		t.Fatalf("expected no rebuild, got %d", observations.playlistRebuilds)
	}
}

func TestBuildVideoMetadataReconcilePlan(t *testing.T) {
	plan := buildVideoMetadataReconcilePlan()
	if err := plan.Validate(); err != nil {
		t.Fatalf("plan must be valid: %v", err)
	}
	if plan.Type != job.JobTypeVideoMetadataReconcile || plan.Priority != job.JobPriorityLow {
		t.Fatalf("unexpected plan type/priority: %+v", plan)
	}
	if len(plan.Steps) != 1 || plan.Steps[0].Type != job.StepTypeVideoMetadataReconcile {
		t.Fatalf("expected single reconcile step, got %+v", plan.Steps)
	}
}

func TestEnqueueVideoMetadataReconcileJob(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	ctx := newVideoReconcileContext(newFakeVideoMetadataRepository(t, 0), &workerFilesServiceMock{}, &videoReconcileObservations{})
	ctx.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)

	if err := enqueueVideoMetadataReconcileJob(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := enqueueVideoMetadataReconcileJob(ctx); err != nil {
		t.Fatalf("unexpected error on second enqueue: %v", err)
	}
	if len(jobsRepository.jobs) != 1 {
		t.Fatalf("expected one deduplicated reconcile job, got %d", len(jobsRepository.jobs))
	}
	if err := enqueueVideoMetadataReconcileJob(nil); err != nil {
		t.Fatalf("nil context must be a no-op, got %v", err)
	}
}
