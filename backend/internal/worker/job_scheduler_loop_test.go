package worker

import (
	"testing"
	"time"

	jobsapi "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
)

func TestJobSchedulerStartLoopAndStop(t *testing.T) {
	previousConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = previousConfig
	})
	config.AppConfig.WorkerSchedulerPollMS = 1

	repository := newFakeJobsRepository()
	executed := make(chan int, 1)
	scheduler := NewJobScheduler(repository, map[StepType]StepExecutor{
		StepTypeScanFilesystem: func(step jobsapi.StepModel) error {
			executed <- step.JobID
			return nil
		},
	})

	job, err := repository.CreateJob(nil, jobsapi.JobModel{
		Type:     string(JobTypeStartupScan),
		Priority: string(JobPriorityLow),
		Status:   string(JobStatusQueued),
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}
	if _, err := repository.CreateStep(nil, jobsapi.StepModel{
		JobID:       job.ID,
		Type:        string(StepTypeScanFilesystem),
		Status:      string(StepStatusQueued),
		MaxAttempts: 1,
	}); err != nil {
		t.Fatalf("CreateStep returned error: %v", err)
	}

	if scheduler.Enqueue(0) {
		t.Fatalf("expected invalid job id to be rejected")
	}
	if !scheduler.Enqueue(job.ID) {
		t.Fatalf("expected job id to be enqueued")
	}
	if !scheduler.Enqueue(job.ID) {
		t.Fatalf("expected duplicate enqueue to be ignored but reported as success")
	}

	scheduler.Start()
	scheduler.Start()

	select {
	case got := <-executed:
		if got != job.ID {
			t.Fatalf("unexpected executed job id %d", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected scheduler loop to execute job")
	}

	scheduler.Stop()
}

func TestJobSchedulerScheduleQueuedJobsQueuesRepositoryJobs(t *testing.T) {
	repository := newFakeJobsRepository()
	scheduler := NewJobScheduler(repository, map[StepType]StepExecutor{})

	high, _ := repository.CreateJob(nil, jobsapi.JobModel{
		Type:     string(JobTypeStartupScan),
		Priority: string(JobPriorityHigh),
		Status:   string(JobStatusQueued),
	})
	_, _ = repository.CreateStep(nil, jobsapi.StepModel{
		JobID:       high.ID,
		Type:        string(StepTypeScanFilesystem),
		Status:      string(StepStatusQueued),
		MaxAttempts: 1,
	})

	scheduler.scheduleQueuedJobs()

	if len(scheduler.queue) != 1 {
		t.Fatalf("expected one queued job in channel, got %d", len(scheduler.queue))
	}
	if _, exists := scheduler.queued[high.ID]; !exists {
		t.Fatalf("expected queued job id %d", high.ID)
	}
}

func TestStartWorkersSchedulerWithOrchestrator(t *testing.T) {
	previousConfig := config.AppConfig
	t.Cleanup(func() {
		config.AppConfig = previousConfig
	})

	root := t.TempDir()
	config.AppConfig.EntryPoint = root

	repository := newFakeJobsRepository()
	context := &WorkerContext{JobOrchestrator: NewJobOrchestrator(repository, nil)}

	startWorkersScheduler(context)

	if len(repository.jobs) != 1 {
		t.Fatalf("expected startup scan job to be scheduled, got %d", len(repository.jobs))
	}
}
