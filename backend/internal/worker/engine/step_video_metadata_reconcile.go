package engine

import (
	"errors"
	"fmt"

	jobs "nas-go/api/internal/api/v1/jobs"
	videodom "nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/applog"
)

const (
	videoMetadataReconcilePageSize = 200
	videoMetadataReconcileJobPath  = "video_metadata_reconcile"
)

type videoMetadataReconcileOutcome struct {
	reconciled int
	failed     int
}

func executeVideoMetadataReconcileStep(workerContext *WorkerContext, step jobs.StepModel) error {
	if workerContext == nil || workerContext.VideoMetadataRepository == nil || workerContext.FilesService == nil {
		return fmt.Errorf("video metadata repository and files service are required for video metadata reconcile step")
	}

	outcome, err := reconcileVideosWithoutMetadata(workerContext)
	if err != nil {
		return err
	}

	applog.Info("video metadata reconcile finished", "reconciled", outcome.reconciled, "failed", outcome.failed)
	if outcome.reconciled == 0 && outcome.failed == 0 {
		return ErrStepSkipped
	}

	if outcome.reconciled == 0 {
		return nil
	}
	return rebuildVideoPlaylistsAfterReconcile(workerContext)
}

func rebuildVideoPlaylistsAfterReconcile(workerContext *WorkerContext) error {
	if workerContext.VideoService == nil {
		return nil
	}
	return executePlaylistIndexStep(workerContext, jobs.StepModel{})
}

func reconcileVideosWithoutMetadata(workerContext *WorkerContext) (videoMetadataReconcileOutcome, error) {
	outcome := videoMetadataReconcileOutcome{}
	afterFileID := 0

	for {
		missingVideos, err := workerContext.VideoMetadataRepository.ListVideosWithoutMetadata(afterFileID, videoMetadataReconcilePageSize)
		if err != nil {
			return outcome, fmt.Errorf("video metadata reconcile: list videos without metadata: %w", err)
		}
		if len(missingVideos) == 0 {
			return outcome, nil
		}

		for _, missingVideo := range missingVideos {
			afterFileID = missingVideo.FileID

			reconcileErr := reconcileVideoMetadataAndThumbnail(workerContext, missingVideo)
			if errors.Is(reconcileErr, ErrStepSkipped) {
				continue
			}
			if reconcileErr != nil {
				outcome.failed++
				applog.Warn("video metadata reconcile failed",
					"file_id", missingVideo.FileID, "path", missingVideo.Path, "error", reconcileErr.Error())
				continue
			}
			outcome.reconciled++
		}

		if len(missingVideos) < videoMetadataReconcilePageSize {
			return outcome, nil
		}
	}
}

func reconcileVideoMetadataAndThumbnail(workerContext *WorkerContext, missingVideo videodom.VideoWithoutMetadata) error {
	payload, err := marshalPayload(StepFilePayload{FileID: missingVideo.FileID, Path: missingVideo.Path})
	if err != nil {
		return err
	}

	step := jobs.StepModel{Payload: payload}
	if err := executeMetadataStep(workerContext, step); err != nil {
		return err
	}
	return executeThumbnailStep(workerContext, step)
}

func buildVideoMetadataReconcilePlan() PlannedJob {
	return PlannedJob{
		Type:     job.JobTypeVideoMetadataReconcile,
		Priority: job.JobPriorityLow,
		Scope:    job.JobScope{Path: videoMetadataReconcileJobPath},
		Steps: []PlannedStep{
			{
				Key:         "reconcile",
				Type:        job.StepTypeVideoMetadataReconcile,
				MaxAttempts: 1,
			},
		},
	}
}

func enqueueVideoMetadataReconcileJob(workerContext *WorkerContext) error {
	if workerContext == nil || workerContext.JobOrchestrator == nil || workerContext.VideoMetadataRepository == nil {
		return nil
	}

	jobID, err := workerContext.JobOrchestrator.CreateJob(buildVideoMetadataReconcilePlan())
	if err != nil {
		return err
	}

	applog.Info("video metadata reconcile job enqueued", "job_id", jobID)
	return nil
}
