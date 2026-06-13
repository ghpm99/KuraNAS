package com.kuranas.mobile.presentation.home;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

import org.junit.Test;

public class BackoffTest {

    private static final long BASE = 60_000L;
    private static final long MAX = 300_000L;

    @Test
    public void startsAtBaseAndNotFailing() {
        Backoff backoff = new Backoff(BASE, MAX);
        assertEquals(BASE, backoff.currentDelayMs());
        assertFalse(backoff.isFailing());
    }

    @Test
    public void failuresDoubleUntilCeiling() {
        Backoff backoff = new Backoff(BASE, MAX);
        assertEquals(120_000L, backoff.recordFailure());
        assertEquals(240_000L, backoff.recordFailure());
        assertEquals(300_000L, backoff.recordFailure()); // 480k capped at 300k
        assertEquals(300_000L, backoff.recordFailure()); // stays at ceiling
        assertTrue(backoff.isFailing());
    }

    @Test
    public void successResetsToBase() {
        Backoff backoff = new Backoff(BASE, MAX);
        backoff.recordFailure();
        backoff.recordFailure();
        backoff.recordSuccess();
        assertEquals(BASE, backoff.currentDelayMs());
        assertFalse(backoff.isFailing());
    }

    @Test
    public void emailCadenceCapsAtCeiling() {
        Backoff backoff = new Backoff(120_000L, MAX);
        assertEquals(240_000L, backoff.recordFailure());
        assertEquals(300_000L, backoff.recordFailure());
        assertEquals(300_000L, backoff.recordFailure());
    }

    @Test(expected = IllegalArgumentException.class)
    public void rejectsMaxBelowBase() {
        new Backoff(120_000L, 60_000L);
    }
}
