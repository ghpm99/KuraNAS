package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/documenttext"
	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/doctext"
	"nas-go/api/pkg/utils"
)

type fakeDocumentTextRepository struct {
	pendingPages [][]documenttext.PendingDocument
	listCalls    []int
	listErr      error
	upsertErr    error
	stored       []documenttext.DocumentTextModel
}

func (repository *fakeDocumentTextRepository) ListPendingIndexing(afterFileID int, limit int) ([]documenttext.PendingDocument, error) {
	repository.listCalls = append(repository.listCalls, afterFileID)
	if repository.listErr != nil {
		return nil, repository.listErr
	}
	if len(repository.pendingPages) == 0 {
		return nil, nil
	}
	page := repository.pendingPages[0]
	repository.pendingPages = repository.pendingPages[1:]
	return page, nil
}

func (repository *fakeDocumentTextRepository) UpsertDocumentText(documentText documenttext.DocumentTextModel) error {
	if repository.upsertErr != nil {
		return repository.upsertErr
	}
	repository.stored = append(repository.stored, documentText)
	return nil
}

func (repository *fakeDocumentTextRepository) SearchDocuments(utils.SearchTermPatterns, int, int) ([]documenttext.DocumentMatchModel, error) {
	return nil, nil
}

func stubDocumentExtractor(t *testing.T, extractor func(ctx context.Context, path string, format string) (doctext.Result, error)) {
	t.Helper()
	original := extractDocumentText
	extractDocumentText = extractor
	t.Cleanup(func() { extractDocumentText = original })
}

func documentTextContext(repository documenttext.RepositoryInterface) *WorkerContext {
	return &WorkerContext{DocumentTextRepository: repository}
}

func TestExecuteDocumentTextIndexStep_RequiresRepository(t *testing.T) {
	if err := executeDocumentTextIndexStep(nil, jobs.StepModel{}); err == nil {
		t.Fatal("expected error for nil context")
	}
	if err := executeDocumentTextIndexStep(&WorkerContext{}, jobs.StepModel{}); err == nil {
		t.Fatal("expected error without repository")
	}
}

func TestExecuteDocumentTextIndexStep_NothingPendingIsSkipped(t *testing.T) {
	err := executeDocumentTextIndexStep(documentTextContext(&fakeDocumentTextRepository{}), jobs.StepModel{})
	if !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestExecuteDocumentTextIndexStep_ListErrorFails(t *testing.T) {
	err := executeDocumentTextIndexStep(documentTextContext(&fakeDocumentTextRepository{listErr: errors.New("boom")}), jobs.StepModel{})
	if err == nil || errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected list error, got %v", err)
	}
}

func TestExecuteDocumentTextIndexStep_PagesThroughBatchesWithKeysetCursor(t *testing.T) {
	stubDocumentExtractor(t, func(ctx context.Context, path string, format string) (doctext.Result, error) {
		return doctext.Result{Text: "texto de " + path, Truncated: path == "/truncated.txt"}, nil
	})

	fullPage := make([]documenttext.PendingDocument, 0, documentTextIndexPageSize)
	for fileID := 1; fileID <= documentTextIndexPageSize; fileID++ {
		fullPage = append(fullPage, documenttext.PendingDocument{FileID: fileID, Path: "/doc.txt", Format: ".txt"})
	}
	lastPage := []documenttext.PendingDocument{{FileID: 900, Path: "/truncated.txt", Format: ".txt", UpdatedAt: time.Unix(10, 0)}}
	repository := &fakeDocumentTextRepository{pendingPages: [][]documenttext.PendingDocument{fullPage, lastPage}}

	if err := executeDocumentTextIndexStep(documentTextContext(repository), jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repository.stored) != documentTextIndexPageSize+1 {
		t.Fatalf("expected %d stored rows, got %d", documentTextIndexPageSize+1, len(repository.stored))
	}
	if len(repository.listCalls) != 2 || repository.listCalls[0] != 0 || repository.listCalls[1] != documentTextIndexPageSize {
		t.Fatalf("unexpected keyset cursor sequence %v", repository.listCalls)
	}
	lastStored := repository.stored[len(repository.stored)-1]
	if lastStored.FileID != 900 || !lastStored.Truncated || lastStored.TextLength != len([]rune("texto de /truncated.txt")) || !lastStored.SourceUpdatedAt.Equal(time.Unix(10, 0)) {
		t.Fatalf("unexpected stored row %+v", lastStored)
	}
	if lastStored.ErrorCode != "" {
		t.Fatalf("unexpected error code %q", lastStored.ErrorCode)
	}
}

