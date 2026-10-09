package search

import (
	"strings"
	"sync"
	"time"
)

const (
	expansionCacheCapacity = 200
	expansionCacheTTL      = 10 * time.Minute
)

type cachedExpansion struct {
	expansion aiSearchExpansion
	expiresAt time.Time
}

type expansionCache struct {
	mutex    sync.Mutex
	entries  map[string]cachedExpansion
	capacity int
	ttl      time.Duration
	now      func() time.Time
}

func newExpansionCache(capacity int, ttl time.Duration) *expansionCache {
	return &expansionCache{
		entries:  make(map[string]cachedExpansion, capacity),
		capacity: capacity,
		ttl:      ttl,
		now:      time.Now,
	}
}

func normalizeExpansionKey(query string) string {
	return strings.ToLower(strings.Join(strings.Fields(query), " "))
}

func (cache *expansionCache) get(query string) (aiSearchExpansion, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	key := normalizeExpansionKey(query)
	entry, found := cache.entries[key]
	if !found {
		return aiSearchExpansion{}, false
	}
	if !cache.now().Before(entry.expiresAt) {
		delete(cache.entries, key)
		return aiSearchExpansion{}, false
	}
	return entry.expansion, true
}

func (cache *expansionCache) put(query string, expansion aiSearchExpansion) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	key := normalizeExpansionKey(query)
	if _, found := cache.entries[key]; !found && len(cache.entries) >= cache.capacity {
		cache.evictOne()
	}
	cache.entries[key] = cachedExpansion{expansion: expansion, expiresAt: cache.now().Add(cache.ttl)}
}

func (cache *expansionCache) evictOne() {
	now := cache.now()
	oldestKey := ""
	var oldestExpiry time.Time
	for key, entry := range cache.entries {
		if !now.Before(entry.expiresAt) {
			delete(cache.entries, key)
			return
		}
		if oldestKey == "" || entry.expiresAt.Before(oldestExpiry) {
			oldestKey = key
			oldestExpiry = entry.expiresAt
		}
	}
	delete(cache.entries, oldestKey)
}
