package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.VideoItem;
import com.kuranas.mobile.domain.model.VideoPlaybackState;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertTrue;

public class VideoMapperTest {

    @Test
    public void fromJson_withValidData_mapsCorrectly() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("id", 5);
        json.put("name", "movie.mp4");
        json.put("path", "/videos/movie.mp4");
        json.put("parent_path", "/videos");
        json.put("format", "mp4");
        json.put("size", 5242880L);
        json.put("created_at", "2026-02-01T09:00:00Z");
        json.put("updated_at", "2026-02-10T14:00:00Z");

        VideoItem item = VideoMapper.fromJson(json);

        assertEquals(5, item.getId());
        assertEquals("movie.mp4", item.getName());
        assertEquals("/videos/movie.mp4", item.getPath());
        assertEquals("/videos", item.getParentPath());
        assertEquals("mp4", item.getFormat());
        assertEquals(5242880L, item.getSize());
        assertEquals("2026-02-01T09:00:00Z", item.getCreatedAt());
        assertEquals("2026-02-10T14:00:00Z", item.getUpdatedAt());
    }

    @Test
    public void fromJson_withMissingFields_usesDefaults() throws JSONException {
        JSONObject json = new JSONObject();

        VideoItem item = VideoMapper.fromJson(json);

        assertEquals(0, item.getId());
        assertEquals("", item.getName());
        assertEquals("", item.getPath());
        assertEquals("", item.getParentPath());
        assertEquals("", item.getFormat());
        assertEquals(0L, item.getSize());
    }

    @Test
    public void fromPaginatedJson_withItems_parsesList() throws JSONException {
        JSONObject v1 = new JSONObject();
        v1.put("id", 1);
        v1.put("name", "clip1.mkv");

        JSONObject v2 = new JSONObject();
        v2.put("id", 2);
        v2.put("name", "clip2.avi");

        JSONArray items = new JSONArray();
        items.put(v1);
        items.put(v2);

        JSONObject pagination = new JSONObject();
        pagination.put("page", 2);
        pagination.put("page_size", 10);
        pagination.put("has_next", false);
        pagination.put("has_prev", true);

        JSONObject json = new JSONObject();
        json.put("items", items);
        json.put("pagination", pagination);

        PaginatedResult<VideoItem> result = VideoMapper.fromPaginatedJson(json);

        assertNotNull(result);
        assertEquals(2, result.getItems().size());
        assertEquals("clip1.mkv", result.getItems().get(0).getName());
        assertEquals("clip2.avi", result.getItems().get(1).getName());
        assertFalse(result.hasNext());
        assertTrue(result.getPagination().hasPrev());
    }

    @Test
    public void playbackStateFromJson_withAllFields_mapsCorrectly() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("id", 10);
        json.put("client_id", "device-001");
        json.put("playlist_id", 3);
        json.put("video_id", 77);
        json.put("current_time", 300.0);
        json.put("duration", 7200.0);
        json.put("is_paused", true);
        json.put("completed", false);
        json.put("last_update", "2026-03-20T18:00:00Z");

        VideoPlaybackState state = VideoMapper.playbackStateFromJson(json);

        assertEquals(10, state.getId());
        assertEquals("device-001", state.getClientId());
        assertEquals(3, state.getPlaylistId());
        assertEquals(77, state.getVideoId());
        assertEquals(300.0, state.getCurrentTime(), 0.001);
        assertEquals(7200.0, state.getDuration(), 0.001);
        assertTrue(state.isPaused());
        assertFalse(state.isCompleted());
        assertEquals("2026-03-20T18:00:00Z", state.getLastUpdate());
    }

    @Test
    public void playbackStateFromJson_withNullOptionalFields_usesDefaults() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("id", 1);
        json.put("client_id", "device-002");

        VideoPlaybackState state = VideoMapper.playbackStateFromJson(json);

        assertEquals(1, state.getId());
        assertEquals("device-002", state.getClientId());
        assertEquals(0, state.getPlaylistId());
        assertEquals(0, state.getVideoId());
        assertEquals(0.0, state.getCurrentTime(), 0.001);
        assertEquals(0.0, state.getDuration(), 0.001);
        assertFalse(state.isPaused());
        assertFalse(state.isCompleted());
        assertEquals("", state.getLastUpdate());
    }
}
