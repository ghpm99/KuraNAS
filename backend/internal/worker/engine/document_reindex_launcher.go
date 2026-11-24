package engine

import (
	"errors"
	"sync"

	"nas-go/api/internal/api/v1/jobs"
)

var errDocumentReindexJobNotCreated = errors.New("document reindex job was not created")

type DocumentReindexLauncher struct {
	launchMutex  sync.Mutex
	orchestrator *JobOrchestrator
	repository   jobs.RepositoryInterface
}

func NewDocumentReindexLauncher(repository jobs.RepositoryInterface) *DocumentReindexLauncher {
	return &DocumentReindexLauncher{
		orchestrator: NewJobOrchestrator(repository, nil),
		repository:   repository,
	}
}

// LaunchDocumentReindex queues the text indexing of one file and returns the
// job id; a queued or running job for the same file is returned instead of a
// new one.
func (launcher *DocumentReindexLauncher) LaunchDocumentReindex(fileID int) (int, error) {
	launcher.launchMutex.Lock()
	defer launcher.launchMutex.Unlock()

	pendingJobID, hasPendingJob, err := launcher.repository.FindPendingReindexJobID(documentReindexKey(fileID))
	if err != nil {
		return 0, err
	}
	if hasPendingJob {
		return pendingJobID, nil
	}

	plan, err := buildDocumentReindexPlan(fileID)
	if err != nil {
		return 0, err
	}
	createdJobID, err := launcher.orchestrator.CreateJob(plan)
	if err != nil {
		return 0, err
	}
	if createdJobID <= 0 {
		return 0, errDocumentReindexJobNotCreated
	}
	return createdJobID, nil
}
