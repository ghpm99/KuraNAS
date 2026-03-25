package com.kuranas.mobile.infra.cache;

import android.graphics.Bitmap;
import android.util.LruCache;

public final class BitmapCache {

    private final LruCache<String, Bitmap> cache;

    public BitmapCache() {
        int maxMemory = (int) (Runtime.getRuntime().maxMemory() / 1024);
        int cacheSize = maxMemory / 8;
        cache = new LruCache<String, Bitmap>(cacheSize) {
            @Override
            protected int sizeOf(String key, Bitmap bitmap) {
                return bitmap.getRowBytes() * bitmap.getHeight() / 1024;
            }
        };
    }

    public Bitmap get(String key) {
        return cache.get(key);
    }

    public void put(String key, Bitmap bitmap) {
        if (bitmap != null && get(key) == null) {
            cache.put(key, bitmap);
        }
    }

    public void clear() {
        cache.evictAll();
    }
}
