package watcher

import (
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/api/v1/watchfolders"
	"nas-go/api/pkg/i18n"
	"sync"
	"time"
)

type FolderWatcher struct {
	watchFolderService WatchFolderService
	librariesService   libraries.ServiceInterface
	filesService       files.ServiceInterface
	notification       notifications.ServiceInterface
	interval           time.Duration
	stopChan           chan struct{}
	doneChan           chan struct{}
	mu                 sync.Mutex
	running            bool
}

type WatchFolderService interface {
	GetEnabledWatchFolders() ([]watchfolders.WatchFolderModel, error)
	UpdateWatchFolderLastScan(id int, lastScanAt time.Time) error
}

func NewFolderWatcher(
	watchFolderService WatchFolderService,
	librariesService libraries.ServiceInterface,
	filesService files.ServiceInterface,
	notificationService notifications.ServiceInterface,
	interval time.Duration,
) *FolderWatcher {
	if interval <= 0 {
		interval = 60 * time.Second
	}

	return &FolderWatcher{
		watchFolderService: watchFolderService,
		librariesService:   librariesService,
		filesService:       filesService,
		notification:       notificationService,
		interval:           interval,
		stopChan:           make(chan struct{}),
		doneChan:           make(chan struct{}),
	}
}

func (fw *FolderWatcher) Start() {
	fw.mu.Lock()
	if fw.running {
		fw.mu.Unlock()
		return
	}
	fw.running = true
	fw.mu.Unlock()

	go fw.loop()
}

func (fw *FolderWatcher) Stop() {
	fw.mu.Lock()
	if !fw.running {
		fw.mu.Unlock()
		return
	}
	close(fw.stopChan)
	fw.running = false
	fw.mu.Unlock()

	<-fw.doneChan
}

func (fw *FolderWatcher) loop() {
	defer close(fw.doneChan)

	fw.scanOnce()
	ticker := time.NewTicker(fw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-fw.stopChan:
			return
		case <-ticker.C:
			fw.scanOnce()
		}
	}
}

func (fw *FolderWatcher) scanOnce() {
	if fw.watchFolderService == nil || fw.librariesService == nil || fw.filesService == nil {
		return
	}

	watchFolders, err := fw.watchFolderService.GetEnabledWatchFolders()
	if err != nil {
		log.Printf("[watcher] list enabled watch folders: %v", err)
		return
	}

	for _, watchFolder := range watchFolders {
		log.Println(i18n.Translate("WATCH_FOLDER_SCAN_STARTED", watchFolder.Path))

		scannedFiles, scanErr := ScanWatchFolder(watchFolder)
		if scanErr != nil {
			log.Printf("[watcher] scan watch folder %s: %v", watchFolder.Path, scanErr)
			continue
		}

		importedCount, processErr := fw.ProcessScannedFiles(watchFolder, scannedFiles)
		if processErr != nil {
			log.Printf("[watcher] process scanned files for %s: %v", watchFolder.Path, processErr)
		}

		if updateErr := fw.watchFolderService.UpdateWatchFolderLastScan(watchFolder.ID, time.Now()); updateErr != nil {
			log.Printf("[watcher] update last_scan_at for %s: %v", watchFolder.Path, updateErr)
		}

		log.Println(i18n.Translate("WATCH_FOLDER_SCAN_COMPLETED", len(scannedFiles), watchFolder.Path))

		if importedCount > 0 {
			fw.emitFolderImportedNotification(watchFolder, importedCount)
		}
	}
}
