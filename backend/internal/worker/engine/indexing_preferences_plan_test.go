package engine

import (
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/internal/api/v1/configuration"
	"nas-go/api/internal/api/v1/files"
	jobdomain "nas-go/api/internal/worker/job"
)

func planStepTypes(plan PlannedJob) map[jobdomain.StepType]bool {
	stepTypes := map[jobdomain.StepType]bool{}
	for _, step := range plan.Steps {
		stepTypes[step.Type] = true
	}
	return stepTypes
}

func buildPlanWithPreferences(t *testing.T, format string, preferences configuration.IndexingPreferences) map[jobdomain.StepType]bool {
	t.Helper()
	plan, err := buildFileProcessingPlan(
		files.FileDto{Path: "/tmp/file" + format, Format: format},
		jobdomain.JobTypeFSEvent,
		jobdomain.JobPriorityLow,
		preferences,
	)
	if err != nil {
		t.Fatalf("buildFileProcessingPlan error: %v", err)
	}
	return planStepTypes(plan)
}

func TestBuildFileProcessingPlanOmitsThumbnailWithoutPreviews(t *testing.T) {
	for _, format := range []string{".jpg", ".mp4"} {
		preferences := configuration.DefaultIndexingPreferences()
		preferences.GeneratePreviews = false
		withoutPreviews := buildPlanWithPreferences(t, format, preferences)
		if withoutPreviews[jobdomain.StepTypeThumbnail] {
			t.Fatalf("%s: thumbnail step must be omitted", format)
		}
		if !withoutPreviews[jobdomain.StepTypeMetadata] || !withoutPreviews[jobdomain.StepTypePersist] {
			t.Fatalf("%s: metadata and persist must remain: %v", format, withoutPreviews)
		}

		withPreviews := buildPlanWithPreferences(t, format, configuration.DefaultIndexingPreferences())
		if !withPreviews[jobdomain.StepTypeThumbnail] {
			t.Fatalf("%s: thumbnail step must return when previews are back on", format)
		}
	}
}

func TestBuildFileProcessingPlanOmitsMetadataWithoutExtraction(t *testing.T) {
	preferences := configuration.DefaultIndexingPreferences()
	preferences.ExtractMetadata = false
	stepTypes := buildPlanWithPreferences(t, ".jpg", preferences)
	if stepTypes[jobdomain.StepTypeMetadata] {
		t.Fatalf("metadata step must be omitted: %v", stepTypes)
	}
	for _, kept := range []jobdomain.StepType{jobdomain.StepTypePersist, jobdomain.StepTypeChecksum, jobdomain.StepTypeThumbnail} {
		if !stepTypes[kept] {
			t.Fatalf("step %v must remain: %v", kept, stepTypes)
		}
	}
}

func TestDispatchWatcherChangesHonorsIndexingPreferences(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "photo.jpg")
	if err := os.WriteFile(filePath, []byte("image"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	preferences := configuration.DefaultIndexingPreferences()
	preferences.GeneratePreviews = false
	preferences.ExtractMetadata = false
	repository := newFakeJobsRepository()
	context := &WorkerContext{
		JobOrchestrator:     NewJobOrchestrator(repository, nil),
		IndexingPreferences: stubIndexingPreferences{preferences: preferences},
	}

	dispatchWatcherChanges(context, root, []string{filePath})

	if len(repository.steps) != 2 {
		t.Fatalf("expected only persist and checksum steps, got %+v", repository.steps)
	}
	for _, step := range repository.steps {
		if step.Type == string(jobdomain.StepTypeThumbnail) || step.Type == string(jobdomain.StepTypeMetadata) {
			t.Fatalf("unexpected step %q", step.Type)
		}
	}
}

func TestStartWorkersSchedulerSkipsMetadataReconcileWithoutExtraction(t *testing.T) {
	preferences := configuration.DefaultIndexingPreferences()
	preferences.ExtractMetadata = false
	repository := newFakeJobsRepository()
	context := &WorkerContext{
		JobOrchestrator:     NewJobOrchestrator(repository, nil),
		IndexingPreferences: stubIndexingPreferences{preferences: preferences},
	}

	startWorkersScheduler(context)

	for _, createdJob := range repository.jobs {
		switch createdJob.Type {
		case "image_metadata_reconcile", "audio_metadata_reconcile", "video_metadata_reconcile":
			t.Fatalf("unexpected reconcile job %q", createdJob.Type)
		}
	}
}
