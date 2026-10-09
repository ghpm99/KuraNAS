package engine

import (
	"errors"
	"fmt"

	jobs "nas-go/api/internal/api/v1/jobs"
	musicdom "nas-go/api/internal/api/v1/music"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/applog"
)

const (
	audioMetadataReconcilePageSize = 200
	audioMetadataReconcileJobPath  = "audio_metadata_reconcile"
)

type audioMetadataReconcileOutcome struct {
	reconciled int
	failed     int
}

func executeAudioMetadataReconcileStep(workerContext *WorkerContext, step jobs.StepModel) error {
	if workerContext == nil || workerContext.AudioMetadataRepository == nil || workerContext.FilesService == nil {
		return fmt.Errorf("audio metadata repository and files service are required for audio metadata reconcile step")
	}

	outcome, err := reconcileAudioWithoutMetadata(workerContext)
	if err != nil {
		return err
	}

	applog.Info("audio metadata reconcile finished", "reconciled", outcome.reconciled, "failed", outcome.failed)
	if outcome.reconciled == 0 && outcome.failed == 0 {
		return ErrStepSkipped
	}
	return nil
}

func reconcileAudioWithoutMetadata(workerContext *WorkerContext) (audioMetadataReconcileOutcome, error) {
	outcome := audioMetadataReconcileOutcome{}
	afterFileID := 0

	for {
		missingAudioFiles, err := workerContext.AudioMetadataRepository.ListAudioWithoutMetadata(afterFileID, audioMetadataReconcilePageSize)
		if err != nil {
			return outcome, fmt.Errorf("audio metadata reconcile: list audio without metadata: %w", err)
		}
		if len(missingAudioFiles) == 0 {
			return outcome, nil
		}

		for _, missingAudio := range missingAudioFiles {
			afterFileID = missingAudio.FileID

			reconcileErr := reconcileAudioMetadata(workerContext, missingAudio)
			if errors.Is(reconcileErr, ErrStepSkipped) {
				continue
			}
			if reconcileErr != nil {
				outcome.failed++
				applog.Warn("audio metadata reconcile failed",
					"file_id", missingAudio.FileID, "path", missingAudio.Path, "error", reconcileErr.Error())
				continue
			}
			outcome.reconciled++
		}

		if len(missingAudioFiles) < audioMetadataReconcilePageSize {
			return outcome, nil
		}
	}
}

func reconcileAudioMetadata(workerContext *WorkerContext, missingAudio musicdom.AudioWithoutMetadata) error {
	payload, err := marshalPayload(StepFilePayload{FileID: missingAudio.FileID, Path: missingAudio.Path})
	if err != nil {
		return err
	}
	return executeMetadataStep(workerContext, jobs.StepModel{Payload: payload})
}

func buildAudioMetadataReconcilePlan() PlannedJob {
	return PlannedJob{
		Type:     job.JobTypeAudioMetadataReconcile,
		Priority: job.JobPriorityLow,
		Scope:    job.JobScope{Path: audioMetadataReconcileJobPath},
		Steps: []PlannedStep{
			{
				Key:         "reconcile",
				Type:        job.StepTypeAudioMetadataReconcile,
				MaxAttempts: 1,
			},
		},
	}
}

func enqueueAudioMetadataReconcileJob(workerContext *WorkerContext) error {
	if workerContext == nil || workerContext.JobOrchestrator == nil || workerContext.AudioMetadataRepository == nil {
		return nil
	}

	jobID, err := workerContext.JobOrchestrator.CreateJob(buildAudioMetadataReconcilePlan())
	if err != nil {
		return err
	}

	applog.Info("audio metadata reconcile job enqueued", "job_id", jobID)
	return nil
}
