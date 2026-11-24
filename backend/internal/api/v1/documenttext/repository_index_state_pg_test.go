package documenttext

import (
	"database/sql"
	"slices"
	"testing"
	"time"

	"nas-go/api/pkg/utils"
)

func softDeleteFile(t *testing.T, repositoryDb interface {
	ExecTx(func(*sql.Tx) error) error
}, fileID int) {
	t.Helper()
	err := repositoryDb.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`UPDATE home_file SET deleted_at = now() WHERE id = $1`, fileID)
		return execErr
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIndexStateIsQueuedWithoutRowOrWithStaleRow_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	withoutRow := seedDocumentFile(t, dbContext, "new.pdf", ".pdf", now)
	stale := seedDocumentFile(t, dbContext, "old.docx", ".docx", now)
	storeText(t, repository, stale, "texto", now.Add(-time.Hour), "")

	for _, fileID := range []int{withoutRow, stale} {
		indexState, isFound, err := repository.FindIndexState(fileID)
		if err != nil || !isFound || indexState.State != StateQueued || indexState.IndexedBytes != 0 || indexState.FailureReason != "" {
			t.Fatalf("file %d: %+v found=%v err=%v", fileID, indexState, isFound, err)
		}
	}
}

func TestIndexStateRoundTripsEveryStoredState_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	testCases := []struct {
		name  string
		model DocumentTextModel
	}{
		{"partial.txt", DocumentTextModel{State: StatePartial, ExtractedText: "abcd", IndexedBytes: 4, Truncated: true}},
		{"scan.pdf", DocumentTextModel{State: StateNameOnly, FailureReason: FailureReasonNoText}},
		{"locked.pdf", DocumentTextModel{State: StateNameOnly, FailureReason: "encrypted", ErrorCode: "encrypted"}},
		{"broken.docx", DocumentTextModel{State: StateFailed, FailureReason: "extract_failed", ErrorCode: "extract_failed"}},
		{"ok.md", DocumentTextModel{State: StateIndexed, ExtractedText: "ok", IndexedBytes: 2}},
	}
	for _, testCase := range testCases {
		testCase.model.FileID = seedDocumentFile(t, dbContext, testCase.name, utils.ExtensionOf(testCase.name), now)
		testCase.model.SourceUpdatedAt = now
		if err := repository.UpsertDocumentText(testCase.model); err != nil {
			t.Fatalf("%s: %v", testCase.name, err)
		}

		indexState, isFound, err := repository.FindIndexState(testCase.model.FileID)
		if err != nil || !isFound {
			t.Fatalf("%s: found=%v err=%v", testCase.name, isFound, err)
		}
		if indexState.State != testCase.model.State || indexState.IndexedBytes != testCase.model.IndexedBytes || indexState.FailureReason != testCase.model.FailureReason {
			t.Fatalf("%s: got %+v want %+v", testCase.name, indexState, testCase.model)
		}
	}
}

func TestFindIndexStateAndDocumentIgnoreNonDocumentsAndDeletedFiles_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	sourceCode := seedDocumentFile(t, dbContext, "main.go", ".go", now)
	photo := seedDocumentFile(t, dbContext, "photo.jpg", ".jpg", now)
	deleted := seedDocumentFile(t, dbContext, "gone.pdf", ".pdf", now)
	softDeleteFile(t, dbContext, deleted)

	for _, fileID := range []int{sourceCode, photo, deleted, 9999} {
		if _, isFound, err := repository.FindIndexState(fileID); err != nil || isFound {
			t.Fatalf("file %d: index state found=%v err=%v", fileID, isFound, err)
		}
		if _, isFound, err := repository.FindDocumentForIndexing(fileID); err != nil || isFound {
			t.Fatalf("file %d: document found=%v err=%v", fileID, isFound, err)
		}
	}
}

func TestFindDocumentForIndexingReturnsFileData_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	fileID := seedDocumentFile(t, dbContext, "book.epub", ".epub", now)

	document, isFound, err := repository.FindDocumentForIndexing(fileID)
	if err != nil || !isFound {
		t.Fatalf("found=%v err=%v", isFound, err)
	}
	if document.FileID != fileID || document.Path != "/docs/book.epub" || document.Format != ".epub" || document.Size != 100 || !document.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected document %+v", document)
	}
}

func TestListPendingIndexingIncludesEveryDocumentFormat_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	expectedIDs := []int{}
	for _, name := range []string{"a.doc", "b.rtf", "c.xls", "d.xlsx", "e.ods", "f.ppt", "g.pptx", "h.odp", "i.epub", "j.odt"} {
		expectedIDs = append(expectedIDs, seedDocumentFile(t, dbContext, name, utils.ExtensionOf(name), now))
	}
	pending, err := repository.ListPendingIndexing(0, 50)
	if err != nil {
		t.Fatal(err)
	}
	pendingIDs := []int{}
	for _, pendingDocument := range pending {
		pendingIDs = append(pendingIDs, pendingDocument.FileID)
	}
	if !slices.Equal(pendingIDs, expectedIDs) {
		t.Fatalf("pending %v, want %v", pendingIDs, expectedIDs)
	}
}
