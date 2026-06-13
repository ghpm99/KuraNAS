package engine

import (
	"log"
	"time"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker/job"
	"nas-go/api/pkg/i18n"
)

// executeEmailFetchStep pulls new messages for every sync-enabled account and
// stores them sanitized as 'pending'. A rejected token marks that account
// reauth_required (handled in the service) and surfaces a notification, but does
// NOT fail the step — the other accounts still sync.
func executeEmailFetchStep(context *WorkerContext, step jobs.StepModel) error {
	_ = step

	// The e-mail feature is off (no EMAIL_TOKEN_KEY); there is nothing to fetch.
	if context == nil || context.EmailService == nil {
		return ErrStepSkipped
	}

	stats, err := context.EmailService.SyncEnabledAccounts()
	if err != nil {
		return err
	}

	for _, address := range stats.ReauthRequired {
		emitNotification(
			context,
			"error",
			i18n.GetMessage("NOTIFICATION_EMAIL_REAUTH_TITLE"),
			i18n.Translate("NOTIFICATION_EMAIL_REAUTH_MESSAGE", address),
			"email_reauth",
		)
	}

	if stats.Fetched == 0 {
		return ErrStepSkipped
	}

	emitNotification(
		context,
		"info",
		i18n.GetMessage("NOTIFICATION_EMAIL_SYNC_COMPLETED_TITLE"),
		i18n.Translate("NOTIFICATION_EMAIL_SYNC_COMPLETED_MESSAGE", stats.Fetched),
		"email_sync",
	)
	return nil
}

// executeEmailPrefilterStep runs the deterministic pre-filter over pending
// messages and then expurges anything past the retention window. Spam flagged
// here never reaches the LLM (task 16).
func executeEmailPrefilterStep(context *WorkerContext, step jobs.StepModel) error {
	_ = step

	if context == nil || context.EmailService == nil {
		return ErrStepSkipped
	}

	flagged, err := context.EmailService.PrefilterPending()
	if err != nil {
		return err
	}

	purged, purgeErr := context.EmailService.PurgeExpired()
	if purgeErr != nil {
		return purgeErr
	}

	if flagged == 0 && purged == 0 {
		return ErrStepSkipped
	}
	return nil
}

// startEmailSyncScheduler enqueues an email_sync job on a fixed interval
// (EMAIL_SYNC_INTERVAL_MINUTES, default 10). The job is a no-op when no account
// is enabled, so the ticker stays a dumb timer.
func startEmailSyncScheduler(context *WorkerContext) {
	if context == nil || context.EmailService == nil || context.JobOrchestrator == nil {
		return
	}

	interval := time.Duration(config.AppConfig.EmailSyncIntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = 10 * time.Minute
	}

	// A short initial delay lets the worker pool settle, then a first pass so
	// the kiosk has data without waiting a full interval.
	time.Sleep(30 * time.Second)
	maybeEnqueueEmailSync(context)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		maybeEnqueueEmailSync(context)
	}
}

func maybeEnqueueEmailSync(context *WorkerContext) {
	plan := PlannedJob{
		Type:     job.JobTypeEmailSync,
		Priority: job.JobPriorityLow,
		Steps: []PlannedStep{
			{
				Key:         "email_fetch",
				Type:        job.StepTypeEmailFetch,
				MaxAttempts: 1,
			},
			{
				Key:         "email_prefilter",
				Type:        job.StepTypeEmailPrefilter,
				DependsOn:   []string{"email_fetch"},
				MaxAttempts: 1,
			},
			{
				Key:         "email_analyze",
				Type:        job.StepTypeEmailAnalyze,
				DependsOn:   []string{"email_prefilter"},
				MaxAttempts: 1,
			},
		},
	}

	jobID, err := context.JobOrchestrator.CreateJob(plan)
	if err != nil {
		log.Printf("[email] failed to enqueue email_sync job: %v\n", err)
		return
	}
	if jobID > 0 {
		log.Printf("email_sync job enqueued id=%d\n", jobID)
	}
}
