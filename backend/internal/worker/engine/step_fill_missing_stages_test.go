package engine

import (
	"encoding/json"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/utils"
)

const fillMissingStagesRoot = "/data/scoped"

type missingImagesRepository struct {
	imagedom.RepositoryInterface
	missingImages    []imagedom.ImageWithoutMetadata
	receivedPrefixes []string
}

func (repository *missingImagesRepository) ListImagesWithoutMetadataInPath(pathPrefix string, afterFileID int, limit int) ([]imagedom.ImageWithoutMetadata, error) {
	repository.receivedPrefixes = append(repository.receivedPrefixes, pathPrefix)
	remaining := []imagedom.ImageWithoutMetadata{}
	for _, missingImage := range repository.missingImages {
		if missingImage.FileID > afterFileID {
			remaining = append(remaining, missingImage)
		}
	}
	return remaining, nil
}

type fillMissingStagesScenario struct {
	storedFiles          []files.FileDto
	filesWithoutMetadata []imagedom.ImageWithoutMetadata
	preferences          configuration.IndexingPreferences
	fileIDsWithThumbnail map[int]bool
}

func stubThumbnailPresence(t *testing.T, fileIDsWithThumbnail map[int]bool) {
	t.Helper()
	originalImageCheck := hasCachedImageThumbnail
	originalVideoCheck := hasVideoMasterFrame
	hasCachedImageThumbnail = func(fileDto files.FileDto) bool { return fileIDsWithThumbnail[fileDto.ID] }
	hasVideoMasterFrame = func(fileID int) bool { return fileIDsWithThumbnail[fileID] }
	t.Cleanup(func() {
		hasCachedImageThumbnail = originalImageCheck
		hasVideoMasterFrame = originalVideoCheck
	})
}

func runFillMissingStages(t *testing.T, scenario fillMissingStagesScenario) (map[string][]job.StepType, *missingImagesRepository) {
	t.Helper()
	stubThumbnailPresence(t, scenario.fileIDsWithThumbnail)

	jobsRepository := newFakeJobsRepository()
	imageRepository := &missingImagesRepository{missingImages: scenario.filesWithoutMetadata}
	workerContext := &WorkerContext{
		FilesService: &workerFilesServiceMock{
			getByPathPrefixFn: func(prefix string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
				return utils.PaginationResponse[files.FileDto]{Items: scenario.storedFiles}, nil
			},
		},
		ImageRepository:     imageRepository,
		JobOrchestrator:     NewJobOrchestrator(jobsRepository, nil),
		IndexingPreferences: stubIndexingPreferences{preferences: scenario.preferences},
	}

	payload, err := marshalPayload(StepFilePayload{Path: fillMissingStagesRoot})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := executeFillMissingStagesStep(workerContext, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("executeFillMissingStagesStep returned error: %v", err)
	}
	return stagesByPath(t, jobsRepository), imageRepository
}

func stagesByPath(t *testing.T, jobsRepository *fakeJobsRepository) map[string][]job.StepType {
	t.Helper()
	listedJobs, err := jobsRepository.ListJobs(jobs.JobFilter{}, 1, 1000)
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	stagesByFilePath := map[string][]job.StepType{}
	for _, listedJob := range listedJobs.Items {
		var scope job.JobScope
		if err := json.Unmarshal(listedJob.Scope, &scope); err != nil {
			t.Fatalf("decode scope: %v", err)
		}
		steps, _ := jobsRepository.GetStepsByJobID(listedJob.ID)
		for _, step := range steps {
			stagesByFilePath[scope.Path] = append(stagesByFilePath[scope.Path], job.StepType(step.Type))
		}
	}
	return stagesByFilePath
}

func scopedImage(fileID int, name string) files.FileDto {
	return files.FileDto{ID: fileID, Name: name, Path: fillMissingStagesRoot + "/" + name, Format: ".jpg", Type: files.File, UpdatedAt: time.Now()}
}

func allPreferences() configuration.IndexingPreferences {
	return configuration.IndexingPreferences{ExtractMetadata: true, GeneratePreviews: true}
}

func assertStages(t *testing.T, stagesByFilePath map[string][]job.StepType, path string, expectedStages ...job.StepType) {
	t.Helper()
	actualStages := stagesByFilePath[path]
	if len(actualStages) != len(expectedStages) {
		t.Fatalf("%s: expected stages %v, got %v", path, expectedStages, actualStages)
	}
	for stageIndex, expectedStage := range expectedStages {
		if actualStages[stageIndex] != expectedStage {
			t.Fatalf("%s: expected stages %v, got %v", path, expectedStages, actualStages)
		}
	}
}

