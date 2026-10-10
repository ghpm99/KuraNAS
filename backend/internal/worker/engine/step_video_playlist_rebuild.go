package engine

import (
	"fmt"
	"sync/atomic"

	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/utils"
)

const videoPlaylistRebuildJobPath = "video_playlist_rebuild"

type VideoPlaylistRebuildCoordinator struct {
	isRerunRequested atomic.Bool
}

func NewVideoPlaylistRebuildCoordinator() *VideoPlaylistRebuildCoordinator {
	return &VideoPlaylistRebuildCoordinator{}
}

func (coordinator *VideoPlaylistRebuildCoordinator) requestRerun() {
	if coordinator == nil {
		return
	}
	coordinator.isRerunRequested.Store(true)
}

func (coordinator *VideoPlaylistRebuildCoordinator) consumeRerunRequest() bool {
	if coordinator == nil {
		return false
	}
	return coordinator.isRerunRequested.Swap(false)
}

func buildVideoPlaylistRebuildPlan() PlannedJob {
	return PlannedJob{
		Type:     job.JobTypeVideoPlaylistRebuild,
		Priority: job.JobPriorityLow,
		Scope:    job.JobScope{Path: videoPlaylistRebuildJobPath},
		Steps: []PlannedStep{
			{
				Key:         "rebuild",
				Type:        job.StepTypeVideoPlaylistRebuild,
				MaxAttempts: 1,
			},
		},
	}
}

func enqueueVideoPlaylistRebuildIfVideo(workerContext *WorkerContext, persistedFile files.FileDto) {
	if workerContext == nil || workerContext.JobOrchestrator == nil {
		return
	}
	if utils.GetFormatTypeByExtension(persistedFile.Format).Type != utils.FormatTypeVideo {
		return
	}

	createdJobID, err := workerContext.JobOrchestrator.CreateJob(buildVideoPlaylistRebuildPlan())
	if err != nil {
		applog.Warn("failed to enqueue video playlist rebuild job", "path", persistedFile.Path, "error", err.Error())
		return
	}
	if createdJobID == 0 {
		workerContext.VideoPlaylistRebuild.requestRerun()
	}
}

func executeVideoPlaylistRebuildStep(workerContext *WorkerContext, step jobs.StepModel) error {
	if workerContext == nil || workerContext.VideoService == nil {
		return fmt.Errorf("video service is required for video playlist rebuild step")
	}

	workerContext.VideoPlaylistRebuild.consumeRerunRequest()
	for {
		if err := executePlaylistIndexStep(workerContext, step); err != nil {
			return err
		}
		if !workerContext.VideoPlaylistRebuild.consumeRerunRequest() {
			return nil
		}
	}
}
