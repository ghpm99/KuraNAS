package engine

import (
	"fmt"
	"log"

	imagedom "nas-go/api/internal/api/v1/image"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/i18n"
)

// imageClassifyEnumerateBatchSize bounds how many pending images are pulled per
// keyset page while enumerating, so a large backfill never loads the whole set
// into memory at once.
const imageClassifyEnumerateBatchSize = 500

// executeImageClassifyEnumerateStep walks the images still awaiting AI
// classification and enqueues one metadata-only job per file. The metadata step
// re-runs classification (calling the AI when the toggle is on) and stamps
// ai_classified_at, which removes the file from the pending set.
func executeImageClassifyEnumerateStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil || context.ImageRepository == nil || context.JobOrchestrator == nil {
		return fmt.Errorf("image repository and job orchestrator are required for image classify enumerate step")
	}

	// Respect the toggle: with AI image classification disabled the per-file
	// metadata steps would not call the AI, so the backfill would be a no-op.
	if context.AISettings != nil {
		enabled, err := context.AISettings.IsAIImageClassificationEnabled()
		if err != nil {
			return fmt.Errorf("image classify enumerate: read AI setting: %w", err)
		}
		if !enabled {
			emitNotification(
				context,
				"info",
				i18n.GetMessage("NOTIFICATION_IMAGE_CLASSIFY_BACKFILL_DISABLED_TITLE"),
				i18n.GetMessage("NOTIFICATION_IMAGE_CLASSIFY_BACKFILL_DISABLED_MESSAGE"),
				"image_classify_backfill",
			)
			return ErrStepSkipped
		}
	}

	enqueued := 0
	afterFileID := 0
	for {
		pending, err := context.ImageRepository.ListPendingAIClassification(
			imagedom.AIClassificationConfidenceThreshold,
			afterFileID,
			imageClassifyEnumerateBatchSize,
		)
		if err != nil {
			return fmt.Errorf("image classify enumerate: list pending: %w", err)
		}
		if len(pending) == 0 {
			break
		}

		for _, item := range pending {
			afterFileID = item.FileID

			plan, planErr := buildImageClassifyMetadataPlan(item)
			if planErr != nil {
				log.Printf("[image-classify] skipping file %q: %v\n", item.Path, planErr)
				continue
			}

			jobID, createErr := context.JobOrchestrator.CreateJob(plan)
			if createErr != nil {
				return fmt.Errorf("image classify enumerate: create job for %q: %w", item.Path, createErr)
			}
			if jobID > 0 {
				enqueued++
			}
		}

		if len(pending) < imageClassifyEnumerateBatchSize {
			break
		}
	}

	emitNotification(
		context,
		"info",
		i18n.GetMessage("NOTIFICATION_IMAGE_CLASSIFY_BACKFILL_DONE_TITLE"),
		i18n.Translate("NOTIFICATION_IMAGE_CLASSIFY_BACKFILL_DONE_MESSAGE", enqueued),
		"image_classify_backfill",
	)

	if enqueued == 0 {
		return ErrStepSkipped
	}
	return nil
}

// buildImageClassifyMetadataPlan builds a metadata-only job for a single
// already-indexed image. The file is unchanged, so checksum/persist/thumbnail
// are skipped; only the metadata step (which reclassifies) runs. The path scope
// makes the orchestrator dedupe against any pending job for the same file.
func buildImageClassifyMetadataPlan(item imagedom.PendingImageClassification) (PlannedJob, error) {
	fileID := item.FileID
	payload, err := marshalPayload(StepFilePayload{
		FileID: fileID,
		Path:   item.Path,
	})
	if err != nil {
		return PlannedJob{}, fmt.Errorf("marshal image classify payload: %w", err)
	}

	return PlannedJob{
		Type:     job.JobTypeFSEvent,
		Priority: job.JobPriorityLow,
		Scope: job.JobScope{
			Path:   item.Path,
			FileID: &fileID,
		},
		Steps: []PlannedStep{
			{
				Key:         "metadata",
				Type:        job.StepTypeMetadata,
				MaxAttempts: 3,
				Payload:     payload,
			},
		},
	}, nil
}
