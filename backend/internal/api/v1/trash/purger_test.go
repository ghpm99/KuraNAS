package trash

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPurgerRemovesExpiredItemsOnSchedule(t *testing.T) {
	s, repo, _, root := newTrashServiceForTest(t)

	expired := filepath.Join(root, "velho.txt")
	writeFile(t, expired, "velho")
	if err := s.MoveToTrash(expired, 1); err != nil {
		t.Fatalf("MoveToTrash: %v", err)
	}
	aged := repo.items[1]
	aged.DeletedAt = time.Now().AddDate(0, 0, -(DefaultRetentionDays + 1))
	repo.items[1] = aged
	trashPath := aged.TrashPath

	purger := NewPurger(s, 50*time.Millisecond)
	purger.Start()
	t.Cleanup(purger.Stop)

	deadline := time.Now().Add(2 * time.Second)
	for {
		if items, _ := repo.GetAllItems(); len(items) == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("purger did not remove the expired item in time")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if _, err := os.Stat(trashPath); !os.IsNotExist(err) {
		t.Fatalf("expired bytes must be gone from disk, stat err=%v", err)
	}
}

func TestPurgerStopTerminatesLoop(t *testing.T) {
	s, _, _, _ := newTrashServiceForTest(t)

	purger := NewPurger(s, 0) // 0 falls back to the default interval
	if purger.interval != DefaultPurgeInterval {
		t.Fatalf("expected default interval, got %v", purger.interval)
	}
	purger.Start()

	done := make(chan struct{})
	go func() {
		purger.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Stop did not return")
	}
}
