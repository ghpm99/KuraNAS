package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/api/v1/takeout"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"os"
)

func executeTakeoutExtractStep(context *WorkerContext, step jobs.StepModel) error {
	if context == nil || context.LibrariesService == nil {
		return fmt.Errorf("libraries service is required for takeout extract step")
	}

	var payload TakeoutStepPayload
	if err := json.Unmarshal(step.Payload, &payload); err != nil {
		return fmt.Errorf("decode takeout payload: %w", err)
	}
	if payload.ZipPath == "" {
		return fmt.Errorf("takeout payload zip_path is required")
	}

	defer removeTakeoutUploadSession(payload.ZipPath)

	recordTakeoutImportRunning(context, step)

	progressReporter := newTakeoutProgressReporter(context, step)
	extraction, err := takeout.ExtractTakeoutWithProgress(payload.ZipPath, context.LibrariesService, progressReporter.report)
	if err != nil {
		persistExtractedTakeoutFiles(context, extraction)
		applog.Error("takeout import failed", "step_id", step.ID, "job_id", step.JobID, "zip_path", payload.ZipPath, "error", err.Error())
		failureMessage := i18n.GetMessage(takeoutFailureMessageKey(err))
		recordTakeoutImportFailed(context, step, failureMessage, extraction)
		if context.NotificationService != nil {
			_, _ = context.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
				Type:     string(notifications.NotificationTypeError),
				Title:    i18n.GetMessage("NOTIFICATION_TAKEOUT_IMPORT_FAILED_TITLE"),
				Message:  failureMessage,
				GroupKey: "",
				Metadata: takeoutImportEventMetadata("takeout_import_failed", payload.ImportID, step.JobID),
			})
		}
		return err
	}

	persistExtractedTakeoutFiles(context, extraction)

	recordTakeoutImportCompleted(context, step, extraction)

	if context.NotificationService != nil {
		_, _ = context.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
			Type:  string(notifications.NotificationTypeSuccess),
			Title: i18n.GetMessage("NOTIFICATION_TAKEOUT_IMPORT_COMPLETED_TITLE"),
			Message: i18n.Translate(
				"NOTIFICATION_TAKEOUT_IMPORT_COMPLETED_MESSAGE",
				extraction.Imported,
				extraction.AlreadyExisting,
				extraction.Ignored,
				extraction.FailedEntries,
			),
			GroupKey: "takeout_import",
			Metadata: takeoutImportEventMetadata("takeout_import_completed", payload.ImportID, step.JobID),
		})
	}

	return nil
}

func takeoutImportEventMetadata(event string, importID int, jobID int) map[string]any {
	return map[string]any{
		"event":     event,
		"import_id": importID,
		"job_id":    jobID,
	}
}

func removeTakeoutUploadSession(zipPath string) {
	sessionDir, isInsideSessionsArea := takeout.UploadSessionDirOfZip(zipPath)
	if !isInsideSessionsArea {
		applog.Warn("takeout cleanup: zip is outside the upload sessions area, nothing removed", "zip_path", zipPath)
		return
	}
	if err := os.RemoveAll(sessionDir); err != nil {
		applog.Error("takeout cleanup: remove upload session failed", "session_dir", sessionDir, "error", err.Error())
	}
}

func persistExtractedTakeoutFiles(context *WorkerContext, extraction takeout.ExtractResult) {
	recordTakeoutSidecars(context, extraction)
	enqueueTakeoutIndexing(context, extraction)
}

func enqueueTakeoutIndexing(context *WorkerContext, extraction takeout.ExtractResult) {
	if context.FilesService == nil || len(extraction.Files) == 0 {
		return
	}
	extractedPaths := make([]string, 0, len(extraction.Files))
	for _, extracted := range extraction.Files {
		extractedPaths = append(extractedPaths, extracted.DestinationPath)
	}
	_, _ = context.FilesService.CreateUploadProcessJob(extractedPaths)
}

func recordTakeoutSidecars(context *WorkerContext, extraction takeout.ExtractResult) {
	if context.TakeoutSidecarRecorder == nil {
		return
	}
	sidecars := takeout.BuildPhotoSidecars(extraction.Files)
	if len(sidecars) == 0 {
		return
	}
	if err := context.TakeoutSidecarRecorder.SaveSidecars(sidecars); err != nil {
		applog.Error("takeout sidecar: save failed", "error", err.Error())
	}
}

func takeoutFailureMessageKey(extractionErr error) string {
	if errors.Is(extractionErr, takeout.ErrInvalidZipFile) {
		return "NOTIFICATION_TAKEOUT_IMPORT_FAILED_INVALID_ZIP_MESSAGE"
	}
	return "NOTIFICATION_TAKEOUT_IMPORT_FAILED_MESSAGE"
}

func recordTakeoutImportRunning(context *WorkerContext, step jobs.StepModel) {
	if context.TakeoutImports == nil {
		return
	}
	if err := context.TakeoutImports.MarkImportRunning(step.JobID); err != nil {
		applog.Error("takeout import record: mark running failed", "job_id", step.JobID, "error", err.Error())
	}
}

func recordTakeoutImportCompleted(context *WorkerContext, step jobs.StepModel, extraction takeout.ExtractResult) {
	if context.TakeoutImports == nil {
		return
	}
	summary := takeout.ImportSummary{
		TotalEntries:    extraction.TotalEntries,
		Imported:        extraction.Imported,
		AlreadyExisting: extraction.AlreadyExisting,
		Ignored:         extraction.Ignored,
		FailedEntries:   extraction.FailedEntries,
	}
	if err := context.TakeoutImports.CloseImportCompleted(step.JobID, summary); err != nil {
		applog.Error("takeout import record: close completed failed", "job_id", step.JobID, "error", err.Error())
	}
}

func recordTakeoutImportFailed(context *WorkerContext, step jobs.StepModel, failureMessage string, extraction takeout.ExtractResult) {
	if context.TakeoutImports == nil {
		return
	}
	summary := takeout.ImportSummary{
		TotalEntries:    extraction.TotalEntries,
		Imported:        extraction.Imported,
		AlreadyExisting: extraction.AlreadyExisting,
		Ignored:         extraction.Ignored,
		FailedEntries:   extraction.FailedEntries,
	}
	if err := context.TakeoutImports.CloseImportFailed(step.JobID, failureMessage, summary); err != nil {
		applog.Error("takeout import record: close failed failed", "job_id", step.JobID, "error", err.Error())
	}
}
