package com.kuranas.mobile.domain.model;

import org.junit.Test;

import static org.junit.Assert.assertEquals;

public class VideoItemTest {

    private VideoItem createVideoItem(String name) {
        return new VideoItem(1, name, "/videos/" + name, "/videos", "mp4", 1024L,
                "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z");
    }

    @Test
    public void getDisplayName_removesExtension() {
        VideoItem item = createVideoItem("movie.mp4");
        assertEquals("movie", item.getDisplayName());
    }

    @Test
    public void getDisplayName_removesLastExtensionOnly() {
        VideoItem item = createVideoItem("my.video.file.mkv");
        assertEquals("my.video.file", item.getDisplayName());
    }

    @Test
    public void getDisplayName_withNoExtension_returnsFullName() {
        VideoItem item = createVideoItem("noextension");
        assertEquals("noextension", item.getDisplayName());
    }

    @Test
    public void getDisplayName_withNullName_returnsEmptyString() {
        VideoItem item = new VideoItem(1, null, "/videos/null", "/videos", "mp4",
                1024L, "", "");
        assertEquals("", item.getDisplayName());
    }

    @Test
    public void getDisplayName_withDotAtStart_returnsFullName() {
        VideoItem item = createVideoItem(".hidden");
        assertEquals(".hidden", item.getDisplayName());
    }
}
