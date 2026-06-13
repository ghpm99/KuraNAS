package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	jobsapi "nas-go/api/internal/api/v1/jobs"
	backupengine "nas-go/api/internal/worker/backup"
	jobdomain "nas-go/api/internal/worker/job"
)

type fakeBackupService struct {
	enabled bool
	opts    backupengine.Options
	optsErr error
	due     bool
	dueErr  error
}

func (f *fakeBackupService) RunOptions() (bool, backupengine.Options, error) {
	return f.enabled, f.opts, f.optsErr
}

func (f *fakeBackupService) NextRunDue(now time.Time) (bool, error) {
	return f.due, f.dueErr
}

func TestExecuteBackupRunStepRequiresService(t *testing.T) {
	err := executeBackupRunStep(&WorkerContext{}, jobsapi.StepModel{})
	if err == nil {
		t.Fatalf("expected error without backup service")
	}
}

func TestExecuteBackupRunStepSkipsWhenDisabled(t *testing.T) {
	context := &WorkerContext{BackupService: &fakeBackupService{enabled: false}}

	err := executeBackupRunStep(context, jobsapi.StepModel{})
	if !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestExecuteBackupRunStepCopiesFiles(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "doc.txt")
	if err := os.WriteFile(sourceFile, []byte("payload"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	context := &WorkerContext{BackupService: &fakeBackupService{
		enabled: true,
		opts: backupengine.Options{
			Roots:         []backupengine.Root{{Label: "main", Path: sourceDir}},
			Destination:   destDir,
			RetentionDays: 30,
		},
	}}

	if err := executeBackupRunStep(context, jobsapi.StepModel{}); err != nil {
		t.Fatalf("backup step error: %v", err)
	}

	copied := filepath.Join(destDir, backupengine.CurrentDirName, "main", "doc.txt")
	if _, err := os.Stat(copied); err != nil {
		t.Fatalf("expected backup copy at %q: %v", copied, err)
	}
}

func TestExecuteBackupRunStepSkipsWhenNothingChanged(t *testing.T) {
	sourceDir := t.TempDir()
	destDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "doc.txt")
	if err := os.WriteFile(sourceFile, []byte("payload"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	context := &WorkerContext{BackupService: &fakeBackupService{
		enabled: true,
		opts: backupengine.Options{
			Roots:         []backupengine.Root{{Label: "main", Path: sourceDir}},
			Destination:   destDir,
			RetentionDays: 30,
		},
	}}

	if err := executeBackupRunStep(context, jobsapi.StepModel{}); err != nil {
		t.Fatalf("first backup step error: %v", err)
	}

	err := executeBackupRunStep(context, jobsapi.StepModel{})
	if !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected second pass to skip, got %v", err)
	}
}

func TestMaybeEnqueueBackupRunCreatesJobWhenDue(t *testing.T) {
	repo := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repo, nil)
	context := &WorkerContext{
		BackupService:   &fakeBackupService{due: true},
		JobOrchestrator: orchestrator,
	}

	maybeEnqueueBackupRun(context, time.Now())

	jobModel, err := repo.GetJobByID(1)
	if err != nil {
		t.Fatalf("expected backup_run job to exist: %v", err)
	}
	if jobModel.Type != string(jobdomain.JobTypeBackupRun) {
		t.Fatalf("expected backup_run job type, got %q", jobModel.Type)
	}
}

func TestMaybeEnqueueBackupRunDoesNothingWhenNotDue(t *testing.T) {
	repo := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repo, nil)
	context := &WorkerContext{
		BackupService:   &fakeBackupService{due: false},
		JobOrchestrator: orchestrator,
	}

	maybeEnqueueBackupRun(context, time.Now())

	if _, err := repo.GetJobByID(1); err == nil {
		t.Fatalf("expected no job to be created")
	}
}
