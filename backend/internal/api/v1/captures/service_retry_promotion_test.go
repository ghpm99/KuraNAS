package captures

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
)

func TestRetryCapturePromotion(t *testing.T) {
	storageErr := errors.New("db down")
	testCases := []struct {
		name          string
		status        CaptureStatus
		lookupErr     error
		expectedErr   error
		expectedJobID int
		shouldEnqueue bool
	}{
		{"failed capture is enqueued", CaptureStatusFailed, nil, nil, 42, true},
		{"uploaded capture is rejected", CaptureStatusUploaded, nil, ErrCaptureNotRetryable, 0, false},
		{"promoting capture is rejected", CaptureStatusPromoting, nil, ErrCaptureNotRetryable, 0, false},
		{"promoted capture is rejected", CaptureStatusPromoted, nil, ErrCaptureNotRetryable, 0, false},
		{"missing capture", "", fmt.Errorf("GetCaptureByID: %w", sql.ErrNoRows), ErrCaptureNotFound, 0, false},
		{"lookup failure", "", storageErr, storageErr, 0, false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &repoMock{getByIDFn: func(id int) (CaptureModel, error) {
				return CaptureModel{ID: id, Status: testCase.status}, testCase.lookupErr
			}}
			enqueuedCaptureID := 0
			dispatcher := &uploadJobDispatcherMock{createCaptureProcessJobFn: func(captureID int) (int, error) {
				enqueuedCaptureID = captureID
				return 42, nil
			}}
			service := newServiceForTest(t, repository, dispatcher, nil)

			jobID, err := service.RetryCapturePromotion(5)

			if !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("expected error %v, got %v", testCase.expectedErr, err)
			}
			if jobID != testCase.expectedJobID {
				t.Fatalf("expected job id %d, got %d", testCase.expectedJobID, jobID)
			}
			if testCase.shouldEnqueue != (enqueuedCaptureID == 5) {
				t.Fatalf("unexpected enqueue state, enqueued capture id %d", enqueuedCaptureID)
			}
		})
	}
}

func TestRetryCapturePromotionWithoutDispatcher(t *testing.T) {
	repository := &repoMock{getByIDFn: func(id int) (CaptureModel, error) {
		return CaptureModel{ID: id, Status: CaptureStatusFailed}, nil
	}}
	service := newServiceForTest(t, repository, nil, nil)

	_, err := service.RetryCapturePromotion(5)

	if !errors.Is(err, ErrPromotionDispatcherUnavailable) {
		t.Fatalf("expected dispatcher unavailable error, got %v", err)
	}
}
