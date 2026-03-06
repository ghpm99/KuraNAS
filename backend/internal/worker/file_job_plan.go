package worker

import (
	"path/filepath"

	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/utils"
)

func BuildUploadProcessPlan(filePath string) JobPlan {
	return buildFileJobPlan(filePath, domain.JobTypeUploadProcess, domain.JobPriorityHigh)
}

func BuildFileEventProcessingPlan(filePath string, priority domain.JobPriority) JobPlan {
	return buildFileJobPlan(filePath, domain.JobTypeFSEvent, priority)
}

func buildFileJobPlan(filePath string, jobType domain.JobType, priority domain.JobPriority) JobPlan {
	fileFormatType := utils.GetFormatTypeByExtension(filepath.Ext(filePath)).Type
	maxAttempts := defaultPlannedStepMaxAttempts()

	steps := make([]PlannedStep, 0, 5)
	checksumDependencies := []domain.StepType{}

	if shouldExtractMetadata(fileFormatType) {
		steps = append(steps, PlannedStep{
			Type:        domain.StepTypeMetadata,
			MaxAttempts: maxAttempts,
		})
		checksumDependencies = append(checksumDependencies, domain.StepTypeMetadata)
	}

	steps = append(steps, PlannedStep{
		Type:        domain.StepTypeChecksum,
		DependsOn:   checksumDependencies,
		MaxAttempts: maxAttempts,
	})
	steps = append(steps, PlannedStep{
		Type:        domain.StepTypePersist,
		DependsOn:   []domain.StepType{domain.StepTypeChecksum},
		MaxAttempts: maxAttempts,
	})

	if shouldGenerateThumbnail(fileFormatType) {
		steps = append(steps, PlannedStep{
			Type:        domain.StepTypeThumbnail,
			DependsOn:   []domain.StepType{domain.StepTypePersist},
			MaxAttempts: maxAttempts,
		})
	}

	if fileFormatType == utils.FormatTypeVideo {
		steps = append(steps, PlannedStep{
			Type:        domain.StepTypePlaylistIndex,
			DependsOn:   []domain.StepType{domain.StepTypePersist},
			MaxAttempts: maxAttempts,
		})
	}

	return JobPlan{
		Type:     jobType,
		Priority: priority,
		Scope: domain.NewFileScopePayload(domain.FileScope{
			Name: filepath.Base(filePath),
			Path: filePath,
		}),
		Steps: steps,
	}
}
