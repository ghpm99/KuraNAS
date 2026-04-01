package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.MusicPlayerState;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.Track;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertTrue;

public class MusicMapperTest {

    @Test
    public void trackFromJson_withFullStructure_mapsCorrectly() throws JSONException {
        JSONObject metadata = new JSONObject();
        metadata.put("title", "My Song");
        metadata.put("artist", "Artist Name");
        metadata.put("album", "Album Name");
        metadata.put("genre", "Rock");
        metadata.put("year", "2025");
        metadata.put("track_number", 3);
        metadata.put("length", 245.5);
        metadata.put("format", "flac");

        JSONObject file = new JSONObject();
        file.put("id", 100);
        file.put("name", "my_song.flac");
        file.put("format", "flac");
        file.put("metadata", metadata);

        JSONObject json = new JSONObject();
        json.put("id", 10);
        json.put("position", 5);
        json.put("added_at", "2026-03-01T12:00:00Z");
        json.put("file", file);

        Track track = MusicMapper.trackFromJson(json);

        assertEquals(10, track.getId());
        assertEquals(100, track.getFileId());
        assertEquals(5, track.getPosition());
        assertEquals("my_song.flac", track.getName());
        assertEquals("My Song", track.getTitle());
        assertEquals("Artist Name", track.getArtist());
        assertEquals("Album Name", track.getAlbum());
        assertEquals("Rock", track.getGenre());
        assertEquals("2025", track.getYear());
        assertEquals(3, track.getTrackNumber());
        assertEquals(245.5, track.getDurationSeconds(), 0.001);
        assertEquals("flac", track.getFormat());
        assertEquals("2026-03-01T12:00:00Z", track.getAddedAt());
    }

    @Test
    public void trackFromJson_withoutFileObject_usesDefaults() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("id", 7);
        json.put("position", 1);

        Track track = MusicMapper.trackFromJson(json);

