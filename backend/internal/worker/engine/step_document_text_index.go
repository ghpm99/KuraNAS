package engine

import (
	"context"
	"fmt"
	"time"

	"nas-go/api/internal/api/v1/documenttext"
	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/doctext"
	"nas-go/api/pkg/utils"
)

const (
	documentTextIndexPageSize = 50
	documentTextIndexJobPath  = "document_text_index"
)

var documentTextFileTimeout = 30 * time.Second

var extractDocumentText = doctext.Extract

type documentTextIndexOutcome struct {
	indexed int
	failed  int
}

func executeDocumentTextIndexStep(workerContext *WorkerContext, step jobs.StepModel) error {
	if workerContext == nil || workerContext.DocumentTextRepository == nil {
		return fmt.Errorf("document text repository is required for document text index step")
	}

	outcome, err := indexPendingDocuments(workerContext.DocumentTextRepository)
	if err != nil {
		return err
	}

	applog.Info("document text index finished", "indexed", outcome.indexed, "failed", outcome.failed)
	if outcome.indexed == 0 && outcome.failed == 0 {
		return ErrStepSkipped
	}
	return nil
}

func indexPendingDocuments(repository documenttext.RepositoryInterface) (documentTextIndexOutcome, error) {
	outcome := documentTextIndexOutcome{}
	afterFileID := 0

	for {
		pendingDocuments, err := repository.ListPendingIndexing(afterFileID, documentTextIndexPageSize)
		if err != nil {
			return outcome, fmt.Errorf("document text index: list pending: %w", err)
		}
		if len(pendingDocuments) == 0 {
			return outcome, nil
		}

		for _, pendingDocument := range pendingDocuments {
			afterFileID = pendingDocument.FileID

			if err := repository.UpsertDocumentText(buildDocumentText(pendingDocument)); err != nil {
				outcome.failed++
				applog.Warn("document text store failed",
					"file_id", pendingDocument.FileID, "path", pendingDocument.Path, "error", err.Error())
				continue
			}
			outcome.indexed++
		}

		if len(pendingDocuments) < documentTextIndexPageSize {
			return outcome, nil
		}
	}
}

func buildDocumentText(pendingDocument documenttext.PendingDocument) documenttext.DocumentTextModel {
	documentText := documenttext.DocumentTextModel{
		FileID:          pendingDocument.FileID,
		SourceUpdatedAt: pendingDocument.UpdatedAt,
	}

	if pendingDocument.Size > doctext.MaxFileSizeBytes {
		documentText.ErrorCode = doctext.ErrorCodeTooLarge
		return documentText
	}

	result, err := extractWithTimeout(pendingDocument)
	if err != nil {
		documentText.ErrorCode = doctext.ErrorCode(err)
		applog.Warn("document text extraction failed",
			"file_id", pendingDocument.FileID, "path", pendingDocument.Path, "code", documentText.ErrorCode, "error", err.Error())
		return documentText
	}

	documentText.ExtractedText = result.Text
	documentText.TextLength = len([]rune(result.Text))
	documentText.Truncated = result.Truncated
	return documentText
}

type documentExtraction struct {
	result doctext.Result
	err    error
}

func extractWithTimeout(pendingDocument documenttext.PendingDocument) (doctext.Result, error) {
	extractionContext, cancel := context.WithTimeout(context.Background(), documentTextFileTimeout)
	defer cancel()

	extractionDone := make(chan documentExtraction, 1)
	go func() {
		result, err := extractDocumentText(extractionContext, pendingDocument.Path, pendingDocument.Format)
		extractionDone <- documentExtraction{result: result, err: err}
	}()

	select {
	case extraction := <-extractionDone:
		return extraction.result, extraction.err
	case <-extractionContext.Done():
		return doctext.Result{}, extractionContext.Err()
	}
}

func buildDocumentTextIndexPlan() PlannedJob {
	return PlannedJob{
		Type:     job.JobTypeDocumentTextIndex,
		Priority: job.JobPriorityLow,
		Scope:    job.JobScope{Path: documentTextIndexJobPath},
		Steps: []PlannedStep{
			{
				Key:         "index",
				Type:        job.StepTypeDocumentTextIndex,
				MaxAttempts: 1,
			},
		},
	}
}

func enqueueDocumentTextIndexJob(workerContext *WorkerContext) error {
	if workerContext == nil || workerContext.JobOrchestrator == nil || workerContext.DocumentTextRepository == nil {
		return nil
	}

	jobID, err := workerContext.JobOrchestrator.CreateJob(buildDocumentTextIndexPlan())
	if err != nil {
		return err
	}
	if jobID > 0 {
		applog.Info("document text index job enqueued", "job_id", jobID)
	}
	return nil
}

func enqueueDocumentTextIndexIfDocument(workerContext *WorkerContext, persistedFile files.FileDto) {
	if !utils.IsDocumentTextFormat(persistedFile.Format) && !utils.IsDocumentTextFormat(utils.ExtensionOf(persistedFile.Name)) {
		return
	}
	if err := enqueueDocumentTextIndexJob(workerContext); err != nil {
		applog.Warn("failed to enqueue document text index job", "path", persistedFile.Path, "error", err.Error())
	}
}
