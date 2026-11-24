package documenttext

import (
	"errors"
	"testing"

	"nas-go/api/pkg/i18n"
)

type recordingReindexer struct {
	jobID     int
	err       error
	gotFileID int
}

func (reindexer *recordingReindexer) LaunchDocumentReindex(fileID int) (int, error) {
	reindexer.gotFileID = fileID
	return reindexer.jobID, reindexer.err
}

func TestGetIndexStateOfFailedDocumentCarriesTranslatedMessage(t *testing.T) {
	repository := &recordingRepository{isFound: true, indexState: IndexStateModel{State: StateFailed, FailureReason: "timeout"}}

	indexState, err := NewService(repository).GetIndexState(7)
	if err != nil {
		t.Fatal(err)
	}
	if indexState.FileID != 7 || indexState.State != StateFailed || indexState.Reason != "timeout" || indexState.Error != i18n.GetMessage("ERROR_DOCUMENT_INDEX_TIMEOUT") {
		t.Fatalf("unexpected state %+v", indexState)
	}
}

func TestGetIndexStateOfFailedDocumentWithUnknownReasonFallsBackToExtractFailure(t *testing.T) {
	repository := &recordingRepository{isFound: true, indexState: IndexStateModel{State: StateFailed, FailureReason: "something_new"}}

	indexState, err := NewService(repository).GetIndexState(7)
	if err != nil || indexState.Error != i18n.GetMessage("ERROR_DOCUMENT_INDEX_EXTRACT_FAILED") {
		t.Fatalf("unexpected state %+v err=%v", indexState, err)
	}
}

func TestGetIndexStateOfNonFailedDocumentHasNoError(t *testing.T) {
	repository := &recordingRepository{isFound: true, indexState: IndexStateModel{State: StatePartial, IndexedBytes: 2097152, FailureReason: "ignored"}}

	indexState, err := NewService(repository).GetIndexState(7)
	if err != nil || indexState.Reason != "" || indexState.Error != "" || indexState.IndexedBytes != 2097152 || indexState.State != StatePartial {
		t.Fatalf("unexpected state %+v err=%v", indexState, err)
	}
}

func TestGetIndexStateReportsMissingDocumentAndRepositoryErrors(t *testing.T) {
	if _, err := NewService(&recordingRepository{}).GetIndexState(7); !errors.Is(err, ErrDocumentNotFound) {
		t.Fatalf("expected ErrDocumentNotFound, got %v", err)
	}
	if _, err := NewService(&recordingRepository{findErr: errors.New("db")}).GetIndexState(7); err == nil || errors.Is(err, ErrDocumentNotFound) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestReindexDocumentLaunchesJobForDocument(t *testing.T) {
	reindexer := &recordingReindexer{jobID: 42}
	service := NewServiceWithReindexer(&recordingRepository{isFound: true}, reindexer)

	jobID, err := service.ReindexDocument(9)
	if err != nil || jobID != 42 || reindexer.gotFileID != 9 {
		t.Fatalf("jobID=%d fileID=%d err=%v", jobID, reindexer.gotFileID, err)
	}
}

func TestReindexDocumentRejectsNonDocumentWithoutLaunching(t *testing.T) {
	reindexer := &recordingReindexer{jobID: 42}
	service := NewServiceWithReindexer(&recordingRepository{}, reindexer)

	if _, err := service.ReindexDocument(9); !errors.Is(err, ErrDocumentNotFound) || reindexer.gotFileID != 0 {
		t.Fatalf("expected ErrDocumentNotFound without launch, got %v", err)
	}
}

func TestReindexDocumentPropagatesLauncherAndRepositoryErrors(t *testing.T) {
	launchFailure := NewServiceWithReindexer(&recordingRepository{isFound: true}, &recordingReindexer{err: errors.New("queue")})
	if _, err := launchFailure.ReindexDocument(9); err == nil {
		t.Fatal("expected launcher error")
	}
	lookupFailure := NewServiceWithReindexer(&recordingRepository{findErr: errors.New("db")}, &recordingReindexer{})
	if _, err := lookupFailure.ReindexDocument(9); err == nil {
		t.Fatal("expected repository error")
	}
	if _, err := NewService(&recordingRepository{isFound: true}).ReindexDocument(9); err == nil {
		t.Fatal("expected error without reindexer")
	}
}

func TestGetIndexStateOfNameOnlyDocumentCarriesReasonWithoutError(t *testing.T) {
	repository := &recordingRepository{isFound: true, indexState: IndexStateModel{State: StateNameOnly, FailureReason: "encrypted"}}

	indexState, err := NewService(repository).GetIndexState(7)
	if err != nil || indexState.Reason != "encrypted" || indexState.Error != "" {
		t.Fatalf("unexpected state %+v err=%v", indexState, err)
	}
}