        assertEquals(7, track.getId());
        assertEquals(7, track.getFileId());
        assertEquals(1, track.getPosition());
        assertEquals("", track.getName());
        assertEquals("", track.getTitle());
        assertEquals("", track.getArtist());
        assertEquals("", track.getAlbum());
        assertEquals(0.0, track.getDurationSeconds(), 0.001);
    }

    @Test
    public void trackFromJson_withFileButNoMetadata_usesFileFormat() throws JSONException {
        JSONObject file = new JSONObject();
        file.put("id", 50);
        file.put("name", "track.mp3");
        file.put("format", "mp3");

        JSONObject json = new JSONObject();
        json.put("id", 1);
        json.put("file", file);

        Track track = MusicMapper.trackFromJson(json);

        assertEquals(50, track.getFileId());
        assertEquals("track.mp3", track.getName());
        assertEquals("mp3", track.getFormat());
        assertEquals("", track.getTitle());
        assertEquals("", track.getArtist());
    }

    @Test
    public void trackFromJson_withDirectFileDto_mapsCorrectly() throws JSONException {
        JSONObject metadata = new JSONObject();
        metadata.put("title", "Direct Song");
        metadata.put("artist", "Direct Artist");
        metadata.put("track_number", "2/12");
        metadata.put("duration", 187.0);

        JSONObject json = new JSONObject();
        json.put("id", 77);
        json.put("name", "direct_song.mp3");
        json.put("format", ".mp3");
        json.put("metadata", metadata);

        Track track = MusicMapper.trackFromJson(json);

        assertEquals(77, track.getId());
        assertEquals(77, track.getFileId());
        assertEquals("direct_song.mp3", track.getName());
        assertEquals("Direct Song", track.getTitle());
        assertEquals("Direct Artist", track.getArtist());
        assertEquals(2, track.getTrackNumber());
        assertEquals(187.0, track.getDurationSeconds(), 0.001);
        assertEquals(".mp3", track.getFormat());
    }

    @Test
    public void trackListFromJson_withItems_parsesPaginated() throws JSONException {
        JSONObject file1 = new JSONObject();
        file1.put("id", 10);
        file1.put("name", "song1.mp3");
        file1.put("format", "mp3");

        JSONObject track1 = new JSONObject();
        track1.put("id", 1);
        track1.put("position", 0);
        track1.put("file", file1);

        JSONObject file2 = new JSONObject();
        file2.put("id", 20);
        file2.put("name", "song2.mp3");
        file2.put("format", "mp3");

        JSONObject track2 = new JSONObject();
        track2.put("id", 2);
        track2.put("position", 1);
        track2.put("file", file2);

        JSONArray items = new JSONArray();
        items.put(track1);
        items.put(track2);

        JSONObject pagination = new JSONObject();
        pagination.put("page", 1);
        pagination.put("page_size", 20);
        pagination.put("has_next", false);
        pagination.put("has_prev", false);

        JSONObject json = new JSONObject();
        json.put("items", items);
        json.put("pagination", pagination);

        PaginatedResult<Track> result = MusicMapper.trackListFromJson(json);

        assertNotNull(result);
        assertEquals(2, result.getItems().size());
        assertEquals("song1.mp3", result.getItems().get(0).getName());
        assertEquals("song2.mp3", result.getItems().get(1).getName());
        assertFalse(result.hasNext());
    }

    @Test
    public void trackListFromJson_withDirectFileItems_parsesPaginated() throws JSONException {
        JSONObject metadata = new JSONObject();
        metadata.put("title", "Library Song");
        metadata.put("artist", "Library Artist");
        metadata.put("length", 120.5);

        JSONObject file = new JSONObject();
        file.put("id", 11);
        file.put("name", "library_song.mp3");
        file.put("format", ".mp3");
        file.put("metadata", metadata);

        JSONArray items = new JSONArray();
        items.put(file);

        JSONObject pagination = new JSONObject();
        pagination.put("page", 1);
        pagination.put("page_size", 30);
        pagination.put("has_next", false);
        pagination.put("has_prev", false);

        JSONObject json = new JSONObject();
        json.put("items", items);
        json.put("pagination", pagination);

        PaginatedResult<Track> result = MusicMapper.trackListFromJson(json);

        assertEquals(1, result.getItems().size());
        assertEquals(11, result.getItems().get(0).getFileId());
        assertEquals("Library Song", result.getItems().get(0).getTitle());
        assertEquals("Library Artist", result.getItems().get(0).getArtist());
    }

    @Test
    public void playerStateFromJson_withAllFields_mapsCorrectly() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("id", 1);
        json.put("client_id", "client-abc");
        json.put("playlist_id", 42);
        json.put("current_file_id", 99);
        json.put("current_position", 120.5);
        json.put("volume", 0.8);
        json.put("shuffle", true);
        json.put("repeat_mode", "all");
        json.put("updated_at", "2026-03-20T15:00:00Z");

        MusicPlayerState state = MusicMapper.playerStateFromJson(json);

        assertEquals(1, state.getId());
        assertEquals("client-abc", state.getClientId());
        assertEquals(42, state.getPlaylistId());
        assertEquals(99, state.getCurrentFileId());
        assertEquals(120.5, state.getCurrentPosition(), 0.001);
        assertEquals(0.8, state.getVolume(), 0.001);
        assertTrue(state.isShuffle());
        assertEquals("all", state.getRepeatMode());
        assertEquals("2026-03-20T15:00:00Z", state.getUpdatedAt());
    }

    @Test
    public void playerStateFromJson_withNullOptionalFields_usesDefaults() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("id", 1);
        json.put("client_id", "client-xyz");

        MusicPlayerState state = MusicMapper.playerStateFromJson(json);

        assertEquals(1, state.getId());
        assertEquals("client-xyz", state.getClientId());
        assertEquals(0, state.getPlaylistId());
        assertEquals(0, state.getCurrentFileId());
        assertEquals(0.0, state.getCurrentPosition(), 0.001);
        assertEquals(1.0, state.getVolume(), 0.001);
        assertFalse(state.isShuffle());
        assertEquals("none", state.getRepeatMode());
        assertEquals("", state.getUpdatedAt());
    }
}
