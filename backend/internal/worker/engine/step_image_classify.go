package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	imagedom "nas-go/api/internal/api/v1/image"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
)

const (
	imageClassifyPageSize                = 100
	imageClassifyMaxConsecutiveFailures  = 5
	imageClassifyBackfillNotificationKey = "image_classify_backfill"
)

var imageClassifyCallTimeout = 60 * time.Second

var errImageClassifyAborted = errors.New("image classify aborted after consecutive AI failures")

type imageClassifyOutcome struct {
	classified int
	failed     int
}

func executeImageClassifyBatchStep(workerContext *WorkerContext, step jobs.StepModel) error {
	if workerContext == nil || workerContext.ImageRepository == nil || workerContext.FilesService == nil {
		return fmt.Errorf("image repository and files service are required for image classify step")
	}

	aiService := aiServiceForImageClassification(workerContext)
	if aiService == nil {
		emitNotification(
			workerContext,
			"info",
			i18n.GetMessage("NOTIFICATION_IMAGE_CLASSIFY_BACKFILL_DISABLED_TITLE"),
			i18n.GetMessage("NOTIFICATION_IMAGE_CLASSIFY_BACKFILL_DISABLED_MESSAGE"),
			imageClassifyBackfillNotificationKey,
		)
		return ErrStepSkipped
	}

	outcome, err := classifyPendingImages(workerContext, aiService)
	if err != nil {
		return err
	}

	if outcome.classified == 0 && outcome.failed == 0 {
		return ErrStepSkipped
	}

	emitNotification(
		workerContext,
		"info",
		i18n.GetMessage("NOTIFICATION_IMAGE_CLASSIFY_BACKFILL_DONE_TITLE"),
		i18n.Translate("NOTIFICATION_IMAGE_CLASSIFY_BACKFILL_DONE_MESSAGE", outcome.classified),
		imageClassifyBackfillNotificationKey,
	)
	return nil
}

func classifyPendingImages(workerContext *WorkerContext, aiService ai.ServiceInterface) (imageClassifyOutcome, error) {
	outcome := imageClassifyOutcome{}
	consecutiveFailures := 0
	afterFileID := 0

	for {
		pendingImages, err := workerContext.ImageRepository.ListPendingAIClassification(
			imagedom.AIClassificationConfidenceThreshold,
			afterFileID,
			imageClassifyPageSize,
		)
		if err != nil {
			return outcome, fmt.Errorf("image classify: list pending: %w", err)
		}
		if len(pendingImages) == 0 {
			return outcome, nil
		}

		for _, pendingImage := range pendingImages {
			afterFileID = pendingImage.FileID

			classifyErr := classifyPendingImage(workerContext, aiService, pendingImage)
			if errors.Is(classifyErr, ErrStepSkipped) {
				continue
			}
			if classifyErr != nil {
				outcome.failed++
				consecutiveFailures++
				applog.Warn("image AI classification failed",
					"file_id", pendingImage.FileID, "path", pendingImage.Path, "error", classifyErr.Error())
				if consecutiveFailures >= imageClassifyMaxConsecutiveFailures {
					return outcome, fmt.Errorf("%w: last error: %v", errImageClassifyAborted, classifyErr)
				}
				continue
			}

			outcome.classified++
			consecutiveFailures = 0
		}

		if len(pendingImages) < imageClassifyPageSize {
			return outcome, nil
		}
	}
}

func classifyPendingImage(workerContext *WorkerContext, aiService ai.ServiceInterface, pendingImage imagedom.PendingImageClassification) error {
	fileDto, err := workerContext.FilesService.GetFileById(pendingImage.FileID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrStepSkipped
	}
	if err != nil {
		return fmt.Errorf("load file: %w", err)
	}

	metadata, err := workerContext.ImageRepository.GetImageMetadataByID(pendingImage.MetadataID)
	if err != nil {
		return fmt.Errorf("load image metadata: %w", err)
	}

	callContext, cancel := context.WithTimeout(context.Background(), imageClassifyCallTimeout)
	defer cancel()

	classification, err := imagedom.ClassifyImageByAI(callContext, fileDto, metadata, aiService)
	if err != nil {
		return err
	}

	if err := workerContext.ImageRepository.UpdateAIClassification(pendingImage.FileID, classification); err != nil {
		return fmt.Errorf("store AI classification: %w", err)
	}
	return nil
}

func buildImageAIClassifyPlan() PlannedJob {
	return PlannedJob{
		Type:     job.JobTypeImageClassifyBackfill,
		Priority: job.JobPriorityLow,
		Scope:    job.JobScope{Path: imagedom.AIClassificationJobScopePath},
		Steps: []PlannedStep{
			{
				Key:         "classify",
				Type:        job.StepTypeImageClassifyBatch,
				MaxAttempts: 1,
			},
		},
	}
}

func enqueueImageAIClassifyIfNeeded(workerContext *WorkerContext, classification imagedom.ClassificationModel, filePath string) {
	if workerContext == nil || workerContext.JobOrchestrator == nil {
		return
	}
	if classification.Confidence >= imagedom.AIClassificationConfidenceThreshold {
		return
	}
	if aiServiceForImageClassification(workerContext) == nil {
		return
	}

	if _, err := workerContext.JobOrchestrator.CreateJob(buildImageAIClassifyPlan()); err != nil {
		applog.Warn("failed to enqueue image AI classification job", "path", filePath, "error", err.Error())
	}
}
