package engine

import (
	"errors"
	"testing"

	jobs "nas-go/api/internal/api/v1/jobs"
)

type capturePromoterMock struct {
	promotedID int
	err        error
}

func (m *capturePromoterMock) PromoteCapture(captureID int) error {
	m.promotedID = captureID
	return m.err
}

func TestExecuteCapturePromoteStep(t *testing.T) {
	t.Run("requires captures service", func(t *testing.T) {
		if err := executeCapturePromoteStep(&WorkerContext{}, jobs.StepModel{}); err == nil {
			t.Fatal("expected error when captures service is missing")
		}
	})

	t.Run("skips without capture id", func(t *testing.T) {
		ctx := &WorkerContext{CapturesService: &capturePromoterMock{}}
		if err := executeCapturePromoteStep(ctx, jobs.StepModel{Payload: []byte(`{}`)}); !errors.Is(err, ErrStepSkipped) {
			t.Fatalf("expected ErrStepSkipped, got %v", err)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		ctx := &WorkerContext{CapturesService: &capturePromoterMock{}}
		if err := executeCapturePromoteStep(ctx, jobs.StepModel{Payload: []byte(`{not-json`)}); err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("delegates to the captures service", func(t *testing.T) {
		mock := &capturePromoterMock{}
		ctx := &WorkerContext{CapturesService: mock}
		if err := executeCapturePromoteStep(ctx, jobs.StepModel{Payload: []byte(`{"capture_id":77}`)}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mock.promotedID != 77 {
			t.Fatalf("expected promotion of capture 77, got %d", mock.promotedID)
		}
	})

	t.Run("propagates promotion error", func(t *testing.T) {
		ctx := &WorkerContext{CapturesService: &capturePromoterMock{err: errors.New("boom")}}
		if err := executeCapturePromoteStep(ctx, jobs.StepModel{Payload: []byte(`{"capture_id":1}`)}); err == nil {
			t.Fatal("expected promotion error to propagate")
		}
	})
}
