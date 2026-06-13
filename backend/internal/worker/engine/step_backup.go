package engine

import (
	"fmt"
	"log"
	"time"

	"nas-go/api/internal/api/v1/jobs"
	backupengine "nas-go/api/internal/worker/backup"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/i18n"
)

// executeBackupRunStep runs one incremental backup pass with the options
// resolved from the persisted settings. A disabled or unconfigured backup
// skips the step; a pass that finds nothing to do also skips, so the daily
// job does not pile "0 files copied" notifications on the operator.
func executeBackupRunStep(context *WorkerContext, step jobs.StepModel) error {
	_ = step

	if context == nil || context.BackupService == nil {
		return fmt.Errorf("backup service is required for backup_run step")
	}

	enabled, opts, err := context.BackupService.RunOptions()
	if err != nil {
		emitBackupFailure(context)
		return err
	}
	if !enabled {
		return ErrStepSkipped
	}

	stats, runErr := backupengine.Run(opts)
	if runErr != nil {
		emitBackupFailure(context)
		return runErr
	}

	if stats.Copied == 0 && stats.Versioned == 0 && stats.Purged == 0 && stats.Failures == 0 {
		return ErrStepSkipped
	}

	emitNotification(
		context,
		"info",
		i18n.GetMessage("NOTIFICATION_BACKUP_COMPLETED_TITLE"),
		i18n.Translate("NOTIFICATION_BACKUP_COMPLETED_MESSAGE", stats.Copied, stats.Versioned, stats.Failures),
		"backup_run",
	)
	return nil
}

func emitBackupFailure(context *WorkerContext) {
	emitNotification(
		context,
		"error",
		i18n.GetMessage("NOTIFICATION_BACKUP_FAILED_TITLE"),
		i18n.GetMessage("NOTIFICATION_BACKUP_FAILED_MESSAGE"),
		"backup_run",
	)
}

// startBackupScheduler polls the backup schedule once a minute and enqueues a
// backup_run job when the configured interval has elapsed. The whole policy
// (feature on, destination set, no run in flight, interval) lives in
// NextRunDue, so the loop stays a dumb ticker.
func startBackupScheduler(context *WorkerContext) {
	if context == nil || context.BackupService == nil || context.JobOrchestrator == nil {
		return
	}
	for {
		time.Sleep(time.Minute)
		maybeEnqueueBackupRun(context, time.Now())
	}
}

func maybeEnqueueBackupRun(context *WorkerContext, now time.Time) {
	due, err := context.BackupService.NextRunDue(now)
	if err != nil {
		log.Printf("[backup] could not evaluate schedule: %v\n", err)
		return
	}
	if !due {
		return
	}

	plan := PlannedJob{
		Type:     job.JobTypeBackupRun,
		Priority: job.JobPriorityLow,
		Steps: []PlannedStep{
			{
				Key:         "backup_run",
				Type:        job.StepTypeBackupRun,
				MaxAttempts: 1,
			},
		},
	}

	jobID, createErr := context.JobOrchestrator.CreateJob(plan)
	if createErr != nil {
		log.Printf("[backup] failed to enqueue backup_run job: %v\n", createErr)
		return
	}
	if jobID > 0 {
		log.Printf("backup_run job enqueued id=%d\n", jobID)
	}
}
