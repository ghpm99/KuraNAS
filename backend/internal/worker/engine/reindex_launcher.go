package engine

import (
	"errors"
	"fmt"
	"sync"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
)

var errReindexJobNotCreated = errors.New("reindex job was not created")

// ReindexLauncher creates one job per reindex request. The job holds one
// scan, diff, mark-deleted and fill-missing-stages step chain per path, so GET /jobs/:id shows the
// progress of the whole request.
type ReindexLauncher struct {
	launchMutex  sync.Mutex
	orchestrator *JobOrchestrator
	repository   jobs.RepositoryInterface
}

func NewReindexLauncher(repository jobs.RepositoryInterface) *ReindexLauncher {
	return &ReindexLauncher{
		orchestrator: NewJobOrchestrator(repository, nil),
		repository:   repository,
	}
}

// LaunchReindex creates the job for reindexKey over paths. When a queued or
// running job already holds reindexKey it creates nothing and returns that
// job's id with isAlreadyRunning set.
func (launcher *ReindexLauncher) LaunchReindex(reindexKey string, paths []string) (jobID int, isAlreadyRunning bool, err error) {
	launcher.launchMutex.Lock()
	defer launcher.launchMutex.Unlock()

	pendingJobID, hasPendingJob, err := launcher.repository.FindPendingReindexJobID(reindexKey)
	if err != nil {
		return 0, false, err
	}
	if hasPendingJob {
		return pendingJobID, true, nil
	}

	plan, err := buildReindexPlan(reindexKey, paths)
	if err != nil {
		return 0, false, err
	}
	createdJobID, err := launcher.orchestrator.CreateJob(plan)
	if err != nil {
		return 0, false, err
	}
	if createdJobID <= 0 {
		return 0, false, errReindexJobNotCreated
	}
	return createdJobID, false, nil
}

func buildReindexPlan(reindexKey string, paths []string) (PlannedJob, error) {
	if len(paths) == 0 {
		return PlannedJob{}, fmt.Errorf("reindex requires at least one path")
	}

	steps := make([]PlannedStep, 0, len(paths)*4)
	for pathIndex, path := range paths {
		pathSteps, err := buildReindexPathSteps(pathIndex, path)
		if err != nil {
			return PlannedJob{}, err
		}
		steps = append(steps, pathSteps...)
	}

	return PlannedJob{
		Type:     job.JobTypeReindexFolder,
		Priority: job.JobPriorityNormal,
		Scope:    job.JobScope{ReindexKey: reindexKey},
		Steps:    steps,
	}, nil
}

func buildReindexPathSteps(pathIndex int, path string) ([]PlannedStep, error) {
	payload, err := marshalPayload(StepFilePayload{Path: path})
	if err != nil {
		return nil, fmt.Errorf("marshal reindex payload: %w", err)
	}

	scanKey := fmt.Sprintf("scan_filesystem_%d", pathIndex)
	diffKey := fmt.Sprintf("diff_against_db_%d", pathIndex)
	markDeletedKey := fmt.Sprintf("mark_deleted_%d", pathIndex)
	fillMissingStagesKey := fmt.Sprintf("fill_missing_stages_%d", pathIndex)
	return []PlannedStep{
		{Key: scanKey, Type: job.StepTypeScanFilesystem, MaxAttempts: 1, Payload: payload},
		{Key: diffKey, Type: job.StepTypeDiffAgainstDB, DependsOn: []string{scanKey}, MaxAttempts: 1, Payload: payload},
		{Key: markDeletedKey, Type: job.StepTypeMarkDeleted, DependsOn: []string{diffKey}, MaxAttempts: 1, Payload: payload},
		{Key: fillMissingStagesKey, Type: job.StepTypeFillMissingStages, DependsOn: []string{markDeletedKey}, MaxAttempts: 1, Payload: payload},
	}, nil
}
