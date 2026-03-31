package worker

import (
	"encoding/json"
	"fmt"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/api/v1/takeout"
	"nas-go/api/pkg/i18n"
	"os"
	"path/filepath"
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

	extractResult, err := takeout.ExtractTakeout(payload.ZipPath, context.LibrariesService)
	if err != nil {
		if context.NotificationService != nil {
			_, _ = context.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
				Type:     string(notifications.NotificationTypeError),
				Title:    i18n.GetMessage("NOTIFICATION_TAKEOUT_IMPORT_FAILED_TITLE"),
				Message:  i18n.Translate("NOTIFICATION_TAKEOUT_IMPORT_FAILED_MESSAGE", err.Error()),
				GroupKey: "",
			})
		}
		return err
	}

	if context.FilesService != nil {
		paths := make([]string, 0, len(extractResult.Files))
		for _, extracted := range extractResult.Files {
			paths = append(paths, extracted.DestinationPath)
		}
		if len(paths) > 0 {
			_, _ = context.FilesService.CreateUploadProcessJob(paths)
		}
	}

	_ = os.Remove(payload.ZipPath)
	_ = os.RemoveAll(filepath.Dir(payload.ZipPath))

	if context.NotificationService != nil {
		_, _ = context.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
			Type:     string(notifications.NotificationTypeSuccess),
			Title:    i18n.GetMessage("NOTIFICATION_TAKEOUT_IMPORT_COMPLETED_TITLE"),
			Message:  i18n.Translate("NOTIFICATION_TAKEOUT_IMPORT_COMPLETED_MESSAGE", len(extractResult.Files)),
			GroupKey: "takeout_import",
		})
	}

	return nil
}
