package engine

import (
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
)

type reindexFixture struct {
	rootDir       string
	scopedFolder  string
	siblingFolder string
	lookalikeDir  string
	workerContext *WorkerContext
	dbContext     *database.DbContext
}

func newReindexFixture(t *testing.T) reindexFixture {
	t.Helper()
	dbContext := testutil.NewPostgresDB(t, "kuranas_worker_it")
	truncateWorkerAndFiles(t, dbContext)

	rootDir := t.TempDir()
	fixture := reindexFixture{
		rootDir:       rootDir,
		scopedFolder:  filepath.Join(rootDir, "scoped"),
		siblingFolder: filepath.Join(rootDir, "sibling"),
		lookalikeDir:  filepath.Join(rootDir, "scoped-lookalike"),
		dbContext:     dbContext,
	}
	for _, folder := range []string{fixture.scopedFolder, fixture.siblingFolder, filepath.Join(fixture.scopedFolder, "nested"), fixture.lookalikeDir} {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatalf("mkdir %q: %v", folder, err)
		}
	}

	jobsRepository := jobs.NewRepository(dbContext)
	fixture.workerContext = &WorkerContext{
		FilesService:    files.NewService(files.NewRepository(dbContext), jobsRepository, nil),
		JobOrchestrator: NewJobOrchestrator(jobsRepository, nil),
	}
	return fixture
}

func writeFixtureFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func pendingJobPaths(t *testing.T, dbContext *database.DbContext) map[string]bool {
	t.Helper()
	jobsRepository := jobs.NewRepository(dbContext)
	listed, err := jobsRepository.ListJobs(jobs.JobFilter{}, 1, 100)
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	paths := map[string]bool{}
	for _, listedJob := range listed.Items {
		steps, stepsErr := jobsRepository.GetStepsByJobID(listedJob.ID)
		if stepsErr != nil {
			t.Fatalf("list steps: %v", stepsErr)
		}
		for _, step := range steps {
			payload, _ := decodeStepPayload(step.Payload)
			paths[payload.Path] = true
		}
	}
	return paths
}

func TestFolderReindexDiffVisitsOnlyTheFolder_Postgres(t *testing.T) {
	fixture := newReindexFixture(t)
	insideFile := filepath.Join(fixture.scopedFolder, "inside.txt")
	nestedFile := filepath.Join(fixture.scopedFolder, "nested", "deep.txt")
	siblingFile := filepath.Join(fixture.siblingFolder, "outside.txt")
	lookalikeFile := filepath.Join(fixture.lookalikeDir, "outside-too.txt")
	for _, path := range []string{insideFile, nestedFile, siblingFile, lookalikeFile} {
		writeFixtureFile(t, path)
	}

	payload, _ := marshalPayload(StepFilePayload{Path: fixture.scopedFolder})
	if err := executeDiffAgainstDBStep(fixture.workerContext, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("diff step: %v", err)
	}

	processedPaths := pendingJobPaths(t, fixture.dbContext)
	if !processedPaths[insideFile] || !processedPaths[nestedFile] {
		t.Fatalf("files inside the folder must be processed, got %v", processedPaths)
	}
	if processedPaths[siblingFile] || processedPaths[lookalikeFile] {
		t.Fatalf("files outside the folder must not be visited, got %v", processedPaths)
	}
}

func TestFolderReindexMarkDeletedTouchesOnlyTheFolder_Postgres(t *testing.T) {
	fixture := newReindexFixture(t)
	missingInside := filepath.Join(fixture.scopedFolder, "gone.txt")
	missingSibling := filepath.Join(fixture.siblingFolder, "gone.txt")
	missingLookalike := filepath.Join(fixture.lookalikeDir, "gone.txt")
	insertHomeFile(t, fixture.dbContext, "gone.txt", missingInside, fixture.scopedFolder)
	insertHomeFile(t, fixture.dbContext, "gone.txt", missingSibling, fixture.siblingFolder)
	insertHomeFile(t, fixture.dbContext, "gone.txt", missingLookalike, fixture.lookalikeDir)

	payload, _ := marshalPayload(StepFilePayload{Path: fixture.scopedFolder})
	if err := executeMarkDeletedStep(fixture.workerContext, jobs.StepModel{Payload: payload}); err != nil {
		t.Fatalf("mark_deleted step: %v", err)
	}

	if !isMarkedDeleted(t, fixture.dbContext, missingInside) {
		t.Fatalf("missing file inside the folder must be marked deleted")
	}
	if isMarkedDeleted(t, fixture.dbContext, missingSibling) || isMarkedDeleted(t, fixture.dbContext, missingLookalike) {
		t.Fatalf("rows outside the folder must stay untouched")
	}
}

func TestFindPendingReindexJobIDMatchesOnlyPendingJobsOfSameKey_Postgres(t *testing.T) {
	fixture := newReindexFixture(t)
	jobsRepository := jobs.NewRepository(fixture.dbContext)
	launcher := NewReindexLauncher(jobsRepository)

	launchedJobID, _, err := launcher.LaunchReindex("folder:7", []string{fixture.scopedFolder})
	if err != nil {
		t.Fatalf("launch: %v", err)
	}

	pendingJobID, hasPending, err := jobsRepository.FindPendingReindexJobID("folder:7")
	if err != nil || !hasPending || pendingJobID != launchedJobID {
		t.Fatalf("expected pending job %d, got id=%d pending=%v err=%v", launchedJobID, pendingJobID, hasPending, err)
	}
	if _, hasOther, _ := jobsRepository.FindPendingReindexJobID("folder:8"); hasOther {
		t.Fatalf("another scope must not match")
	}
}
