package com.kuranas.mobile.presentation.home;

/**
 * Exponential backoff for the kiosk pollers. While the server answers, polling
 * runs at the base interval; on each consecutive failure the delay doubles up to
 * a ceiling, and the first success resets it. Pure logic (no Android, no Handler)
 * so the scheduling rules are unit-testable; the {@code HomeFragment} owns the
 * actual {@code Handler.postDelayed} loop.
 */
public final class Backoff {

    private final long baseMs;
    private final long maxMs;
    private long currentMs;
    private boolean failing;

    public Backoff(long baseMs, long maxMs) {
        if (baseMs <= 0 || maxMs < baseMs) {
            throw new IllegalArgumentException("require 0 < baseMs <= maxMs");
        }
        this.baseMs = baseMs;
        this.maxMs = maxMs;
        this.currentMs = baseMs;
        this.failing = false;
    }

    /** The delay to wait before the next poll. */
    public long currentDelayMs() {
        return currentMs;
    }

    /** True once at least one failure has pushed the delay above the base interval. */
    public boolean isFailing() {
        return failing;
    }

    /** A successful poll restores the normal cadence. */
    public void recordSuccess() {
        currentMs = baseMs;
        failing = false;
    }

    /** A failed poll doubles the delay, capped at the ceiling, and returns it. */
    public long recordFailure() {
        failing = true;
        long doubled = currentMs * 2;
        currentMs = doubled > maxMs ? maxMs : doubled;
        return currentMs;
    }
}
