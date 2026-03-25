package com.kuranas.mobile.data.remote.api;

import com.kuranas.mobile.data.mapper.SettingsMapper;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.AppSettings;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.infra.http.HttpClient;
import com.kuranas.mobile.infra.http.HttpResponse;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

public final class ConfigApi {

    private final HttpClient httpClient;

    public ConfigApi(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    public void getSettings(final ApiCallback<AppSettings> callback) {
        String path = "/api/v1/configuration/settings";
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                if (response.getError() != null) {
                    callback.onError(AppError.networkUnavailable(response.getError()));
                    return;
                }
                if (!response.isSuccessful()) {
                    callback.onError(AppError.fromHttpResponse(response.getStatusCode(), null));
                    return;
                }
                try {
                    JSONObject json = response.toJsonObject();
                    AppSettings settings = SettingsMapper.fromJson(json);
                    callback.onSuccess(settings);
                } catch (JSONException e) {
                    callback.onError(AppError.invalidPayload(e));
                }
            }
        });
    }

    public void updateLanguage(final String language, final ApiCallback<AppSettings> callback) {
        // First fetch current settings to preserve other fields
        getSettings(new ApiCallback<AppSettings>() {
            @Override
            public void onSuccess(AppSettings currentSettings) {
                try {
                    JSONObject body = new JSONObject();

                    JSONObject players = new JSONObject();
                    players.put("remember_music_queue", currentSettings.isRememberMusicQueue());
                    players.put("remember_video_progress", currentSettings.isRememberVideoProgress());
                    players.put("autoplay_next_video", currentSettings.isAutoplayNextVideo());
                    players.put("image_slideshow_seconds", currentSettings.getImageSlideshowSeconds());
                    body.put("players", players);

                    JSONObject languageObj = new JSONObject();
                    languageObj.put("current", language);
                    JSONArray available = new JSONArray();
                    if (currentSettings.getAvailableLanguages() != null) {
                        for (int i = 0; i < currentSettings.getAvailableLanguages().size(); i++) {
                            available.put(currentSettings.getAvailableLanguages().get(i));
                        }
                    }
                    languageObj.put("available", available);
                    body.put("language", languageObj);

                    String path = "/api/v1/configuration/settings";
                    httpClient.put(path, body.toString(), new HttpClient.Callback() {
                        @Override
                        public void onResponse(HttpResponse response) {
                            if (response.getError() != null) {
                                callback.onError(AppError.networkUnavailable(response.getError()));
                                return;
                            }
                            if (!response.isSuccessful()) {
                                callback.onError(AppError.fromHttpResponse(response.getStatusCode(), null));
                                return;
                            }
                            try {
                                JSONObject json = response.toJsonObject();
                                AppSettings settings = SettingsMapper.fromJson(json);
                                callback.onSuccess(settings);
                            } catch (JSONException e) {
                                callback.onError(AppError.invalidPayload(e));
                            }
                        }
                    });
                } catch (JSONException e) {
                    callback.onError(AppError.invalidPayload(e));
                }
            }

            @Override
            public void onError(AppError error) {
                callback.onError(error);
            }
        });
    }
}
