package app

import (
	"fmt"
	"path/filepath"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker"
	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/utils"
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

		job, err := s.orchestrator.CreatePlannedJob(buildUploadProcessPlan(uploadedPath))
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

func buildUploadProcessPlan(uploadedPath string) worker.JobPlan {
	fileFormatType := utils.GetFormatTypeByExtension(filepath.Ext(uploadedPath)).Type
	maxAttempts := config.AppConfig.WorkerRetryDefaultMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	steps := make([]worker.PlannedStep, 0, 5)

	checksumDependencies := []domain.StepType{}
	if shouldExtractMetadata(fileFormatType) {
		steps = append(steps, worker.PlannedStep{
			Type:        domain.StepTypeMetadata,
			MaxAttempts: maxAttempts,
		})
		checksumDependencies = append(checksumDependencies, domain.StepTypeMetadata)
	}

	steps = append(steps, worker.PlannedStep{
		Type:        domain.StepTypeChecksum,
		DependsOn:   checksumDependencies,
		MaxAttempts: maxAttempts,
	})
	steps = append(steps, worker.PlannedStep{
		Type:        domain.StepTypePersist,
		DependsOn:   []domain.StepType{domain.StepTypeChecksum},
		MaxAttempts: maxAttempts,
	})

	if shouldGenerateThumbnail(fileFormatType) {
		steps = append(steps, worker.PlannedStep{
			Type:        domain.StepTypeThumbnail,
			DependsOn:   []domain.StepType{domain.StepTypePersist},
			MaxAttempts: maxAttempts,
		})
	}

	if fileFormatType == utils.FormatTypeVideo {
		steps = append(steps, worker.PlannedStep{
			Type:        domain.StepTypePlaylistIndex,
			DependsOn:   []domain.StepType{domain.StepTypePersist},
			MaxAttempts: maxAttempts,
		})
	}

	return worker.JobPlan{
		Type:     domain.JobTypeUploadProcess,
		Priority: domain.JobPriorityHigh,
		Scope: domain.NewFileScopePayload(domain.FileScope{
			Name: filepath.Base(uploadedPath),
			Path: uploadedPath,
		}),
		Steps: steps,
	}
}

func shouldExtractMetadata(fileFormatType string) bool {
	return fileFormatType == utils.FormatTypeImage ||
		fileFormatType == utils.FormatTypeAudio ||
		fileFormatType == utils.FormatTypeVideo
}

func shouldGenerateThumbnail(fileFormatType string) bool {
	return fileFormatType == utils.FormatTypeImage ||
		fileFormatType == utils.FormatTypeVideo
}