func TestExecuteDocumentTextIndexStep_ExtractionFailuresBecomeErrorRows(t *testing.T) {
	stubDocumentExtractor(t, func(ctx context.Context, path string, format string) (doctext.Result, error) {
		switch path {
		case "/binary.txt":
			return doctext.Result{}, doctext.ErrBinary
		case "/locked.pdf":
			return doctext.Result{}, doctext.ErrEncrypted
		default:
			return doctext.Result{Text: "ok"}, nil
		}
	})
	repository := &fakeDocumentTextRepository{pendingPages: [][]documenttext.PendingDocument{{
		{FileID: 1, Path: "/binary.txt", Format: ".txt"},
		{FileID: 2, Path: "/locked.pdf", Format: ".pdf"},
		{FileID: 3, Path: "/fine.txt", Format: ".txt"},
	}}}

	if err := executeDocumentTextIndexStep(documentTextContext(repository), jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errorCodes := []string{repository.stored[0].ErrorCode, repository.stored[1].ErrorCode, repository.stored[2].ErrorCode}
	if errorCodes[0] != "binary" || errorCodes[1] != "encrypted" || errorCodes[2] != "" {
		t.Fatalf("unexpected error codes %v", errorCodes)
	}
	if repository.stored[0].ExtractedText != "" || repository.stored[2].ExtractedText != "ok" {
		t.Fatalf("unexpected stored text %+v", repository.stored)
	}
}

func TestExecuteDocumentTextIndexStep_OversizedFileIsRecordedWithoutExtraction(t *testing.T) {
	stubDocumentExtractor(t, func(ctx context.Context, path string, format string) (doctext.Result, error) {
		t.Fatal("oversized file must not be extracted")
		return doctext.Result{}, nil
	})
	repository := &fakeDocumentTextRepository{pendingPages: [][]documenttext.PendingDocument{{
		{FileID: 1, Path: "/huge.pdf", Format: ".pdf", Size: doctext.MaxFileSizeBytes + 1},
	}}}

	if err := executeDocumentTextIndexStep(documentTextContext(repository), jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.stored[0].ErrorCode != "too_large" {
		t.Fatalf("expected too_large row, got %+v", repository.stored[0])
	}
}

func TestExecuteDocumentTextIndexStep_SlowExtractionTimesOut(t *testing.T) {
	originalTimeout := documentTextFileTimeout
	documentTextFileTimeout = 20 * time.Millisecond
	t.Cleanup(func() { documentTextFileTimeout = originalTimeout })

	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	stubDocumentExtractor(t, func(ctx context.Context, path string, format string) (doctext.Result, error) {
		<-release
		return doctext.Result{}, nil
	})
	repository := &fakeDocumentTextRepository{pendingPages: [][]documenttext.PendingDocument{{
		{FileID: 1, Path: "/stuck.txt", Format: ".txt"},
	}}}

	if err := executeDocumentTextIndexStep(documentTextContext(repository), jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.stored[0].ErrorCode != "timeout" {
		t.Fatalf("expected timeout row, got %+v", repository.stored[0])
	}
}

func TestExecuteDocumentTextIndexStep_StoreFailuresDoNotAbortTheBatch(t *testing.T) {
	stubDocumentExtractor(t, func(ctx context.Context, path string, format string) (doctext.Result, error) {
		return doctext.Result{Text: "x"}, nil
	})
	repository := &fakeDocumentTextRepository{
		upsertErr:    errors.New("disk full"),
		pendingPages: [][]documenttext.PendingDocument{{{FileID: 1, Path: "/a.txt", Format: ".txt"}, {FileID: 2, Path: "/b.txt", Format: ".txt"}}},
	}

	if err := executeDocumentTextIndexStep(documentTextContext(repository), jobs.StepModel{}); err != nil {
		t.Fatalf("store failures must not fail the step, got %v", err)
	}
}

func TestEnqueueDocumentTextIndexJob_IsDeduplicatedAndLowPriority(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	workerContext := documentTextContext(&fakeDocumentTextRepository{})
	workerContext.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)

	for range 2 {
		if err := enqueueDocumentTextIndexJob(workerContext); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if len(jobsRepository.jobs) != 1 {
		t.Fatalf("expected one deduplicated job, got %d", len(jobsRepository.jobs))
	}
	if err := enqueueDocumentTextIndexJob(nil); err != nil {
		t.Fatalf("nil context must be a no-op, got %v", err)
	}

	plan := buildDocumentTextIndexPlan()
	if plan.Type != job.JobTypeDocumentTextIndex || plan.Priority != job.JobPriorityLow || plan.Steps[0].Type != job.StepTypeDocumentTextIndex {
		t.Fatalf("unexpected plan %+v", plan)
	}
}

func TestEnqueueDocumentTextIndexIfDocument_OnlyForDocumentFormats(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	workerContext := documentTextContext(&fakeDocumentTextRepository{})
	workerContext.JobOrchestrator = NewJobOrchestrator(jobsRepository, nil)

	enqueueDocumentTextIndexIfDocument(workerContext, files.FileDto{Name: "photo.jpg", Format: ".jpg"})
	if len(jobsRepository.jobs) != 0 {
		t.Fatalf("image must not enqueue a document job, got %d", len(jobsRepository.jobs))
	}

	enqueueDocumentTextIndexIfDocument(workerContext, files.FileDto{Name: "Notes.MD"})
	if len(jobsRepository.jobs) != 1 {
		t.Fatalf("document must enqueue one job, got %d", len(jobsRepository.jobs))
	}
}

func TestDocumentTextIndexTypesAreValidAndRegistered(t *testing.T) {
	if !job.JobTypeDocumentTextIndex.IsValid() || !job.StepTypeDocumentTextIndex.IsValid() {
		t.Fatal("document text index types must be valid")
	}
	executors := buildStepExecutors(&WorkerContext{})
	if executors[job.StepTypeDocumentTextIndex] == nil {
		t.Fatal("document text index executor not registered")
	}
}
