package engine

import (
	"database/sql"
	"path/filepath"
	"testing"

	"nas-go/api/internal/api/v1/files"
	imagedom "nas-go/api/internal/api/v1/image"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/database"
)

func insertImageMetadataRow(t *testing.T, dbContext *database.DbContext, path string) {
	t.Helper()
	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`INSERT INTO image_metadata (file_id, path) SELECT id, path FROM home_file WHERE path = $1`, path)
		return execErr
	})
	if err != nil {
		t.Fatalf("insert image_metadata for %q: %v", path, err)
	}
}

func pendingStagesByPath(t *testing.T, dbContext *database.DbContext) map[string][]job.StepType {
	t.Helper()
	jobsRepository := jobs.NewRepository(dbContext)
	listedJobs, err := jobsRepository.ListJobs(jobs.JobFilter{}, 1, 100)
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	stagesByFilePath := map[string][]job.StepType{}
	for _, listedJob := range listedJobs.Items {
		steps, stepsErr := jobsRepository.GetStepsByJobID(listedJob.ID)
		if stepsErr != nil {
			t.Fatalf("list steps: %v", stepsErr)
		}
		for _, step := range steps {
			payload, _ := decodeStepPayload(step.Payload)
			stagesByFilePath[payload.Path] = append(stagesByFilePath[payload.Path], job.StepType(step.Type))
		}
	}
	return stagesByFilePath
}

func TestFillMissingStagesStep_EnqueuesMetadataOnlyForScopedImageWithoutRow_Postgres(t *testing.T) {
	fixture := newReindexFixture(t)
	stubThumbnailPresence(t, map[int]bool{})
	hasCachedImageThumbnail = func(files.FileDto) bool { return true }

	withoutMetadata := filepath.Join(fixture.scopedFolder, "without-metadata.jpg")
	withMetadata := filepath.Join(fixture.scopedFolder, "with-metadata.jpg")
	siblingWithout := filepath.Join(fixture.siblingFolder, "sibling.jpg")
	lookalikeWithout := filepath.Join(fixture.lookalikeDir, "lookalike.jpg")
	insertHomeFile(t, fixture.dbContext, "without-metadata.jpg", withoutMetadata, fixture.scopedFolder)
	insertHomeFile(t, fixture.dbContext, "with-metadata.jpg", withMetadata, fixture.scopedFolder)
	insertHomeFile(t, fixture.dbContext, "sibling.jpg", siblingWithout, fixture.siblingFolder)
	insertHomeFile(t, fixture.dbContext, "lookalike.jpg", lookalikeWithout, fixture.lookalikeDir)
	insertImageMetadataRow(t, fixture.dbContext, withMetadata)
	fixture.workerContext.ImageRepository = imagedom.NewRepository(fixture.dbContext)

	payload, _ := marshalPayload(StepFilePayload{Path: fixture.scopedFolder})
	if err := executeFillMissingStagesStep(fixture.workerContext, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("fill_missing_stages step: %v", err)
	}

	stagesByFilePath := pendingStagesByPath(t, fixture.dbContext)
	assertStages(t, stagesByFilePath, withoutMetadata, job.StepTypeMetadata)
	if len(stagesByFilePath) != 1 {
		t.Fatalf("only the scoped image without metadata may get a job, got %v", stagesByFilePath)
	}
}
