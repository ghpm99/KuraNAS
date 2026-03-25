package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.AppSettings;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertTrue;

public class SettingsMapperTest {

    @Test
    public void fromJson_withFullStructure_mapsCorrectly() throws JSONException {
        JSONObject players = new JSONObject();
        players.put("remember_music_queue", true);
        players.put("remember_video_progress", true);
        players.put("autoplay_next_video", true);
        players.put("image_slideshow_seconds", 10);

        JSONArray available = new JSONArray();
        available.put("en");
        available.put("pt-BR");
        available.put("es");

        JSONObject language = new JSONObject();
        language.put("current", "pt-BR");
        language.put("available", available);

        JSONObject json = new JSONObject();
        json.put("players", players);
        json.put("language", language);

        AppSettings settings = SettingsMapper.fromJson(json);

        assertTrue(settings.isRememberMusicQueue());
        assertTrue(settings.isRememberVideoProgress());
        assertTrue(settings.isAutoplayNextVideo());
        assertEquals(10, settings.getImageSlideshowSeconds());
        assertEquals("pt-BR", settings.getCurrentLanguage());
        assertNotNull(settings.getAvailableLanguages());
        assertEquals(3, settings.getAvailableLanguages().size());
        assertEquals("en", settings.getAvailableLanguages().get(0));
        assertEquals("pt-BR", settings.getAvailableLanguages().get(1));
        assertEquals("es", settings.getAvailableLanguages().get(2));
    }

    @Test
    public void fromJson_withMissingOptionalSections_usesDefaults() throws JSONException {
        JSONObject json = new JSONObject();

        AppSettings settings = SettingsMapper.fromJson(json);

        assertFalse(settings.isRememberMusicQueue());
        assertFalse(settings.isRememberVideoProgress());
        assertFalse(settings.isAutoplayNextVideo());
        assertEquals(5, settings.getImageSlideshowSeconds());
        assertEquals("", settings.getCurrentLanguage());
        assertNotNull(settings.getAvailableLanguages());
        assertTrue(settings.getAvailableLanguages().isEmpty());
    }

    @Test
    public void fromJson_withPlayersButNoLanguage_mapsPlayersOnly() throws JSONException {
        JSONObject players = new JSONObject();
        players.put("remember_music_queue", true);
        players.put("autoplay_next_video", false);
        players.put("image_slideshow_seconds", 3);

        JSONObject json = new JSONObject();
        json.put("players", players);

        AppSettings settings = SettingsMapper.fromJson(json);

        assertTrue(settings.isRememberMusicQueue());
        assertFalse(settings.isAutoplayNextVideo());
        assertEquals(3, settings.getImageSlideshowSeconds());
        assertEquals("", settings.getCurrentLanguage());
        assertTrue(settings.getAvailableLanguages().isEmpty());
    }
}
