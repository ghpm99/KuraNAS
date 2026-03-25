package com.kuranas.mobile.data.mapper;

import com.kuranas.mobile.domain.model.AppSettings;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

public final class SettingsMapper {

    private SettingsMapper() {
    }

    public static AppSettings fromJson(JSONObject json) throws JSONException {
        JSONObject players = json.optJSONObject("players");
        boolean rememberMusicQueue = false;
        boolean rememberVideoProgress = false;
        boolean autoplayNextVideo = false;
        int imageSlideshowSeconds = 5;

        if (players != null) {
            rememberMusicQueue = players.optBoolean("remember_music_queue", false);
            rememberVideoProgress = players.optBoolean("remember_video_progress", false);
            autoplayNextVideo = players.optBoolean("autoplay_next_video", false);
            imageSlideshowSeconds = players.optInt("image_slideshow_seconds", 5);
        }

        JSONObject language = json.optJSONObject("language");
        String currentLanguage = "";
        List<String> availableLanguages = new ArrayList<String>();

        if (language != null) {
            currentLanguage = language.optString("current", "");
            JSONArray available = language.optJSONArray("available");
            if (available != null) {
                for (int i = 0; i < available.length(); i++) {
                    availableLanguages.add(available.optString(i, ""));
                }
            }
        }

        return new AppSettings(
                rememberMusicQueue,
                rememberVideoProgress,
                autoplayNextVideo,
                imageSlideshowSeconds,
                currentLanguage,
                availableLanguages
        );
    }
}
