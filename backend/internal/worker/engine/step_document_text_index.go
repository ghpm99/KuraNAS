package engine

import (
	"context"
	"encoding/json"
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

	if !currentIndexingPreferences(workerContext).IndexDocumentText {
		applog.Info("document text index skipped: preference disabled")
		return ErrStepSkipped
	}

	payload, err := parseStepFilePayload(step.Payload)
	if err != nil {
		return err
	}

	indexDocuments := indexPendingDocuments
	if payload.FileID > 0 {
		indexDocuments = func(repository documenttext.RepositoryInterface) (documentTextIndexOutcome, error) {
			return indexSingleDocument(repository, payload.FileID)
		}
	}

	outcome, err := indexDocuments(workerContext.DocumentTextRepository)
	if err != nil {
		return err
	}

	applog.Info("document text index finished", "indexed", outcome.indexed, "failed", outcome.failed)
	if outcome.indexed == 0 && outcome.failed == 0 {
		return ErrStepSkipped
	}
	return nil
}

func parseStepFilePayload(rawPayload []byte) (StepFilePayload, error) {
	var payload StepFilePayload
	if len(rawPayload) == 0 {
		return payload, nil
	}
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return payload, fmt.Errorf("document text index: decode payload: %w", err)
	}
	return payload, nil
}

func indexSingleDocument(repository documenttext.RepositoryInterface, fileID int) (documentTextIndexOutcome, error) {
	pendingDocument, isFound, err := repository.FindDocumentForIndexing(fileID)
	if err != nil {
		return documentTextIndexOutcome{}, fmt.Errorf("document text index: find document %d: %w", fileID, err)
	}
	if !isFound {
		return documentTextIndexOutcome{}, nil
	}

	if err := repository.UpsertDocumentText(buildDocumentText(pendingDocument)); err != nil {
		applog.Warn("document text store failed",
			"file_id", pendingDocument.FileID, "path", pendingDocument.Path, "error", err.Error())
		return documentTextIndexOutcome{failed: 1}, nil
	}
	return documentTextIndexOutcome{indexed: 1}, nil
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
		return withoutExtractedText(documentText, doctext.ErrorCodeTooLarge)
	}

	result, err := extractWithTimeout(pendingDocument)
	if err != nil {
		applog.Warn("document text extraction failed",
			"file_id", pendingDocument.FileID, "path", pendingDocument.Path, "code", doctext.ErrorCode(err), "error", err.Error())
		return withoutExtractedText(documentText, doctext.ErrorCode(err))
	}

	if result.Text == "" {
		documentText.State = documenttext.StateNameOnly
		documentText.FailureReason = documenttext.FailureReasonNoText
		return documentText
	}

	documentText.ExtractedText = result.Text
	documentText.TextLength = len([]rune(result.Text))
	documentText.IndexedBytes = int64(len(result.Text))
	documentText.Truncated = result.Truncated
	documentText.State = documenttext.StateIndexed
	if result.Truncated {
		documentText.State = documenttext.StatePartial
	}
	return documentText
}

func withoutExtractedText(documentText documenttext.DocumentTextModel, errorCode string) documenttext.DocumentTextModel {
	documentText.ErrorCode = errorCode
	documentText.FailureReason = errorCode
	documentText.State = documenttext.StateFailed
	if isNameOnlyErrorCode(errorCode) {
		documentText.State = documenttext.StateNameOnly
	}
	return documentText
}

func isNameOnlyErrorCode(errorCode string) bool {
	switch errorCode {
	case doctext.ErrorCodeTooLarge, doctext.ErrorCodeEncrypted, doctext.ErrorCodeUnsupportedFormat:
		return true
	}
	return false
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
	if !currentIndexingPreferences(workerContext).IndexDocumentText {
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

func documentReindexKey(fileID int) string {
	return fmt.Sprintf("%s:%d", documentTextIndexJobPath, fileID)
}

func buildDocumentReindexPlan(fileID int) (PlannedJob, error) {
	payload, err := marshalPayload(StepFilePayload{FileID: fileID})
	if err != nil {
		return PlannedJob{}, fmt.Errorf("marshal document reindex payload: %w", err)
	}
	return PlannedJob{
		Type:     job.JobTypeDocumentTextIndex,
		Priority: job.JobPriorityNormal,
		Scope:    job.JobScope{FileID: &fileID, ReindexKey: documentReindexKey(fileID)},
		Steps: []PlannedStep{
			{Key: "index", Type: job.StepTypeDocumentTextIndex, MaxAttempts: 1, Payload: payload},
		},
	}, nil
}
