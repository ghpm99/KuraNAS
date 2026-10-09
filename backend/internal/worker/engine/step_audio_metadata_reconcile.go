package engine

import (
	"errors"
	"fmt"

	jobs "nas-go/api/internal/api/v1/jobs"
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

type audioReconcileCandidate struct {
	fileID int
	path   string
}

type audioReconcileCandidateLister func(afterFileID int, limit int) ([]audioReconcileCandidate, error)

func reconcileAudioWithoutMetadata(workerContext *WorkerContext) (audioMetadataReconcileOutcome, error) {
	repository := workerContext.AudioMetadataRepository
	missingOutcome, err := reconcileAudioCandidates(workerContext, "list audio without metadata", func(afterFileID int, limit int) ([]audioReconcileCandidate, error) {
		missingAudioFiles, listErr := repository.ListAudioWithoutMetadata(afterFileID, limit)
		candidates := make([]audioReconcileCandidate, 0, len(missingAudioFiles))
		for _, missingAudio := range missingAudioFiles {
			candidates = append(candidates, audioReconcileCandidate{fileID: missingAudio.FileID, path: missingAudio.Path})
		}
		return candidates, listErr
	})
	if err != nil {
		return missingOutcome, err
	}

	staleOutcome, err := reconcileAudioCandidates(workerContext, "list audio with stale tags", func(afterFileID int, limit int) ([]audioReconcileCandidate, error) {
		staleAudioFiles, listErr := repository.ListAudioWithStaleTags(afterFileID, limit)
		candidates := make([]audioReconcileCandidate, 0, len(staleAudioFiles))
		for _, staleAudio := range staleAudioFiles {
			candidates = append(candidates, audioReconcileCandidate{fileID: staleAudio.FileID, path: staleAudio.Path})
		}
		return candidates, listErr
	})

	return audioMetadataReconcileOutcome{
		reconciled: missingOutcome.reconciled + staleOutcome.reconciled,
		failed:     missingOutcome.failed + staleOutcome.failed,
	}, err
}

func reconcileAudioCandidates(workerContext *WorkerContext, listDescription string, listCandidates audioReconcileCandidateLister) (audioMetadataReconcileOutcome, error) {
	outcome := audioMetadataReconcileOutcome{}
	afterFileID := 0

	for {
		candidates, err := listCandidates(afterFileID, audioMetadataReconcilePageSize)
		if err != nil {
			return outcome, fmt.Errorf("audio metadata reconcile: %s: %w", listDescription, err)
		}
		if len(candidates) == 0 {
			return outcome, nil
		}

		for _, candidate := range candidates {
			afterFileID = candidate.fileID

			reconcileErr := reconcileAudioMetadata(workerContext, candidate)
			if errors.Is(reconcileErr, ErrStepSkipped) {
				continue
			}
			if reconcileErr != nil {
				outcome.failed++
				applog.Warn("audio metadata reconcile failed",
					"file_id", candidate.fileID, "path", candidate.path, "error", reconcileErr.Error())
				continue
			}
			outcome.reconciled++
		}

		if len(candidates) < audioMetadataReconcilePageSize {
			return outcome, nil
		}
	}
}

func reconcileAudioMetadata(workerContext *WorkerContext, candidate audioReconcileCandidate) error {
	payload, err := marshalPayload(StepFilePayload{FileID: candidate.fileID, Path: candidate.path})
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
