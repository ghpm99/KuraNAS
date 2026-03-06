package app

import (
	"fmt"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker"
)

type uploadProcessScheduler struct {
	orchestrator *worker.JobOrchestrator
}

func newUploadProcessScheduler(repository jobs.RepositoryInterface) files.UploadProcessSchedulerInterface {
	if repository == nil {
		return nil
	}

	return &uploadProcessScheduler{
		orchestrator: worker.NewJobOrchestrator(repository, worker.NewDefaultJobPlanner()),
	}
}

func (s *uploadProcessScheduler) ScheduleUploadProcess(uploadedPaths []string) (files.UploadProcessResult, error) {
	if s == nil || s.orchestrator == nil {
		return files.UploadProcessResult{}, fmt.Errorf("upload process scheduler is not configured")
	}

	result := files.UploadProcessResult{
		Jobs: make([]files.UploadProcessJobReference, 0, len(uploadedPaths)),
	}

	for _, uploadedPath := range uploadedPaths {
		if uploadedPath == "" {
			return files.UploadProcessResult{}, fmt.Errorf("uploaded file path is required")
		}

		job, err := s.orchestrator.CreatePlannedJob(worker.BuildUploadProcessPlan(uploadedPath))
		if err != nil {
			return files.UploadProcessResult{}, err
		}

		if result.JobID == "" {
			result.JobID = job.ID
		}

		result.Jobs = append(result.Jobs, files.UploadProcessJobReference{
			Path:  uploadedPath,
			JobID: job.ID,
		})
	}

	return result, nil
}
