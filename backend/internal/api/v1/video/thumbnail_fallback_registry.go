package video

import (
	"sync"
	"time"

	files "nas-go/api/internal/api/v1/files"
)

const (
	thumbnailFallbackTTL        = 10 * time.Minute
	thumbnailFallbackMaxEntries = 4096
)

type generatedImageKind string

const (
	generatedImageKindThumbnail generatedImageKind = "thumbnail"
	generatedImageKindPreview   generatedImageKind = "preview"
)

type GeneratedImage struct {
	Data       []byte
	IsFallback bool
}

type thumbnailFallbackKey struct {
	fileID int
	width  int
	height int
	kind   generatedImageKind
}

type thumbnailFallbackEntry struct {
	expiresAt       time.Time
	fileUpdatedAt   time.Time
	fileSizeInBytes int64
}

type thumbnailFallbackRegistry struct {
	mutex      sync.Mutex
	entries    map[thumbnailFallbackKey]thumbnailFallbackEntry
	now        func() time.Time
	maxEntries int
}

func newThumbnailFallbackKey(fileDto files.FileDto, width, height int, kind generatedImageKind) thumbnailFallbackKey {
	return thumbnailFallbackKey{fileID: fileDto.ID, width: width, height: height, kind: kind}
}

func (registry *thumbnailFallbackRegistry) currentTime() time.Time {
	if registry.now == nil {
		return time.Now()
	}
	return registry.now()
}

func (registry *thumbnailFallbackRegistry) entryLimit() int {
	if registry.maxEntries <= 0 {
		return thumbnailFallbackMaxEntries
	}
	return registry.maxEntries
}

func (registry *thumbnailFallbackRegistry) isRecentlyFailed(key thumbnailFallbackKey, fileDto files.FileDto) bool {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	entry, isRegistered := registry.entries[key]
	if !isRegistered {
		return false
	}
	isStale := !registry.currentTime().Before(entry.expiresAt) ||
		!entry.fileUpdatedAt.Equal(fileDto.UpdatedAt) ||
		entry.fileSizeInBytes != fileDto.Size
	if isStale {
		delete(registry.entries, key)
		return false
	}
	return true
}

func (registry *thumbnailFallbackRegistry) recordFailure(key thumbnailFallbackKey, fileDto files.FileDto) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if registry.entries == nil {
		registry.entries = make(map[thumbnailFallbackKey]thumbnailFallbackEntry)
	}
	if _, isRegistered := registry.entries[key]; !isRegistered && len(registry.entries) >= registry.entryLimit() {
		registry.evictOne()
	}
	registry.entries[key] = thumbnailFallbackEntry{
		expiresAt:       registry.currentTime().Add(thumbnailFallbackTTL),
		fileUpdatedAt:   fileDto.UpdatedAt,
		fileSizeInBytes: fileDto.Size,
	}
}

func (registry *thumbnailFallbackRegistry) clear(key thumbnailFallbackKey) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	delete(registry.entries, key)
}

func (registry *thumbnailFallbackRegistry) evictOne() {
	currentTime := registry.currentTime()
	var oldestKey thumbnailFallbackKey
	var oldestExpiration time.Time
	isFirst := true
	for key, entry := range registry.entries {
		if !currentTime.Before(entry.expiresAt) {
			delete(registry.entries, key)
			return
		}
		if isFirst || entry.expiresAt.Before(oldestExpiration) {
			oldestKey, oldestExpiration, isFirst = key, entry.expiresAt, false
		}
	}
	delete(registry.entries, oldestKey)
}
