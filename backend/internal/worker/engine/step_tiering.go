package engine

import (
	"fmt"
	"log"
	"time"

	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/worker/job"
	tieringengine "nas-go/api/internal/worker/tiering"
	"nas-go/api/pkg/i18n"
)

// executeTierMigrationStep runs one hot↔cold migration pass with the plan
// resolved from the persisted settings. A disabled or unconfigured feature
// skips the step; a pass that moves nothing also skips, so the nightly job does
// not pile "0 files migrated" notifications on the operator.
func executeTierMigrationStep(context *WorkerContext, step jobs.StepModel) error {
	_ = step

	if context == nil || context.TieringService == nil {
		return fmt.Errorf("tiering service is required for tier_migration step")
	}

	enabled, coldDir, promotions, demotions, err := context.TieringService.MigrationPlan(time.Now())
	if err != nil {
		emitTieringFailure(context)
		return err
	}
	if !enabled {
		return ErrStepSkipped
	}
	if len(promotions) == 0 && len(demotions) == 0 {
		return ErrStepSkipped
	}

	stats := tieringengine.Run(coldDir, promotions, demotions, context.TieringService.SetPhysicalPath)

	emitNotification(
		context,
		"info",
		i18n.GetMessage("NOTIFICATION_TIERING_COMPLETED_TITLE"),
		i18n.Translate("NOTIFICATION_TIERING_COMPLETED_MESSAGE", stats.Demoted, stats.Promoted, stats.Failures),
		"tier_migration",
	)
	return nil
}

func emitTieringFailure(context *WorkerContext) {
	emitNotification(
		context,
		"error",
		i18n.GetMessage("NOTIFICATION_TIERING_FAILED_TITLE"),
		i18n.GetMessage("NOTIFICATION_TIERING_FAILED_MESSAGE"),
		"tier_migration",
	)
}

// startTieringScheduler polls the migration schedule once a minute and enqueues
// a tier_migration job when the configured interval has elapsed. The whole
// policy lives in NextRunDue, so the loop stays a dumb ticker.
func startTieringScheduler(context *WorkerContext) {
	if context == nil || context.TieringService == nil || context.JobOrchestrator == nil {
		return
	}
	for {
		time.Sleep(time.Minute)
		maybeEnqueueTierMigration(context, time.Now())
	}
}

func maybeEnqueueTierMigration(context *WorkerContext, now time.Time) {
	due, err := context.TieringService.NextRunDue(now)
	if err != nil {
		log.Printf("[tiering] could not evaluate schedule: %v\n", err)
		return
	}
	if !due {
		return
	}

	plan := PlannedJob{
		Type:     job.JobTypeTierMigration,
		Priority: job.JobPriorityLow,
		Steps: []PlannedStep{
			{
				Key:         "tier_migration",
				Type:        job.StepTypeTierMigration,
				MaxAttempts: 1,
			},
		},
	}

	jobID, createErr := context.JobOrchestrator.CreateJob(plan)
	if createErr != nil {
		log.Printf("[tiering] failed to enqueue tier_migration job: %v\n", createErr)
		return
	}
	if jobID > 0 {
		log.Printf("tier_migration job enqueued id=%d\n", jobID)
	}
}
