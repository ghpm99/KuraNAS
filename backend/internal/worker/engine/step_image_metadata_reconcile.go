package engine

import (
	"errors"
	"fmt"

	imagedom "nas-go/api/internal/api/v1/image"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/applog"
)

const (
	imageMetadataReconcilePageSize = 200
	imageMetadataReconcileJobPath  = "image_metadata_reconcile"
)

type imageMetadataReconcileOutcome struct {
	reconciled int
	failed     int
}

func executeImageMetadataReconcileStep(workerContext *WorkerContext, step jobs.StepModel) error {
	if workerContext == nil || workerContext.ImageRepository == nil || workerContext.FilesService == nil {
		return fmt.Errorf("image repository and files service are required for image metadata reconcile step")
	}

	outcome, err := reconcileImagesWithoutMetadata(workerContext)
	if err != nil {
		return err
	}

	applog.Info("image metadata reconcile finished", "reconciled", outcome.reconciled, "failed", outcome.failed)
	if outcome.reconciled == 0 && outcome.failed == 0 {
		return ErrStepSkipped
	}
	return nil
}

func reconcileImagesWithoutMetadata(workerContext *WorkerContext) (imageMetadataReconcileOutcome, error) {
	outcome := imageMetadataReconcileOutcome{}
	afterFileID := 0

	for {
		missingImages, err := workerContext.ImageRepository.ListImagesWithoutMetadata(afterFileID, imageMetadataReconcilePageSize)
		if err != nil {
			return outcome, fmt.Errorf("image metadata reconcile: list images without metadata: %w", err)
		}
		if len(missingImages) == 0 {
			return outcome, nil
		}

		for _, missingImage := range missingImages {
			afterFileID = missingImage.FileID

			reconcileErr := reconcileImageMetadata(workerContext, missingImage)
			if errors.Is(reconcileErr, ErrStepSkipped) {
				continue
			}
			if reconcileErr != nil {
				outcome.failed++
				applog.Warn("image metadata reconcile failed",
					"file_id", missingImage.FileID, "path", missingImage.Path, "error", reconcileErr.Error())
				continue
			}
			outcome.reconciled++
		}

		if len(missingImages) < imageMetadataReconcilePageSize {
			return outcome, nil
		}
	}
}

func reconcileImageMetadata(workerContext *WorkerContext, missingImage imagedom.ImageWithoutMetadata) error {
	payload, err := marshalPayload(StepFilePayload{FileID: missingImage.FileID, Path: missingImage.Path})
	if err != nil {
		return err
	}
	return executeMetadataStep(workerContext, jobs.StepModel{Payload: payload})
}

func buildImageMetadataReconcilePlan() PlannedJob {
	return PlannedJob{
		Type:     job.JobTypeImageMetadataReconcile,
		Priority: job.JobPriorityLow,
		Scope:    job.JobScope{Path: imageMetadataReconcileJobPath},
		Steps: []PlannedStep{
			{
				Key:         "reconcile",
				Type:        job.StepTypeImageMetadataReconcile,
				MaxAttempts: 1,
			},
		},
	}
}

func enqueueImageMetadataReconcileJob(workerContext *WorkerContext) error {
	if workerContext == nil || workerContext.JobOrchestrator == nil || workerContext.ImageRepository == nil {
		return nil
	}

	jobID, err := workerContext.JobOrchestrator.CreateJob(buildImageMetadataReconcilePlan())
	if err != nil {
		return err
	}

	applog.Info("image metadata reconcile job enqueued", "job_id", jobID)
	return nil
}
