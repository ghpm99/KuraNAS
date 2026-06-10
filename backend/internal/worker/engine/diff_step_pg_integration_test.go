package engine

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/tests"
)

func diffFixtureScanDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("cannot resolve caller path")
	}
	// this file: backend/internal/worker/engine/ -> up 3 -> backend/
	backendRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	dir := filepath.Join(backendRoot, "tests", "files_test", "worker", "testscan")
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("abs fixture dir: %v", err)
	}
	if _, statErr := os.Stat(abs); statErr != nil {
		t.Fatalf("fixture dir not found at %s: %v", abs, statErr)
	}
	return abs
}

func countWorkerJobs(t *testing.T, ctx *database.DbContext) int {
	t.Helper()
	var count int
	err := ctx.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow("SELECT count(*) FROM worker_job").Scan(&count)
	})
	if err != nil {
		t.Fatalf("count worker_job: %v", err)
	}
	return count
}

func truncateWorkerAndFiles(t *testing.T, ctx *database.DbContext) {
	t.Helper()
	err := ctx.ExecTx(func(tx *sql.Tx) error {
		_, e := tx.Exec("TRUNCATE worker_step, worker_job, home_file RESTART IDENTITY CASCADE")
		return e
	})
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

// TestDiffStep_DoesNotReenqueueUnchangedFiles_Postgres drives the real
// executeDiffAgainstDBStep against a real PostgreSQL, a real files Service and a
// real JobOrchestrator/jobs repository — no mocks. It indexes every fixture file
// as "already processed" (size + mtime from disk), then asserts:
//
//   - a scan over unchanged files enqueues NOTHING (ErrStepSkipped, 0 jobs);
//   - touching a single file's mtime enqueues exactly ONE processing job.
//
// This is the end-to-end proof for the original complaint: files that were
// already processed and not modified must not be pushed back onto the pipeline.
func TestDiffStep_DoesNotReenqueueUnchangedFiles_Postgres(t *testing.T) {
	dbCtx := tests.NewPostgresDB(t, "kuranas_worker_it")
	truncateWorkerAndFiles(t, dbCtx)

	prevEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() { config.AppConfig.EntryPoint = prevEntryPoint })

	root := diffFixtureScanDir(t)
	config.AppConfig.EntryPoint = root

	filesRepo := files.NewRepository(dbCtx)
	jobsRepo := jobs.NewRepository(dbCtx)
	filesSvc := files.NewService(filesRepo, nil, jobsRepo, nil)
	orchestrator := NewJobOrchestrator(jobsRepo, nil)

	workerCtx := &WorkerContext{
		FilesService:    filesSvc,
		JobOrchestrator: orchestrator,
	}

	// Index every fixture file as already-processed, mirroring what the
	// pipeline stores: updated_at = the file's on-disk ModTime.
	var oneFile string
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		dto := files.FileDto{Path: path, ParentPath: filepath.Dir(path)}
		if parseErr := dto.ParseFileInfoToFileDto(info); parseErr != nil {
			return parseErr
		}
		if _, createErr := filesSvc.CreateFile(dto); createErr != nil {
			t.Fatalf("index fixture %q: %v", path, createErr)
		}
		oneFile = path
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk fixtures: %v", walkErr)
	}
	if oneFile == "" {
		t.Fatalf("no fixture files indexed under %s", root)
	}

	diffPayload, _ := marshalPayload(StepFilePayload{Path: root})

	// (1) Nothing changed on disk -> the diff must skip and create no jobs.
	err := executeDiffAgainstDBStep(workerCtx, jobs.StepModel{Payload: diffPayload})
	if err != ErrStepSkipped {
		t.Fatalf("expected ErrStepSkipped for an unchanged tree, got %v", err)
	}
	if jobs := countWorkerJobs(t, dbCtx); jobs != 0 {
		t.Fatalf("expected 0 jobs for an unchanged tree, got %d", jobs)
	}

	// (2) Touch one file's mtime -> exactly one file is now genuinely changed.
	future := time.Now().Add(2 * time.Hour)
	if chErr := os.Chtimes(oneFile, future, future); chErr != nil {
		t.Fatalf("chtimes %q: %v", oneFile, chErr)
	}

	err = executeDiffAgainstDBStep(workerCtx, jobs.StepModel{Payload: diffPayload})
	if err != nil {
		t.Fatalf("diff after touch returned error: %v", err)
	}
	if jobs := countWorkerJobs(t, dbCtx); jobs != 1 {
		t.Fatalf("expected exactly 1 job after touching a single file, got %d", jobs)
	}
}
