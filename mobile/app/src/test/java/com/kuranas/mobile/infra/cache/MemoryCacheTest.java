package com.kuranas.mobile.infra.cache;

import org.junit.Before;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNull;
import static org.junit.Assert.assertTrue;

public class MemoryCacheTest {

    private MemoryCache<String> cache;

    @Before
    public void setUp() {
        cache = new MemoryCache<String>(5000L);
    }

    @Test
    public void putAndGet_returnsStoredValue() {
        cache.put("key1", "value1");

        assertEquals("value1", cache.get("key1"));
    }

    @Test
    public void get_nonExistentKey_returnsNull() {
        assertNull(cache.get("missing"));
    }

    @Test
    public void put_overwritesExistingValue() {
        cache.put("key1", "original");
        cache.put("key1", "updated");

        assertEquals("updated", cache.get("key1"));
    }

    @Test
    public void ttlExpiration_returnsNullAfterExpiry() throws InterruptedException {
        MemoryCache<String> shortCache = new MemoryCache<String>(50L);
        shortCache.put("key1", "value1");

        assertEquals("value1", shortCache.get("key1"));

        Thread.sleep(100);

        assertNull(shortCache.get("key1"));
    }

    @Test
    public void remove_deletesEntry() {
        cache.put("key1", "value1");
        cache.put("key2", "value2");

        cache.remove("key1");

        assertNull(cache.get("key1"));
        assertEquals("value2", cache.get("key2"));
    }

    @Test
    public void remove_nonExistentKey_doesNotThrow() {
        cache.remove("nonexistent");
    }

    @Test
    public void clear_removesAllEntries() {
        cache.put("key1", "value1");
        cache.put("key2", "value2");
        cache.put("key3", "value3");

        cache.clear();

        assertNull(cache.get("key1"));
        assertNull(cache.get("key2"));
        assertNull(cache.get("key3"));
    }

    @Test
    public void has_returnsTrueForExistingKey() {
        cache.put("key1", "value1");

        assertTrue(cache.has("key1"));
    }

    @Test
    public void has_returnsFalseForMissingKey() {
        assertFalse(cache.has("missing"));
    }

    @Test
    public void has_returnsFalseAfterExpiry() throws InterruptedException {
        MemoryCache<String> shortCache = new MemoryCache<String>(50L);
        shortCache.put("key1", "value1");

        assertTrue(shortCache.has("key1"));

        Thread.sleep(100);

        assertFalse(shortCache.has("key1"));
    }
}
