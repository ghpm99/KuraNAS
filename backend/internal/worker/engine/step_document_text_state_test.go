package engine

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/documenttext"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/doctext"
)

func buildWithExtractor(t *testing.T, size int64, format string, result doctext.Result, extractionErr error) documenttext.DocumentTextModel {
	t.Helper()
	stubDocumentExtractor(t, func(context.Context, string, string) (doctext.Result, error) {
		return result, extractionErr
	})
	return buildDocumentText(documenttext.PendingDocument{FileID: 1, Path: "/doc" + format, Format: format, Size: size, UpdatedAt: time.Unix(5, 0)})
}

func TestBuildDocumentTextStateMapping(t *testing.T) {
	testCases := map[string]struct {
		size          int64
		format        string
		result        doctext.Result
		err           error
		wantState     string
		wantReason    string
		wantErrorCode string
		wantBytes     int64
	}{
		"complete text is indexed":       {format: ".txt", result: doctext.Result{Text: "olá"}, wantState: documenttext.StateIndexed, wantBytes: 4},
		"truncated text is partial":      {format: ".docx", result: doctext.Result{Text: "abc", Truncated: true}, wantState: documenttext.StatePartial, wantBytes: 3},
		"pdf without text is name only":  {format: ".pdf", result: doctext.Result{}, wantState: documenttext.StateNameOnly, wantReason: documenttext.FailureReasonNoText},
		"pdf with password is name only": {format: ".pdf", err: doctext.ErrEncrypted, wantState: documenttext.StateNameOnly, wantReason: "encrypted", wantErrorCode: "encrypted"},
		"file above 50 MB is name only":  {size: doctext.MaxFileSizeBytes + 1, format: ".pdf", wantState: documenttext.StateNameOnly, wantReason: "too_large", wantErrorCode: "too_large"},
		"extractor reports too large":    {format: ".txt", err: doctext.ErrTooLarge, wantState: documenttext.StateNameOnly, wantReason: "too_large", wantErrorCode: "too_large"},
		"format without extractor":       {format: ".xlsx", err: doctext.ErrUnsupportedFormat, wantState: documenttext.StateNameOnly, wantReason: "unsupported_format", wantErrorCode: "unsupported_format"},
		"binary content fails":           {format: ".txt", err: doctext.ErrBinary, wantState: documenttext.StateFailed, wantReason: "binary", wantErrorCode: "binary"},
		"read error fails":               {format: ".docx", err: errors.New("zip: not a valid zip file"), wantState: documenttext.StateFailed, wantReason: "extract_failed", wantErrorCode: "extract_failed"},
		"deadline fails with timeout":    {format: ".pdf", err: context.DeadlineExceeded, wantState: documenttext.StateFailed, wantReason: "timeout", wantErrorCode: "timeout"},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			documentText := buildWithExtractor(t, testCase.size, testCase.format, testCase.result, testCase.err)
			if documentText.State != testCase.wantState || documentText.FailureReason != testCase.wantReason ||
				documentText.ErrorCode != testCase.wantErrorCode || documentText.IndexedBytes != testCase.wantBytes {
				t.Fatalf("got %+v", documentText)
			}
		})
	}
}

func TestBuildDocumentTextIndexedBytesCountsBytesNotRunes(t *testing.T) {
	documentText := buildWithExtractor(t, 10, ".txt", doctext.Result{Text: strings.Repeat("ç", 3)}, nil)
	if documentText.IndexedBytes != 6 || documentText.TextLength != 3 {
		t.Fatalf("got bytes=%d length=%d", documentText.IndexedBytes, documentText.TextLength)
	}
}

func documentReindexStep(t *testing.T, fileID int) jobs.StepModel {
	t.Helper()
	payload, err := marshalPayload(StepFilePayload{FileID: fileID})
	if err != nil {
		t.Fatal(err)
	}
	return jobs.StepModel{Payload: payload}
}

func TestExecuteDocumentTextIndexStep_FileIDPayloadReindexesOnlyThatFileRegardlessOfSourceDate(t *testing.T) {
	stubDocumentExtractor(t, func(context.Context, string, string) (doctext.Result, error) {
		return doctext.Result{Text: "novo texto"}, nil
	})
	repository := &fakeDocumentTextRepository{
		documents: map[int]documenttext.PendingDocument{
			7: {FileID: 7, Path: "/a.pdf", Format: ".pdf", UpdatedAt: time.Unix(20, 0)},
		},
		pendingPages: [][]documenttext.PendingDocument{{{FileID: 8, Path: "/other.pdf", Format: ".pdf"}}},
	}

	if err := executeDocumentTextIndexStep(documentTextContext(repository), documentReindexStep(t, 7)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repository.stored) != 1 || repository.stored[0].FileID != 7 || repository.stored[0].State != documenttext.StateIndexed {
		t.Fatalf("expected only file 7 stored, got %+v", repository.stored)
	}
	if len(repository.listCalls) != 0 {
		t.Fatalf("single-file reindex must not scan pending documents, got %v", repository.listCalls)
	}
}

func TestExecuteDocumentTextIndexStep_FileIDPayloadForMissingDocumentIsSkipped(t *testing.T) {
	repository := &fakeDocumentTextRepository{}
	err := executeDocumentTextIndexStep(documentTextContext(repository), documentReindexStep(t, 7))
	if !errors.Is(err, ErrStepSkipped) || len(repository.stored) != 0 {
		t.Fatalf("expected ErrStepSkipped without writes, got %v stored=%d", err, len(repository.stored))
	}
}

func TestExecuteDocumentTextIndexStep_InvalidPayloadFails(t *testing.T) {
	err := executeDocumentTextIndexStep(documentTextContext(&fakeDocumentTextRepository{}), jobs.StepModel{Payload: []byte("{")})
	if err == nil || errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestBuildDocumentReindexPlanScopesJobToOneFile(t *testing.T) {
	plan, err := buildDocumentReindexPlan(7)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Type != job.JobTypeDocumentTextIndex || len(plan.Steps) != 1 || plan.Steps[0].Type != job.StepTypeDocumentTextIndex {
		t.Fatalf("unexpected plan %+v", plan)
	}
	if plan.Scope.FileID == nil || *plan.Scope.FileID != 7 || plan.Scope.ReindexKey != "document_text_index:7" || plan.Scope.Path != "" {
		t.Fatalf("unexpected scope %+v", plan.Scope)
	}
	var payload StepFilePayload
	if err := json.Unmarshal(plan.Steps[0].Payload, &payload); err != nil || payload.FileID != 7 {
		t.Fatalf("unexpected payload %+v err=%v", payload, err)
	}
}

func TestDocumentReindexLauncherReturnsExistingPendingJobForRepeatedRequest(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	launcher := NewDocumentReindexLauncher(jobsRepository)

	firstJobID, err := launcher.LaunchDocumentReindex(7)
	if err != nil || firstJobID <= 0 {
		t.Fatalf("first launch: id=%d err=%v", firstJobID, err)
	}
	repeatedJobID, err := launcher.LaunchDocumentReindex(7)
	if err != nil || repeatedJobID != firstJobID {
		t.Fatalf("repeated launch: id=%d want %d err=%v", repeatedJobID, firstJobID, err)
	}
	otherJobID, err := launcher.LaunchDocumentReindex(8)
	if err != nil || otherJobID == firstJobID {
		t.Fatalf("other file must get its own job: id=%d err=%v", otherJobID, err)
	}
	if len(jobsRepository.jobs) != 2 {
		t.Fatalf("expected two jobs, got %d", len(jobsRepository.jobs))
	}
}
