package documenttext

import (
	"testing"
	"time"
)

func TestFindDocumentTextReturnsStoredTextAndPartialFlag_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	indexedID := seedDocumentFile(t, dbContext, "ok.md", ".md", now)
	partialID := seedDocumentFile(t, dbContext, "big.txt", ".txt", now)
	for _, model := range []DocumentTextModel{
		{FileID: indexedID, State: StateIndexed, ExtractedText: "completo", IndexedBytes: 8},
		{FileID: partialID, State: StatePartial, ExtractedText: "parcial", IndexedBytes: 7, Truncated: true},
	} {
		model.SourceUpdatedAt = now
		if err := repository.UpsertDocumentText(model); err != nil {
			t.Fatal(err)
		}
	}

	testCases := map[int]DocumentTextDto{
		indexedID: {Text: "completo", IsPartial: false},
		partialID: {Text: "parcial", IsPartial: true},
	}
	for fileID, want := range testCases {
		got, err := service.GetDocumentText(fileID)
		if err != nil || got != want {
			t.Fatalf("file %d: got %+v want %+v err=%v", fileID, got, want, err)
		}
	}
}

func TestFindDocumentTextIsEmptyWithoutRowOrWithStaleRow_Postgres(t *testing.T) {
	repository, service, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	withoutRow := seedDocumentFile(t, dbContext, "new.pdf", ".pdf", now)
	stale := seedDocumentFile(t, dbContext, "old.docx", ".docx", now)
	if err := repository.UpsertDocumentText(DocumentTextModel{FileID: stale, State: StatePartial, ExtractedText: "antigo", SourceUpdatedAt: now.Add(-time.Hour)}); err != nil {
		t.Fatal(err)
	}

	for _, fileID := range []int{withoutRow, stale} {
		got, err := service.GetDocumentText(fileID)
		if err != nil || got != (DocumentTextDto{}) {
			t.Fatalf("file %d: got %+v err=%v", fileID, got, err)
		}
	}
}

func TestFindDocumentTextIgnoresNonDocumentsAndDeletedFiles_Postgres(t *testing.T) {
	repository, _, dbContext := newDocumentTextPostgres(t)
	now := time.Now().UTC().Truncate(time.Second)
	photo := seedDocumentFile(t, dbContext, "photo.jpg", ".jpg", now)
	deleted := seedDocumentFile(t, dbContext, "gone.pdf", ".pdf", now)
	softDeleteFile(t, dbContext, deleted)

	for _, fileID := range []int{photo, deleted, 9999} {
		if _, isFound, err := repository.FindDocumentText(fileID); err != nil || isFound {
			t.Fatalf("file %d: found=%v err=%v", fileID, isFound, err)
		}
	}
}