func TestFillMissingStages_UnchangedFileWithoutMetadataGetsOnlyMetadata(t *testing.T) {
	image := scopedImage(1, "a.jpg")
	stagesByFilePath, imageRepository := runFillMissingStages(t, fillMissingStagesScenario{
		storedFiles:          []files.FileDto{image},
		filesWithoutMetadata: []imagedom.ImageWithoutMetadata{{FileID: 1, Path: image.Path}},
		preferences:          allPreferences(),
		fileIDsWithThumbnail: map[int]bool{1: true},
	})

	assertStages(t, stagesByFilePath, image.Path, job.StepTypeMetadata)
	if len(imageRepository.receivedPrefixes) == 0 || imageRepository.receivedPrefixes[0] != fillMissingStagesRoot {
		t.Fatalf("metadata query must be scoped to the root, got %v", imageRepository.receivedPrefixes)
	}
}

func TestFillMissingStages_FileWithMetadataAndWithoutThumbnailGetsOnlyThumbnail(t *testing.T) {
	image := scopedImage(1, "a.jpg")
	stagesByFilePath, _ := runFillMissingStages(t, fillMissingStagesScenario{
		storedFiles:          []files.FileDto{image},
		preferences:          allPreferences(),
		fileIDsWithThumbnail: map[int]bool{},
	})

	assertStages(t, stagesByFilePath, image.Path, job.StepTypeThumbnail)
}

func TestFillMissingStages_FileWithoutAnyStageMissingGetsNoJob(t *testing.T) {
	image := scopedImage(1, "a.jpg")
	stagesByFilePath, _ := runFillMissingStages(t, fillMissingStagesScenario{
		storedFiles:          []files.FileDto{image},
		preferences:          allPreferences(),
		fileIDsWithThumbnail: map[int]bool{1: true},
	})

	if len(stagesByFilePath) != 0 {
		t.Fatalf("expected no job, got %v", stagesByFilePath)
	}
}

func TestFillMissingStages_FileMissingBothStagesGetsBothInOneJob(t *testing.T) {
	image := scopedImage(1, "a.jpg")
	stagesByFilePath, _ := runFillMissingStages(t, fillMissingStagesScenario{
		storedFiles:          []files.FileDto{image},
		filesWithoutMetadata: []imagedom.ImageWithoutMetadata{{FileID: 1, Path: image.Path}},
		preferences:          allPreferences(),
		fileIDsWithThumbnail: map[int]bool{},
	})

	assertStages(t, stagesByFilePath, image.Path, job.StepTypeMetadata, job.StepTypeThumbnail)
}

func TestFillMissingStages_ExtractMetadataDisabledNeverEnqueuesMetadata(t *testing.T) {
	image := scopedImage(1, "a.jpg")
	stagesByFilePath, imageRepository := runFillMissingStages(t, fillMissingStagesScenario{
		storedFiles:          []files.FileDto{image},
		filesWithoutMetadata: []imagedom.ImageWithoutMetadata{{FileID: 1, Path: image.Path}},
		preferences:          configuration.IndexingPreferences{GeneratePreviews: true},
		fileIDsWithThumbnail: map[int]bool{},
	})

	assertStages(t, stagesByFilePath, image.Path, job.StepTypeThumbnail)
	if len(imageRepository.receivedPrefixes) != 0 {
		t.Fatalf("metadata must not even be queried when extraction is disabled")
	}
}

func TestFillMissingStages_GeneratePreviewsDisabledNeverEnqueuesThumbnail(t *testing.T) {
	image := scopedImage(1, "a.jpg")
	stagesByFilePath, _ := runFillMissingStages(t, fillMissingStagesScenario{
		storedFiles:          []files.FileDto{image},
		filesWithoutMetadata: []imagedom.ImageWithoutMetadata{{FileID: 1, Path: image.Path}},
		preferences:          configuration.IndexingPreferences{ExtractMetadata: true},
		fileIDsWithThumbnail: map[int]bool{},
	})

	assertStages(t, stagesByFilePath, image.Path, job.StepTypeMetadata)
}

func TestFillMissingStages_VideoThumbnailUsesMasterFramePresence(t *testing.T) {
	video := files.FileDto{ID: 7, Name: "clip.mp4", Path: fillMissingStagesRoot + "/clip.mp4", Format: ".mp4", Type: files.File}
	stagesByFilePath, _ := runFillMissingStages(t, fillMissingStagesScenario{
		storedFiles:          []files.FileDto{video},
		preferences:          allPreferences(),
		fileIDsWithThumbnail: map[int]bool{},
	})

	assertStages(t, stagesByFilePath, video.Path, job.StepTypeThumbnail)
}

