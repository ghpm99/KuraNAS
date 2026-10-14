package video

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newFakeClock() *fakeClock {
	return &fakeClock{current: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
}

func TestPlaybackAccessTrackerClaimsOncePerWindow(t *testing.T) {
	clock := newFakeClock()
	tracker := newPlaybackAccessTracker(clock.now, 5*time.Minute, 10)

	if !tracker.claim("10.0.0.1", 1) {
		t.Fatal("first claim must succeed")
	}
	clock.advance(4 * time.Minute)
	if tracker.claim("10.0.0.1", 1) {
		t.Fatal("claim inside window must be rejected")
	}
	clock.advance(time.Minute)
	if !tracker.claim("10.0.0.1", 1) {
		t.Fatal("claim after window must succeed")
	}
}

func TestPlaybackAccessTrackerSeparatesClientsAndFiles(t *testing.T) {
	tracker := newPlaybackAccessTracker(newFakeClock().now, 5*time.Minute, 10)

	if !tracker.claim("10.0.0.1", 1) || !tracker.claim("10.0.0.2", 1) || !tracker.claim("10.0.0.1", 2) {
		t.Fatal("distinct ip or file must claim independently")
	}
}

func TestPlaybackAccessTrackerDropsExpiredEntriesWhenFull(t *testing.T) {
	clock := newFakeClock()
	tracker := newPlaybackAccessTracker(clock.now, time.Minute, 3)
	tracker.claim("ip", 1)
	tracker.claim("ip", 2)
	clock.advance(2 * time.Minute)
	tracker.claim("ip", 3)
	tracker.claim("ip", 4)

	if tracker.size() != 2 {
		t.Fatalf("expected expired entries dropped, size %d", tracker.size())
	}
}

func TestPlaybackAccessTrackerEvictsOldestWhenFullOfLiveEntries(t *testing.T) {
	clock := newFakeClock()
	tracker := newPlaybackAccessTracker(clock.now, time.Hour, 3)
	for fileID := 1; fileID <= 3; fileID++ {
		tracker.claim("ip", fileID)
		clock.advance(time.Second)
	}
	tracker.claim("ip", 4)

	if tracker.size() != 3 {
		t.Fatalf("size must stay at limit, got %d", tracker.size())
	}
	if !tracker.claim("ip", 1) {
		t.Fatal("oldest entry must have been evicted")
	}
	if tracker.claim("ip", 4) {
		t.Fatal("newest entry must still be tracked")
	}
}

func TestPlaybackAccessTrackerConcurrentClaimsGrantExactlyOne(t *testing.T) {
	tracker := newPlaybackAccessTracker(newFakeClock().now, time.Hour, 100)
	var grantedClaims atomic.Int32
	var group sync.WaitGroup
	for worker := 0; worker < 50; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if tracker.claim("ip", 7) {
				grantedClaims.Add(1)
			}
			tracker.claim("ip", worker+100)
		}()
	}
	group.Wait()

	if grantedClaims.Load() != 1 {
		t.Fatalf("expected exactly one granted claim, got %d", grantedClaims.Load())
	}
	if tracker.size() > 100 {
		t.Fatal(fmt.Sprintf("size exceeded limit: %d", tracker.size()))
	}
}
