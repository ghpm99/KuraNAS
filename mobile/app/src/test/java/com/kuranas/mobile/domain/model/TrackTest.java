package com.kuranas.mobile.domain.model;

import org.junit.Test;

import static org.junit.Assert.assertEquals;

public class TrackTest {

    private Track createTrack(String name, String title, String artist, double duration) {
        return new Track(1, 10, 0, name, title, artist, "Album", "Rock", "2025",
                1, duration, "mp3", "2026-01-01T00:00:00Z");
    }

    @Test
    public void getDisplayTitle_withTitle_returnsTitle() {
        Track track = createTrack("file.mp3", "My Song", "Artist", 180.0);
        assertEquals("My Song", track.getDisplayTitle());
    }

    @Test
    public void getDisplayTitle_withNullTitle_returnsName() {
        Track track = new Track(1, 10, 0, "file.mp3", null, "Artist", "Album",
                "Rock", "2025", 1, 180.0, "mp3", "");
        assertEquals("file.mp3", track.getDisplayTitle());
    }

    @Test
    public void getDisplayTitle_withEmptyTitle_returnsName() {
        Track track = createTrack("file.mp3", "", "Artist", 180.0);
        assertEquals("file.mp3", track.getDisplayTitle());
    }

    @Test
    public void getDisplayArtist_withArtist_returnsArtist() {
        Track track = createTrack("file.mp3", "Title", "Cool Artist", 180.0);
        assertEquals("Cool Artist", track.getDisplayArtist());
    }

    @Test
    public void getDisplayArtist_withNullArtist_returnsEmptyString() {
        Track track = new Track(1, 10, 0, "file.mp3", "Title", null, "Album",
                "Rock", "2025", 1, 180.0, "mp3", "");
        assertEquals("", track.getDisplayArtist());
    }

    @Test
    public void getDisplayArtist_withEmptyArtist_returnsEmptyString() {
        Track track = createTrack("file.mp3", "Title", "", 180.0);
        assertEquals("", track.getDisplayArtist());
    }

    @Test
    public void getFormattedDuration_twoMinutesFiveSeconds() {
        Track track = createTrack("file.mp3", "Title", "Artist", 125.0);
        assertEquals("2:05", track.getFormattedDuration());
    }

    @Test
    public void getFormattedDuration_exactlyOneMinute() {
        Track track = createTrack("file.mp3", "Title", "Artist", 60.0);
        assertEquals("1:00", track.getFormattedDuration());
    }

    @Test
    public void getFormattedDuration_fiveSeconds() {
        Track track = createTrack("file.mp3", "Title", "Artist", 5.0);
        assertEquals("0:05", track.getFormattedDuration());
    }

    @Test
    public void getFormattedDuration_zero() {
        Track track = createTrack("file.mp3", "Title", "Artist", 0.0);
        assertEquals("0:00", track.getFormattedDuration());
    }

    @Test
    public void getFormattedDuration_longDuration() {
        Track track = createTrack("file.mp3", "Title", "Artist", 3661.0);
        assertEquals("61:01", track.getFormattedDuration());
    }
}
