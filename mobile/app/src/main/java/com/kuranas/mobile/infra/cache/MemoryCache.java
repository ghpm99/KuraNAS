package com.kuranas.mobile.infra.cache;

import java.util.HashMap;
import java.util.Map;

public final class MemoryCache<T> {

    private final long ttlMs;
    private final Map<String, CacheEntry<T>> entries;

    public MemoryCache(long ttlMs) {
        this.ttlMs = ttlMs;
        this.entries = new HashMap<String, CacheEntry<T>>();
    }

    public synchronized T get(String key) {
        CacheEntry<T> entry = entries.get(key);
        if (entry == null) {
            return null;
        }
        if (System.currentTimeMillis() - entry.timestamp > ttlMs) {
            entries.remove(key);
            return null;
        }
        return entry.value;
    }

    public synchronized void put(String key, T value) {
        entries.put(key, new CacheEntry<T>(value, System.currentTimeMillis()));
    }

    public synchronized void remove(String key) {
        entries.remove(key);
    }

    public synchronized void clear() {
        entries.clear();
    }

    public synchronized boolean has(String key) {
        return get(key) != null;
    }

    private static final class CacheEntry<T> {
        final T value;
        final long timestamp;

        CacheEntry(T value, long timestamp) {
            this.value = value;
            this.timestamp = timestamp;
        }
    }
}
