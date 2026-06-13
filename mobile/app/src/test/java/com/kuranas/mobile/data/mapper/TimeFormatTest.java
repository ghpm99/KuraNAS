package com.kuranas.mobile.data.mapper;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNull;

import org.junit.Test;

import java.util.Date;
import java.util.TimeZone;

public class TimeFormatTest {

    // 2026-06-13T10:20:30Z, epoch ms (zone-independent).
    private static final long EPOCH_UTC = 1781346030000L;

    @Test
    public void parse_utcZuluTimestamp_isAbsoluteInstant() {
        Date date = TimeFormat.parse("2026-06-13T10:20:30Z");
        assertEquals(EPOCH_UTC, date.getTime());
    }

    @Test
    public void parse_stripsFractionalSeconds() {
        Date date = TimeFormat.parse("2026-06-13T10:20:30.123456Z");
        assertEquals(EPOCH_UTC, date.getTime());
    }

    @Test
    public void parse_offsetWithColon_isHonored() {
        // -03:00 means the same wall time is 3h later in UTC.
        Date date = TimeFormat.parse("2026-06-13T10:20:30-03:00");
        assertEquals(EPOCH_UTC + 3 * 3600_000L, date.getTime());
    }

    @Test
    public void parse_returnsNullForGarbage() {
        assertNull(TimeFormat.parse(null));
        assertNull(TimeFormat.parse(""));
        assertNull(TimeFormat.parse("not-a-date"));
    }

    @Test
    public void shortTime_formatsInLocalZone() {
        TimeZone original = TimeZone.getDefault();
        try {
            TimeZone.setDefault(TimeZone.getTimeZone("America/Sao_Paulo")); // UTC-3
            assertEquals("07:20", TimeFormat.shortTime("2026-06-13T10:20:30Z"));
        } finally {
            TimeZone.setDefault(original);
        }
    }

    @Test
    public void shortTime_emptyForUnparseable() {
        assertEquals("", TimeFormat.shortTime("garbage"));
    }
}
