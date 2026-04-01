package watcher

import (
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/api/v1/watchfolders"
	"nas-go/api/pkg/i18n"
	"path/filepath"
	"strings"
)

func (fw *FolderWatcher) ProcessScannedFiles(watchFolder watchfolders.WatchFolderModel, files []ScannedFile) (int, error) {
	if len(files) == 0 {
		return 0, nil
	}

	imported := 0
	errorsFound := make([]string, 0)

	for _, scannedFile := range files {
		libraryDto, err := fw.librariesService.GetLibraryByCategory(scannedFile.Category)
		if err != nil {
			errorsFound = append(errorsFound, fmt.Sprintf("resolve library for %s: %v", scannedFile.SourcePath, err))
			continue
		}

		targetPath, err := MoveToLibrary(scannedFile, libraryDto.Path)
		if err != nil {
			errorsFound = append(errorsFound, fmt.Sprintf("move file %s: %v", scannedFile.SourcePath, err))
			continue
		}

		if _, err := fw.filesService.CreateUploadProcessJob([]string{targetPath}); err != nil {
			errorsFound = append(errorsFound, fmt.Sprintf("enqueue processing %s: %v", targetPath, err))
			continue
		}

		imported++
		log.Println(i18n.Translate("WATCH_FOLDER_FILE_IMPORTED", scannedFile.SourcePath, targetPath))
	}

	if len(errorsFound) > 0 {
		return imported, fmt.Errorf("%s", strings.Join(errorsFound, "; "))
	}
	return imported, nil
}

func (fw *FolderWatcher) emitFolderImportedNotification(watchFolder watchfolders.WatchFolderModel, importedCount int) {
	if fw.notification == nil || importedCount <= 0 {
		return
	}

	label := watchFolder.Label
	if strings.TrimSpace(label) == "" {
		label = filepath.Base(watchFolder.Path)
	}
	if strings.TrimSpace(label) == "" {
		label = watchFolder.Path
	}

	_, err := fw.notification.GroupOrCreate(notifications.CreateNotificationDto{
		Type:     string(notifications.NotificationTypeSuccess),
		Title:    i18n.GetMessage("NOTIFICATION_WATCH_IMPORT_TITLE"),
		Message:  i18n.Translate("NOTIFICATION_WATCH_IMPORT_MESSAGE", importedCount, label),
		GroupKey: fmt.Sprintf("watch_import_%d", watchFolder.ID),
		Metadata: map[string]any{
			"event":       "watch_folder_import",
			"watch_id":    watchFolder.ID,
			"watch_path":  watchFolder.Path,
			"watch_label": watchFolder.Label,
			"count":       importedCount,
		},
	})
	if err != nil {
		log.Printf("[watcher] emit watch import notification: %v", err)
	}
}
