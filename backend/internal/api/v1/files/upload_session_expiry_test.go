package files

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

var expiryTestNow = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

func createSessionIdleFor(t *testing.T, uploadID string, idleDuration time.Duration) {
	t.Helper()
	session := UploadSession{UploadID: uploadID, LastActivityAt: expiryTestNow.Add(-idleDuration)}
	if err := createUploadSession(session); err != nil {
		t.Fatalf("createUploadSession failed: %v", err)
	}
}

func uploadSessionDirExists(uploadID string) bool {
	_, err := os.Stat(uploadSessionDir(uploadID))
	return err == nil
}

func TestRemoveExpiredUploadSessionsDeletesSessionIdleFor25Hours(t *testing.T) {
	newUploadTestEnvironment(t, nil)
	uploadID := strings.Repeat("a", 32)
	createSessionIdleFor(t, uploadID, 25*time.Hour)

	removedCount := RemoveExpiredUploadSessions(expiryTestNow)

	if removedCount != 1 || uploadSessionDirExists(uploadID) {
		t.Fatalf("expected idle session removed, removed=%d exists=%v", removedCount, uploadSessionDirExists(uploadID))
	}
}

func TestRemoveExpiredUploadSessionsKeepsSessionIdleFor23Hours(t *testing.T) {
	newUploadTestEnvironment(t, nil)
	uploadID := strings.Repeat("a", 32)
	createSessionIdleFor(t, uploadID, 23*time.Hour)

	removedCount := RemoveExpiredUploadSessions(expiryTestNow)

	if removedCount != 0 || !uploadSessionDirExists(uploadID) {
		t.Fatalf("expected recent session kept, removed=%d", removedCount)
	}
}

func TestRemoveExpiredUploadSessionsSkipsLockedSession(t *testing.T) {
	newUploadTestEnvironment(t, nil)
	uploadID := strings.Repeat("a", 32)
	createSessionIdleFor(t, uploadID, 30*time.Hour)
	unlock := lockUploadSession(uploadID)

	removedWhileLocked := RemoveExpiredUploadSessions(expiryTestNow)
	unlock()
	removedAfterUnlock := RemoveExpiredUploadSessions(expiryTestNow)

	if removedWhileLocked != 0 || removedAfterUnlock != 1 {
		t.Fatalf("expected skip while locked then removal, got %d and %d", removedWhileLocked, removedAfterUnlock)
	}
}

func TestRemoveExpiredUploadSessionsDeletesOldCorruptedMeta(t *testing.T) {
	newUploadTestEnvironment(t, nil)
	uploadID := strings.Repeat("b", 32)
	createSessionIdleFor(t, uploadID, time.Hour)
	if err := os.WriteFile(uploadSessionMetaPath(uploadID), []byte("{"), 0644); err != nil {
		t.Fatalf("corrupt meta: %v", err)
	}
	oldTime := expiryTestNow.Add(-26 * time.Hour)
	if err := os.Chtimes(uploadSessionDir(uploadID), oldTime, oldTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	removedCount := RemoveExpiredUploadSessions(expiryTestNow)

	if removedCount != 1 || uploadSessionDirExists(uploadID) {
		t.Fatalf("expected corrupted old session removed, removed=%d", removedCount)
	}
}

func TestRemoveExpiredUploadSessionsKeepsRecentCorruptedMeta(t *testing.T) {
	newUploadTestEnvironment(t, nil)
	uploadID := strings.Repeat("b", 32)
	createSessionIdleFor(t, uploadID, time.Hour)
	if err := os.WriteFile(uploadSessionMetaPath(uploadID), []byte("{"), 0644); err != nil {
		t.Fatalf("corrupt meta: %v", err)
	}
	if err := os.Chtimes(uploadSessionDir(uploadID), expiryTestNow, expiryTestNow); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	removedCount := RemoveExpiredUploadSessions(expiryTestNow)

	if removedCount != 0 || !uploadSessionDirExists(uploadID) {
		t.Fatalf("expected recent corrupted session kept, removed=%d", removedCount)
	}
}

func TestUploadSessionReaperRemovesOnStartAndStopsCleanly(t *testing.T) {
	newUploadTestEnvironment(t, nil)
	uploadID := strings.Repeat("c", 32)
	createSessionIdleFor(t, uploadID, 48*time.Hour)
	goroutinesBefore := runtime.NumGoroutine()
	reaper := NewUploadSessionReaper(time.Millisecond, func() time.Time { return expiryTestNow })

	reaper.Start()
	deadline := time.Now().Add(2 * time.Second)
	for uploadSessionDirExists(uploadID) && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	reaper.Stop()

	if uploadSessionDirExists(uploadID) {
		t.Fatal("reaper did not remove the expired session")
	}
	if goroutinesAfter := runtime.NumGoroutine(); goroutinesAfter > goroutinesBefore {
		t.Fatalf("goroutine leak: before=%d after=%d", goroutinesBefore, goroutinesAfter)
	}
}

func TestNewUploadSessionReaperAppliesDefaults(t *testing.T) {
	reaper := NewUploadSessionReaper(0, nil)

	if reaper.interval != DefaultUploadReapInterval || reaper.now == nil {
		t.Fatalf("defaults not applied: %+v", reaper)
	}
}
