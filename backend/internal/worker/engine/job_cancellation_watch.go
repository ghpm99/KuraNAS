package engine

import (
	"context"
	"errors"
	"time"

	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
)

// ErrStepCanceled is returned by a step that stopped because its job was
// canceled while it ran. The scheduler records the step as canceled, not failed.
var ErrStepCanceled = errors.New("step canceled")

var jobCancellationPollInterval = 2 * time.Second

// contextCanceledWithJob returns a context that ends when the job is canceled.
// The caller must invoke the returned stop function once the step is over.
// Without a jobs repository the context only ends through stop.
func contextCanceledWithJob(workerContext *WorkerContext, jobID int) (context.Context, context.CancelFunc) {
	jobContext, stopWatching := context.WithCancel(context.Background())
	if workerContext == nil || workerContext.JobsRepository == nil {
		return jobContext, stopWatching
	}
	go cancelWhenJobIsCanceled(jobContext, stopWatching, workerContext.JobsRepository, jobID)
	return jobContext, stopWatching
}

func cancelWhenJobIsCanceled(jobContext context.Context, cancelJobContext context.CancelFunc, repository jobs.RepositoryInterface, jobID int) {
	pollTicker := time.NewTicker(jobCancellationPollInterval)
	defer pollTicker.Stop()
	for {
		select {
		case <-jobContext.Done():
			return
		case <-pollTicker.C:
			if isJobCanceled(repository, jobID) {
				cancelJobContext()
				return
			}
		}
	}
}

func isJobCanceled(repository jobs.RepositoryInterface, jobID int) bool {
	jobModel, err := repository.GetJobByID(jobID)
	if err != nil {
		return false
	}
	return jobModel.CancelRequested || jobModel.Status == string(job.JobStatusCanceled)
}