func TestFillMissingStages_NonThumbnailableFileNeverGetsThumbnail(t *testing.T) {
	document := files.FileDto{ID: 9, Name: "notes.txt", Path: fillMissingStagesRoot + "/notes.txt", Format: ".txt", Type: files.File}
	stagesByFilePath, _ := runFillMissingStages(t, fillMissingStagesScenario{
		storedFiles:          []files.FileDto{document},
		preferences:          allPreferences(),
		fileIDsWithThumbnail: map[int]bool{},
	})

	if len(stagesByFilePath) != 0 {
		t.Fatalf("expected no job, got %v", stagesByFilePath)
	}
}

func TestFillMissingStages_SkipsTieredDeletedDirectoryAndOutOfScopeFiles(t *testing.T) {
	tiered := scopedImage(1, "tiered.jpg")
	tiered.PhysicalPath = "/cold/tiered.jpg"
	deleted := scopedImage(2, "deleted.jpg")
	deleted.DeletedAt = utils.Optional[time.Time]{HasValue: true, Value: time.Now()}
	directory := scopedImage(3, "folder")
	directory.Type = files.Directory
	sibling := scopedImage(4, "sibling.jpg")
	sibling.Path = "/data/sibling/sibling.jpg"
	lookalike := scopedImage(5, "lookalike.jpg")
	lookalike.Path = "/data/scoped-lookalike/lookalike.jpg"
	inside := scopedImage(6, "inside.jpg")

	stagesByFilePath, _ := runFillMissingStages(t, fillMissingStagesScenario{
		storedFiles: []files.FileDto{tiered, deleted, directory, sibling, lookalike, inside},
		filesWithoutMetadata: []imagedom.ImageWithoutMetadata{
			{FileID: 1, Path: tiered.Path}, {FileID: 4, Path: sibling.Path}, {FileID: 5, Path: lookalike.Path},
		},
		preferences:          allPreferences(),
		fileIDsWithThumbnail: map[int]bool{6: true},
	})

	if len(stagesByFilePath) != 0 {
		t.Fatalf("expected no job for tiered, deleted, directory or out-of-scope files, got %v", stagesByFilePath)
	}
}

func TestFillMissingStages_DoesNotDuplicateJobForFileWithPendingJob(t *testing.T) {
	image := scopedImage(1, "a.jpg")
	scenario := fillMissingStagesScenario{
		storedFiles:          []files.FileDto{image},
		preferences:          allPreferences(),
		fileIDsWithThumbnail: map[int]bool{},
	}
	stubThumbnailPresence(t, scenario.fileIDsWithThumbnail)

	jobsRepository := newFakeJobsRepository()
	workerContext := &WorkerContext{
		FilesService: &workerFilesServiceMock{
			getByPathPrefixFn: func(prefix string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
				return utils.PaginationResponse[files.FileDto]{Items: scenario.storedFiles}, nil
			},
		},
		JobOrchestrator: NewJobOrchestrator(jobsRepository, nil),
	}
	payload, _ := marshalPayload(StepFilePayload{Path: fillMissingStagesRoot})

	for range 2 {
		if err := executeFillMissingStagesStep(workerContext, jobs.StepModel{Payload: payload}); err != nil {
			t.Fatalf("executeFillMissingStagesStep returned error: %v", err)
		}
	}

	storedJobs, _ := jobsRepository.ListJobs(jobs.JobFilter{}, 1, 100)
	if len(storedJobs.Items) != 1 {
		t.Fatalf("expected a single job after running the step twice, got %d", len(storedJobs.Items))
	}
}

func TestFillMissingStages_RequiresFilesServiceAndOrchestrator(t *testing.T) {
	if err := executeFillMissingStagesStep(nil, jobs.StepModel{}); err == nil {
		t.Fatal("expected error for nil context")
	}
	if err := executeFillMissingStagesStep(&WorkerContext{}, jobs.StepModel{}); err == nil {
		t.Fatal("expected error without services")
	}
}

func TestFillMissingStages_SkipsStepWithoutPath(t *testing.T) {
	workerContext := &WorkerContext{
		FilesService:    &workerFilesServiceMock{},
		JobOrchestrator: NewJobOrchestrator(newFakeJobsRepository(), nil),
	}
	if err := executeFillMissingStagesStep(workerContext, jobs.StepModel{}); err != ErrStepSkipped {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestBuildMissingStagesPlan_CarriesFileIDAndHasNoPersistStep(t *testing.T) {
	image := scopedImage(11, "a.jpg")
	plan, err := buildMissingStagesPlan(image, []job.StepType{job.StepTypeMetadata, job.StepTypeThumbnail})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("plan must be valid: %v", err)
	}
	for _, step := range plan.Steps {
		payload, _ := decodeStepPayload(step.Payload)
		if payload.FileID != 11 || payload.Path != image.Path {
			t.Fatalf("unexpected payload %+v", payload)
		}
		if len(step.DependsOn) != 0 || step.Type == job.StepTypePersist {
			t.Fatalf("partial plan must not depend on persist: %+v", step)
		}
	}
}
