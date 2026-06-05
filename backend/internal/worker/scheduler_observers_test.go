package worker

import (
	"errors"
	"testing"

	jobsapi "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/systemevent"
)

// recordingSystemEvents captures RecordEvent calls so tests can assert which
// audit/health events the observers emit.
type recordingSystemEvents struct {
	systemevent.ServiceInterface
	events []systemevent.EventType
}

func (r *recordingSystemEvents) RecordEvent(eventType systemevent.EventType, _ string) error {
	r.events = append(r.events, eventType)
	return nil
}

func TestJobSchedulerOnJobFinishedFiresWithFailedStatus(t *testing.T) {
	repository := newFakeJobsRepository()
	scheduler := NewJobScheduler(repository, map[StepType]StepExecutor{
		StepTypeScanFilesystem: func(jobsapi.StepModel) error {
			return errors.New("boom")
		},
	})

	var gotID int
	var gotType string
	var gotStatus JobStatus
	scheduler.SetOnJobFinished(func(jobID int, jobType string, status JobStatus) {
		gotID, gotType, gotStatus = jobID, jobType, status
	})

	job, _ := repository.CreateJob(nil, jobsapi.JobModel{
		Type:     string(JobTypeStartupScan),
		Priority: string(JobPriorityLow),
		Status:   string(JobStatusQueued),
	})
	_, _ = repository.CreateStep(nil, jobsapi.StepModel{
		JobID:       job.ID,
		Type:        string(StepTypeScanFilesystem),
		Status:      string(StepStatusQueued),
		MaxAttempts: 1,
	})

	if err := scheduler.processJob(job.ID); err != nil {
		t.Fatalf("processJob returned error: %v", err)
	}

	if gotID != job.ID {
		t.Fatalf("expected hook job id %d, got %d", job.ID, gotID)
	}
	if gotType != string(JobTypeStartupScan) {
		t.Fatalf("expected hook job type startup_scan, got %s", gotType)
	}
	if gotStatus != JobStatusFailed {
		t.Fatalf("expected hook status failed, got %s", gotStatus)
	}
}

func TestWireSchedulerObserversRecordsAndNotifies(t *testing.T) {
	repository := newFakeJobsRepository()
	scheduler := NewJobScheduler(repository, map[StepType]StepExecutor{})
	events := &recordingSystemEvents{}
	notifier := &fakeWorkerNotifSvc{}
	context := &WorkerContext{
		JobScheduler:        scheduler,
		SystemEvents:        events,
		NotificationService: notifier,
	}

	wireSchedulerObservers(context)

	if scheduler.onJobFinished == nil || scheduler.onStall == nil {
		t.Fatalf("expected both observers to be wired")
	}

	scheduler.onJobFinished(1, string(JobTypeStartupScan), JobStatusCompleted)
	scheduler.onJobFinished(2, string(JobTypeUploadProcess), JobStatusFailed)
	scheduler.onStall(4)

	if len(events.events) != 2 {
		t.Fatalf("expected 2 recorded events, got %d (%v)", len(events.events), events.events)
	}
	if events.events[0] != systemevent.EventTypeScanCompleted {
		t.Fatalf("expected first event SCAN_COMPLETED, got %s", events.events[0])
	}
	if events.events[1] != systemevent.EventTypeJobFailed {
		t.Fatalf("expected second event JOB_FAILED, got %s", events.events[1])
	}

	// One notification for the failed job, one for the stall.
	if len(notifier.dtos) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(notifier.dtos))
	}
	if notifier.dtos[0].GroupKey != "job_failed" {
		t.Fatalf("expected job_failed group, got %s", notifier.dtos[0].GroupKey)
	}
	if notifier.dtos[1].GroupKey != "scheduler_stall" {
		t.Fatalf("expected scheduler_stall group, got %s", notifier.dtos[1].GroupKey)
	}
}

func TestWireSchedulerObserversNilSafe(t *testing.T) {
	// No scheduler: must not panic.
	wireSchedulerObservers(nil)
	wireSchedulerObservers(&WorkerContext{})

	// Scheduler but no recorders: hook still wired, invoking it is a no-op.
	repository := newFakeJobsRepository()
	scheduler := NewJobScheduler(repository, map[StepType]StepExecutor{})
	context := &WorkerContext{JobScheduler: scheduler}
	wireSchedulerObservers(context)
	scheduler.onJobFinished(1, string(JobTypeStartupScan), JobStatusFailed)
	scheduler.onStall(2)
}
