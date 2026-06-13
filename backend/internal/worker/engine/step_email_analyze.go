package engine

import (
	"fmt"

	"nas-go/api/internal/api/v1/email"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/pkg/i18n"
)

// executeEmailAnalyzeStep runs the AI analysis over pending messages: classify
// (verdict + risk + evidence + importance), summarize the legitimate ones, store
// the verdict and drop the body. It is the worker-side gate of the e-mail threat
// model — the body is adversarial input, so the analysis treats the model output
// strictly as data (fail-closed) and the LLM has no tools.
//
// When the AI is off (no provider enabled, or the chosen one disabled) the step
// is SKIPPED, not failed: the messages stay pending and the next scheduled
// email_sync retries — a temporarily unreachable model never breaks the job.
func executeEmailAnalyzeStep(context *WorkerContext, step jobs.StepModel) error {
	_ = step

	// The e-mail feature is off (no EMAIL_TOKEN_KEY); nothing to analyze.
	if context == nil || context.EmailService == nil {
		return ErrStepSkipped
	}

	stats, err := context.EmailService.AnalyzePending()
	if err != nil {
		return err
	}

	emitEmailDetections(context, stats.Malicious, "warning",
		"EMAIL_MALICIOUS_DETECTED", "EMAIL_MALICIOUS_DETECTED_MESSAGE", "email_analysis_malicious")
	emitEmailDetections(context, stats.Suspicious, "warning",
		"EMAIL_SUSPICIOUS_DETECTED", "EMAIL_SUSPICIOUS_DETECTED_MESSAGE", "email_analysis_suspicious")
	emitEmailDetections(context, stats.Important, "info",
		"EMAIL_IMPORTANT_RECEIVED", "EMAIL_IMPORTANT_RECEIVED_MESSAGE", "email_analysis_important")

	// AI unavailable or no message analyzed: skip so the job stays green and the
	// pending messages are retried next cycle.
	if stats.AIUnavailable || stats.Analyzed == 0 {
		return ErrStepSkipped
	}
	return nil
}

// emitEmailDetections notifies one detection per message, grouped per account so
// repeated detections on the same account collapse into a single counted entry.
func emitEmailDetections(context *WorkerContext, detections []email.EmailDetection, notifType, titleKey, messageKey, groupPrefix string) {
	for _, detection := range detections {
		emitNotification(
			context,
			notifType,
			i18n.GetMessage(titleKey),
			i18n.Translate(messageKey, detection.Subject),
			fmt.Sprintf("%s_%d", groupPrefix, detection.AccountID),
		)
	}
}
