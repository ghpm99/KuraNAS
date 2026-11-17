package files

import (
	"nas-go/api/pkg/applog"
	"os"
	"sync"
	"time"
)

const (
	uploadSessionIdleLimit    = 24 * time.Hour
	DefaultUploadReapInterval = time.Hour
)

func tryLockUploadSession(uploadID string) (func(), bool) {
	lock, _ := uploadSessionLocks.LoadOrStore(uploadID, &sync.Mutex{})
	sessionLock := lock.(*sync.Mutex)
	if !sessionLock.TryLock() {
		return nil, false
	}
	return sessionLock.Unlock, true
}

func isUploadSessionExpired(uploadID string, now time.Time) bool {
	session, err := loadUploadSession(uploadID)
	if err == nil {
		return now.Sub(session.LastActivityAt) > uploadSessionIdleLimit
	}
	directoryInfo, statErr := os.Stat(uploadSessionDir(uploadID))
	if statErr != nil {
		return false
	}
	return now.Sub(directoryInfo.ModTime()) > uploadSessionIdleLimit
}

func removeUploadSessionIfExpired(uploadID string, now time.Time) bool {
	unlock, isLocked := tryLockUploadSession(uploadID)
	if !isLocked {
		return false
	}
	defer unlock()
	if !isUploadSessionExpired(uploadID, now) {
		return false
	}
	if err := os.RemoveAll(uploadSessionDir(uploadID)); err != nil {
		applog.Error("remove expired upload session", "upload_id", uploadID, "error", err)
		return false
	}
	return true
}

// RemoveExpiredUploadSessions deletes every upload session whose last
// activity is older than 24 hours (or whose metadata is unreadable and whose
// directory is that old), skipping sessions currently locked by a request.
// It returns how many sessions were removed.
func RemoveExpiredUploadSessions(now time.Time) int {
	entries, err := os.ReadDir(uploadSessionsRootDir())
	if err != nil {
		return 0
	}
	removedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() || !isValidUploadID(entry.Name()) {
			continue
		}
		if removeUploadSessionIfExpired(entry.Name(), now) {
			removedCount++
		}
	}
	if removedCount > 0 {
		applog.Info("expired upload sessions removed", "count", removedCount)
	}
	return removedCount
}

// UploadSessionReaper periodically removes expired upload sessions.
type UploadSessionReaper struct {
	interval time.Duration
	now      func() time.Time
	stop     chan struct{}
	done     chan struct{}
}

// NewUploadSessionReaper builds a reaper; a nil clock defaults to time.Now
// and a non-positive interval to DefaultUploadReapInterval.
func NewUploadSessionReaper(interval time.Duration, now func() time.Time) *UploadSessionReaper {
	if interval <= 0 {
		interval = DefaultUploadReapInterval
	}
	if now == nil {
		now = time.Now
	}
	return &UploadSessionReaper{
		interval: interval,
		now:      now,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (reaper *UploadSessionReaper) Start() {
	go reaper.run()
}

func (reaper *UploadSessionReaper) Stop() {
	close(reaper.stop)
	<-reaper.done
}

func (reaper *UploadSessionReaper) run() {
	defer close(reaper.done)
	RemoveExpiredUploadSessions(reaper.now())
	ticker := time.NewTicker(reaper.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			RemoveExpiredUploadSessions(reaper.now())
		case <-reaper.stop:
			return
		}
	}
}
