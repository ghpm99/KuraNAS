package engine

import (
	"encoding/json"
	"fmt"

	jobs "nas-go/api/internal/api/v1/jobs"
)

type capturePromotePayload struct {
	CaptureID int `json:"capture_id"`
}

func executeCapturePromoteStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil || context.CapturesService == nil {
		return fmt.Errorf("captures service is required for capture promote step")
	}

	var payload capturePromotePayload
	if len(step.Payload) > 0 {
		if err := json.Unmarshal(step.Payload, &payload); err != nil {
			return fmt.Errorf("decode capture promote payload: %w", err)
		}
	}
	if payload.CaptureID <= 0 {
		return ErrStepSkipped
	}

	return context.CapturesService.PromoteCapture(payload.CaptureID)
}
