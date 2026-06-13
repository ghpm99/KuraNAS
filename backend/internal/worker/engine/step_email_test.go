package engine

import (
	"errors"
	"testing"

	emailapi "nas-go/api/internal/api/v1/email"
	jobsapi "nas-go/api/internal/api/v1/jobs"
	jobdomain "nas-go/api/internal/worker/job"
)

type fakeEmailService struct {
	stats        emailapi.SyncStats
	syncErr      error
	flagged      int
	prefilterErr error
	purged       int
	purgeErr     error
	analyzeStats emailapi.AnalyzeStats
	analyzeErr   error
}

func (f *fakeEmailService) SyncEnabledAccounts() (emailapi.SyncStats, error) {
	return f.stats, f.syncErr
}
func (f *fakeEmailService) PrefilterPending() (int, error) { return f.flagged, f.prefilterErr }
func (f *fakeEmailService) PurgeExpired() (int, error)     { return f.purged, f.purgeErr }
func (f *fakeEmailService) AnalyzePending() (emailapi.AnalyzeStats, error) {
	return f.analyzeStats, f.analyzeErr
}

func TestExecuteEmailFetchStepSkipsWhenFeatureOff(t *testing.T) {
	if err := executeEmailFetchStep(&WorkerContext{}, jobsapi.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped when no email service, got %v", err)
	}
}

func TestExecuteEmailFetchStepSkipsWhenNothingFetched(t *testing.T) {
	context := &WorkerContext{EmailService: &fakeEmailService{}}
	if err := executeEmailFetchStep(context, jobsapi.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestExecuteEmailFetchStepSucceedsAndDoesNotFailOnReauth(t *testing.T) {
	context := &WorkerContext{EmailService: &fakeEmailService{
		stats: emailapi.SyncStats{Accounts: 2, Fetched: 3, ReauthRequired: []string{"stale@gmail.com"}},
	}}
	if err := executeEmailFetchStep(context, jobsapi.StepModel{}); err != nil {
		t.Fatalf("reauth must not fail the step, got %v", err)
	}
}

func TestExecuteEmailFetchStepPropagatesError(t *testing.T) {
	context := &WorkerContext{EmailService: &fakeEmailService{syncErr: errors.New("boom")}}
	if err := executeEmailFetchStep(context, jobsapi.StepModel{}); err == nil {
		t.Fatal("expected sync error to propagate")
	}
}

func TestExecuteEmailPrefilterStepRunsAndPurges(t *testing.T) {
	context := &WorkerContext{EmailService: &fakeEmailService{flagged: 2, purged: 1}}
	if err := executeEmailPrefilterStep(context, jobsapi.StepModel{}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestExecuteEmailPrefilterStepSkipsWhenNoWork(t *testing.T) {
	context := &WorkerContext{EmailService: &fakeEmailService{flagged: 0, purged: 0}}
	if err := executeEmailPrefilterStep(context, jobsapi.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped, got %v", err)
	}
}

func TestMaybeEnqueueEmailSyncCreatesTwoStepJob(t *testing.T) {
	repo := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repo, nil)
	context := &WorkerContext{
		EmailService:    &fakeEmailService{},
		JobOrchestrator: orchestrator,
	}

	maybeEnqueueEmailSync(context)

	jobModel, err := repo.GetJobByID(1)
	if err != nil {
		t.Fatalf("expected email_sync job: %v", err)
	}
	if jobModel.Type != string(jobdomain.JobTypeEmailSync) {
		t.Fatalf("unexpected job type: %q", jobModel.Type)
	}

	steps, err := repo.GetStepsByJobID(1)
	if err != nil {
		t.Fatalf("get steps: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps (fetch -> prefilter -> analyze), got %d", len(steps))
	}
}

func TestExecuteEmailAnalyzeStepSkipsWhenFeatureOff(t *testing.T) {
	if err := executeEmailAnalyzeStep(&WorkerContext{}, jobsapi.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped when no email service, got %v", err)
	}
}

func TestExecuteEmailAnalyzeStepSkipsWhenAIUnavailable(t *testing.T) {
	context := &WorkerContext{EmailService: &fakeEmailService{
		analyzeStats: emailapi.AnalyzeStats{AIUnavailable: true},
	}}
	if err := executeEmailAnalyzeStep(context, jobsapi.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("AI unavailable must skip (not fail) the step, got %v", err)
	}
}

func TestExecuteEmailAnalyzeStepSkipsWhenNothingAnalyzed(t *testing.T) {
	context := &WorkerContext{EmailService: &fakeEmailService{analyzeStats: emailapi.AnalyzeStats{}}}
	if err := executeEmailAnalyzeStep(context, jobsapi.StepModel{}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected ErrStepSkipped when nothing analyzed, got %v", err)
	}
}

func TestExecuteEmailAnalyzeStepPropagatesError(t *testing.T) {
	context := &WorkerContext{EmailService: &fakeEmailService{analyzeErr: errors.New("boom")}}
	if err := executeEmailAnalyzeStep(context, jobsapi.StepModel{}); err == nil {
		t.Fatal("expected analysis error to propagate")
	}
}

func TestExecuteEmailAnalyzeStepEmitsNotifications(t *testing.T) {
	notifier := &fakeWorkerNotifSvc{}
	context := &WorkerContext{
		NotificationService: notifier,
		EmailService: &fakeEmailService{analyzeStats: emailapi.AnalyzeStats{
			Analyzed:   3,
			Malicious:  []emailapi.EmailDetection{{AccountID: 1, Subject: "bad"}},
			Suspicious: []emailapi.EmailDetection{{AccountID: 1, Subject: "hmm"}},
			Important:  []emailapi.EmailDetection{{AccountID: 2, Subject: "bank"}},
		}},
	}
	if err := executeEmailAnalyzeStep(context, jobsapi.StepModel{}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(notifier.dtos) != 3 {
		t.Fatalf("expected 3 notifications (malicious, suspicious, important), got %d", len(notifier.dtos))
	}
}
