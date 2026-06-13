package com.kuranas.mobile.data.mapper;

import java.text.ParseException;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.Locale;

/**
 * Parses the backend's RFC3339 timestamps into a device-local {@code HH:mm}
 * string for the kiosk side-labels. Kept tiny and free of Android APIs so it can
 * be unit-tested, and kept on the {@code SimpleDateFormat} parser path because
 * the target tablet (Android 4.1, minSdk 16) lacks {@code java.time} and the
 * {@code X} timezone pattern (API 24+) — hence the manual zone normalization.
 */
public final class TimeFormat {

    private TimeFormat() {
    }

    /** Returns the local {@code HH:mm} of an RFC3339 timestamp, or "" if unparseable. */
    public static String shortTime(String iso) {
        Date date = parse(iso);
        if (date == null) {
            return "";
        }
        return new SimpleDateFormat("HH:mm", Locale.getDefault()).format(date);
    }

    static Date parse(String iso) {
        if (iso == null) {
            return null;
        }
        String s = iso.trim();
        if (s.isEmpty()) {
            return null;
        }

        // Drop fractional seconds (".123456") — not supported uniformly by the parser.
        s = s.replaceAll("\\.\\d+", "");

        if (s.endsWith("Z")) {
            // RFC822 zone parsing wants "+0000", not the literal "Z".
            String withZone = s.substring(0, s.length() - 1) + "+0000";
            return tryParse(withZone, "yyyy-MM-dd'T'HH:mm:ssZ");
        }

        // Collapse "+03:00" / "-03:00" into the RFC822 "+0300" the parser accepts.
        String collapsed = s.replaceAll("([+-]\\d{2}):(\\d{2})$", "$1$2");
        if (!collapsed.equals(s) || collapsed.matches(".*[+-]\\d{4}$")) {
            Date withZone = tryParse(collapsed, "yyyy-MM-dd'T'HH:mm:ssZ");
            if (withZone != null) {
                return withZone;
            }
        }

        // No timezone at all: parse in the device's local zone.
        return tryParse(s, "yyyy-MM-dd'T'HH:mm:ss");
    }

    private static Date tryParse(String value, String pattern) {
        try {
            return new SimpleDateFormat(pattern, Locale.US).parse(value);
        } catch (ParseException e) {
            return null;
        }
    }
}
