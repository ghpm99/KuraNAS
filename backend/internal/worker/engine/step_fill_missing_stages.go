package engine

import (
	"fmt"

	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/applog"
)

const fillMissingStagesPageSize = 500

var (
	hasCachedImageThumbnail = files.HasCachedThumbnail
	hasVideoMasterFrame     = video.HasVideoMasterFrame
)

type missingMetadataFile struct {
	fileID int
	path   string
}

type missingMetadataLister func(pathPrefix string, afterFileID int, limit int) ([]missingMetadataFile, error)

func executeFillMissingStagesStep(workerContext *WorkerContext, step jobs.StepModel) error {
	if workerContext == nil || workerContext.FilesService == nil || workerContext.JobOrchestrator == nil {
		return fmt.Errorf("files service and job orchestrator are required for fill_missing_stages step")
	}

	payload, err := decodeStepPayload(step.Payload)
	if err != nil {
		return err
	}
	root := payload.Path
	if root == "" {
		return ErrStepSkipped
	}

	preferences := currentIndexingPreferences(workerContext)
	fileIDsWithoutMetadata := map[int]bool{}
	if preferences.ExtractMetadata {
		fileIDsWithoutMetadata, err = collectFileIDsWithoutMetadata(workerContext, root)
		if err != nil {
			return err
		}
	}

	enqueuedJobs := 0
	for page := 1; ; page++ {
		filesPage, listErr := workerContext.FilesService.GetFilesByPathPrefix(root, page, fillMissingStagesPageSize)
		if listErr != nil {
			return listErr
		}

		for _, candidate := range filesPage.Items {
			if !isReindexableFileInScope(root, candidate) {
				continue
			}
			missingStages := findMissingStages(candidate, preferences.GeneratePreviews, fileIDsWithoutMetadata[candidate.ID])
			isEnqueued, enqueueErr := enqueueMissingStages(workerContext, candidate, missingStages)
			if enqueueErr != nil {
				return enqueueErr
			}
			if isEnqueued {
				enqueuedJobs++
			}
		}

		if !filesPage.Pagination.HasNext {
			break
		}
	}

	applog.Info("fill missing stages finished", "root", root, "jobs", enqueuedJobs)
	if err := enqueueDocumentTextIndexJob(workerContext); err != nil {
		applog.Warn("failed to enqueue document text index job", "root", root, "error", err.Error())
	}
	return nil
}

func isReindexableFileInScope(root string, candidate files.FileDto) bool {
	if candidate.Type != files.File || candidate.DeletedAt.HasValue {
		return false
	}
	if candidate.PhysicalPath != "" {
		return false
	}
	return isSameOrDescendantPath(root, candidate.Path)
}

func findMissingStages(candidate files.FileDto, shouldGeneratePreviews bool, isMissingMetadata bool) []job.StepType {
	missingStages := []job.StepType{}
	if isMissingMetadata {
		missingStages = append(missingStages, job.StepTypeMetadata)
	}
	if shouldGeneratePreviews && isThumbnailableFile(candidate) && !hasThumbnail(candidate) {
		missingStages = append(missingStages, job.StepTypeThumbnail)
	}
	return missingStages
}

func hasThumbnail(candidate files.FileDto) bool {
	if isVideoFile(candidate) {
		return hasVideoMasterFrame(candidate.ID)
	}
	return hasCachedImageThumbnail(candidate)
}

func enqueueMissingStages(workerContext *WorkerContext, candidate files.FileDto, missingStages []job.StepType) (bool, error) {
	if len(missingStages) == 0 {
		return false, nil
	}
	plan, err := buildMissingStagesPlan(candidate, missingStages)
	if err != nil {
		return false, err
	}
	jobID, err := workerContext.JobOrchestrator.CreateJob(plan)
	if err != nil {
		return false, fmt.Errorf("enqueue missing stages for %q: %w", candidate.Path, err)
	}
	return jobID > 0, nil
}

func collectFileIDsWithoutMetadata(workerContext *WorkerContext, root string) (map[int]bool, error) {
	fileIDsWithoutMetadata := map[int]bool{}
	for _, lister := range missingMetadataListers(workerContext) {
		if err := addFileIDsWithoutMetadata(fileIDsWithoutMetadata, root, lister); err != nil {
			return nil, err
		}
	}
	return fileIDsWithoutMetadata, nil
}

func addFileIDsWithoutMetadata(fileIDsWithoutMetadata map[int]bool, root string, listMissing missingMetadataLister) error {
	afterFileID := 0
	for {
		missingFiles, err := listMissing(root, afterFileID, fillMissingStagesPageSize)
		if err != nil {
			return fmt.Errorf("list files without metadata under %q: %w", root, err)
		}
		for _, missingFile := range missingFiles {
			afterFileID = missingFile.fileID
			if isSameOrDescendantPath(root, missingFile.path) {
				fileIDsWithoutMetadata[missingFile.fileID] = true
			}
		}
		if len(missingFiles) < fillMissingStagesPageSize {
			return nil
		}
	}
}

func missingMetadataListers(workerContext *WorkerContext) []missingMetadataLister {
	listers := []missingMetadataLister{}
	if workerContext.ImageRepository != nil {
		listers = append(listers, func(pathPrefix string, afterFileID int, limit int) ([]missingMetadataFile, error) {
			missingImages, err := workerContext.ImageRepository.ListImagesWithoutMetadataInPath(pathPrefix, afterFileID, limit)
			missingFiles := make([]missingMetadataFile, 0, len(missingImages))
			for _, missingImage := range missingImages {
				missingFiles = append(missingFiles, missingMetadataFile{fileID: missingImage.FileID, path: missingImage.Path})
			}
			return missingFiles, err
		})
	}
	if workerContext.AudioMetadataRepository != nil {
		listers = append(listers, func(pathPrefix string, afterFileID int, limit int) ([]missingMetadataFile, error) {
			missingAudios, err := workerContext.AudioMetadataRepository.ListAudioWithoutMetadataInPath(pathPrefix, afterFileID, limit)
			missingFiles := make([]missingMetadataFile, 0, len(missingAudios))
			for _, missingAudio := range missingAudios {
				missingFiles = append(missingFiles, missingMetadataFile{fileID: missingAudio.FileID, path: missingAudio.Path})
			}
			return missingFiles, err
		})
	}
	if workerContext.VideoMetadataRepository != nil {
		listers = append(listers, func(pathPrefix string, afterFileID int, limit int) ([]missingMetadataFile, error) {
			missingVideos, err := workerContext.VideoMetadataRepository.ListVideosWithoutMetadataInPath(pathPrefix, afterFileID, limit)
			missingFiles := make([]missingMetadataFile, 0, len(missingVideos))
			for _, missingVideo := range missingVideos {
				missingFiles = append(missingFiles, missingMetadataFile{fileID: missingVideo.FileID, path: missingVideo.Path})
			}
			return missingFiles, err
		})
	}
	return listers
}
