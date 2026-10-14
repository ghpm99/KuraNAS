package video

import (
	"sync"
	"time"
)

const (
	defaultPlaybackAccessWindow     = 5 * time.Minute
	defaultPlaybackAccessMaxEntries = 1024
)

type playbackAccessKey struct {
	clientIP string
	fileID   int
}

type playbackAccessTracker struct {
	mutex          sync.Mutex
	now            func() time.Time
	window         time.Duration
	maxEntries     int
	claimedAtByKey map[playbackAccessKey]time.Time
}

func newPlaybackAccessTracker(now func() time.Time, window time.Duration, maxEntries int) *playbackAccessTracker {
	return &playbackAccessTracker{
		now:            now,
		window:         window,
		maxEntries:     maxEntries,
		claimedAtByKey: make(map[playbackAccessKey]time.Time),
	}
}

func newDefaultPlaybackAccessTracker() *playbackAccessTracker {
	return newPlaybackAccessTracker(time.Now, defaultPlaybackAccessWindow, defaultPlaybackAccessMaxEntries)
}

func (tracker *playbackAccessTracker) claim(clientIP string, fileID int) bool {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	currentTime := tracker.now()
	key := playbackAccessKey{clientIP: clientIP, fileID: fileID}
	if claimedAt, isTracked := tracker.claimedAtByKey[key]; isTracked && currentTime.Sub(claimedAt) < tracker.window {
		return false
	}

	delete(tracker.claimedAtByKey, key)
	tracker.makeRoom(currentTime)
	tracker.claimedAtByKey[key] = currentTime
	return true
}

func (tracker *playbackAccessTracker) makeRoom(currentTime time.Time) {
	if len(tracker.claimedAtByKey) < tracker.maxEntries {
		return
	}
	tracker.removeExpired(currentTime)
	if len(tracker.claimedAtByKey) < tracker.maxEntries {
		return
	}
	tracker.removeOldest()
}

func (tracker *playbackAccessTracker) removeExpired(currentTime time.Time) {
	for key, claimedAt := range tracker.claimedAtByKey {
		if currentTime.Sub(claimedAt) >= tracker.window {
			delete(tracker.claimedAtByKey, key)
		}
	}
}

func (tracker *playbackAccessTracker) removeOldest() {
	var oldestKey playbackAccessKey
	var oldestClaimedAt time.Time
	isFirst := true
	for key, claimedAt := range tracker.claimedAtByKey {
		if isFirst || claimedAt.Before(oldestClaimedAt) {
			oldestKey = key
			oldestClaimedAt = claimedAt
			isFirst = false
		}
	}
	if !isFirst {
		delete(tracker.claimedAtByKey, oldestKey)
	}
}

func (tracker *playbackAccessTracker) size() int {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	return len(tracker.claimedAtByKey)
}

type SharedPlaybackAccess struct {
	tracker *playbackAccessTracker
}

func NewSharedPlaybackAccess() *SharedPlaybackAccess {
	return &SharedPlaybackAccess{tracker: newDefaultPlaybackAccessTracker()}
}
