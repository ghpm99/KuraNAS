package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	jobsapi "nas-go/api/internal/api/v1/jobs"
	jobdomain "nas-go/api/internal/worker/job"
	tieringengine "nas-go/api/internal/worker/tiering"
)

type fakeTieringService struct {
	enabled    bool
	coldDir    string
	promotions []tieringengine.Promotion
	demotions  []tieringengine.Demotion
	planErr    error
	due        bool
	dueErr     error
	setCalls   []int
}

func (f *fakeTieringService) MigrationPlan(now time.Time) (bool, string, []tieringengine.Promotion, []tieringengine.Demotion, error) {
	return f.enabled, f.coldDir, f.promotions, f.demotions, f.planErr
}

func (f *fakeTieringService) SetPhysicalPath(fileID int, physicalPath string) error {
	f.setCalls = append(f.setCalls, fileID)
	return nil
}

func (f *fakeTieringService) NextRunDue(now time.Time) (bool, error) {
	return f.due, f.dueErr
}

func TestExecuteTierMigrationStepRequiresService(t *testing.T) {
	err := executeTierMigrationStep(&WorkerContext{}, jobsapi.StepModel{})
	if err == nil {
		t.Fatalf("expected error without tiering service")
	}
}

func TestExecuteTierMigrationStepSkipsWhenDisabled(t *testing.T) {
	context := &WorkerContext{TieringService: &fakeTieringService{enabled: false}}

	err := executeTierMigrationStep(context, jobsapi.StepModel{})
	if !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestExecuteTierMigrationStepSkipsWhenNoWork(t *testing.T) {
	context := &WorkerContext{TieringService: &fakeTieringService{enabled: true, coldDir: "/mnt/cold"}}

	err := executeTierMigrationStep(context, jobsapi.StepModel{})
	if !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestExecuteTierMigrationStepDemotesFile(t *testing.T) {
	dir := t.TempDir()
	hot := filepath.Join(dir, "hot", "a.txt")
	cold := filepath.Join(dir, "cold", "a.txt")
	if err := os.MkdirAll(filepath.Dir(hot), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(hot, []byte("payload"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	service := &fakeTieringService{
		enabled:   true,
		coldDir:   filepath.Join(dir, "cold"),
		demotions: []tieringengine.Demotion{{FileID: 1, HotPath: hot, ColdPath: cold}},
	}

	if err := executeTierMigrationStep(&WorkerContext{TieringService: service}, jobsapi.StepModel{}); err != nil {
		t.Fatalf("tier step error: %v", err)
	}
	if _, err := os.Stat(cold); err != nil {
		t.Fatalf("expected cold copy at %q: %v", cold, err)
	}
	if len(service.setCalls) != 1 || service.setCalls[0] != 1 {
		t.Fatalf("expected physical_path recorded for file 1, got %v", service.setCalls)
	}
}

func TestExecuteTierMigrationStepEmitsFailureOnPlanError(t *testing.T) {
	service := &fakeTieringService{planErr: errors.New("cold volume offline")}
	context := &WorkerContext{TieringService: service}

	err := executeTierMigrationStep(context, jobsapi.StepModel{})
	if err == nil {
		t.Fatalf("expected the plan error to propagate")
	}
}

func TestMaybeEnqueueTierMigrationSilentOnScheduleError(t *testing.T) {
	repo := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repo, nil)
	context := &WorkerContext{
		TieringService:  &fakeTieringService{dueErr: errors.New("db down")},
		JobOrchestrator: orchestrator,
	}

	maybeEnqueueTierMigration(context, time.Now())

	if _, err := repo.GetJobByID(1); err == nil {
		t.Fatalf("a schedule error must not enqueue a job")
	}
}

func TestMaybeEnqueueTierMigrationCreatesJobWhenDue(t *testing.T) {
	repo := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repo, nil)
	context := &WorkerContext{
		TieringService:  &fakeTieringService{due: true},
		JobOrchestrator: orchestrator,
	}

	maybeEnqueueTierMigration(context, time.Now())

	jobModel, err := repo.GetJobByID(1)
	if err != nil {
		t.Fatalf("expected tier_migration job to exist: %v", err)
	}
	if jobModel.Type != string(jobdomain.JobTypeTierMigration) {
		t.Fatalf("expected tier_migration job type, got %q", jobModel.Type)
	}
}

func TestMaybeEnqueueTierMigrationDoesNothingWhenNotDue(t *testing.T) {
	repo := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repo, nil)
	context := &WorkerContext{
		TieringService:  &fakeTieringService{due: false},
		JobOrchestrator: orchestrator,
	}

	maybeEnqueueTierMigration(context, time.Now())

	if _, err := repo.GetJobByID(1); err == nil {
		t.Fatalf("expected no job to be created")
	}
}
