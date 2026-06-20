package image

import (
	"database/sql"
	"errors"
	"fmt"

	"nas-go/api/internal/api/v1/jobs"
)

const (
	imageClassifyBackfillJobType   = "image_classify_backfill"
	imageClassifyEnumerateStepType = "image_classify_enumerate"
)

// ErrBackfillUnavailable means the jobs subsystem is not wired in, so the
// classification backfill cannot be enqueued.
var ErrBackfillUnavailable = errors.New("image classification backfill is unavailable")

// GetPendingAIClassificationCount returns how many indexed images still await AI
// classification (never classified by AI and below the heuristic threshold).
func (s *Service) GetPendingAIClassificationCount() (int, error) {
	count, err := s.Repository.CountPendingAIClassification(AIClassificationConfidenceThreshold)
	if err != nil {
		return 0, fmt.Errorf("GetPendingAIClassificationCount: %w", err)
	}
	return count, nil
}

// EnqueueClassificationBackfill enqueues a background job that reclassifies the
// images still awaiting AI classification. It is idempotent: if a backfill is
// already queued or running it returns that job's id without enqueuing another.
func (s *Service) EnqueueClassificationBackfill() (int, error) {
	if s.JobEnqueuer == nil {
		return 0, ErrBackfillUnavailable
	}

	existingID, err := s.activeBackfillJobID()
	if err != nil {
		return 0, fmt.Errorf("EnqueueClassificationBackfill check active: %w", err)
	}
	if existingID > 0 {
		return existingID, nil
	}

	var createdJob jobs.JobModel
	txErr := s.JobEnqueuer.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		job, createErr := s.JobEnqueuer.CreateJob(tx, jobs.JobModel{
			Type:            imageClassifyBackfillJobType,
			Priority:        "low",
			Scope:           []byte("{}"),
			Status:          "queued",
			CancelRequested: false,
		})
		if createErr != nil {
			return createErr
		}
		createdJob = job

		_, stepErr := s.JobEnqueuer.CreateStep(tx, jobs.StepModel{
			JobID:       createdJob.ID,
			Type:        imageClassifyEnumerateStepType,
			Status:      "queued",
			DependsOn:   []byte("[]"),
			Attempts:    0,
			MaxAttempts: 1,
			Progress:    0,
		})
		return stepErr
	})
	if txErr != nil {
		return 0, fmt.Errorf("EnqueueClassificationBackfill create job: %w", txErr)
	}

	return createdJob.ID, nil
}

// activeBackfillJobID returns the id of a queued or running backfill job, or 0
// when none is active.
func (s *Service) activeBackfillJobID() (int, error) {
	for _, status := range []string{"queued", "running"} {
		filter := jobs.JobFilter{}
		filter.Type.Set(imageClassifyBackfillJobType)
		filter.Status.Set(status)

		result, err := s.JobEnqueuer.ListJobs(filter, 1, 1)
		if err != nil {
			return 0, err
		}
		if len(result.Items) > 0 {
			return result.Items[0].ID, nil
		}
	}
	return 0, nil
}
