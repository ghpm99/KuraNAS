package com.kuranas.mobile.i18n;

import com.kuranas.mobile.infra.http.HttpClient;
import com.kuranas.mobile.infra.http.HttpResponse;
import com.kuranas.mobile.infra.logging.AppLogger;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;

public final class TranslationManager {

    private static final String LOG_TAG = "TranslationManager";

    private final HttpClient httpClient;
    private volatile Map<String, String> translations;
    private volatile boolean loaded;

    public TranslationManager(HttpClient httpClient) {
        this.httpClient = httpClient;
        this.translations = new HashMap<String, String>();
        this.loaded = false;
    }

    public void loadSync() {
        HttpResponse response = httpClient.getSync("/configuration/translation");
        if (response.isSuccessful()) {
            try {
                parseTranslations(response.getBody());
                loaded = true;
                AppLogger.i(LOG_TAG, "Loaded " + translations.size() + " translations");
            } catch (JSONException e) {
                AppLogger.e(LOG_TAG, "Failed to parse translations", e);
            }
        } else {
            AppLogger.e(LOG_TAG, "Failed to load translations: " + response.getStatusCode());
        }
    }

    public void loadAsync(final Runnable onComplete) {
        httpClient.get("/configuration/translation", new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                if (response.isSuccessful()) {
                    try {
                        parseTranslations(response.getBody());
                        loaded = true;
                        AppLogger.i(LOG_TAG, "Loaded " + translations.size() + " translations");
                    } catch (JSONException e) {
                        AppLogger.e(LOG_TAG, "Failed to parse translations", e);
                    }
                } else {
                    AppLogger.e(LOG_TAG, "Failed to load translations: " + response.getStatusCode());
                }
                if (onComplete != null) {
                    onComplete.run();
                }
            }
        });
    }

    public String t(String key) {
        return t(key, key);
    }

    public String t(String key, String fallback) {
        String value = translations.get(key);
        if (value != null) {
            return value;
        }
        if (loaded) {
            AppLogger.w(LOG_TAG, "Missing translation key: " + key);
        }
        return fallback;
    }

    public boolean isLoaded() {
        return loaded;
    }

    public int size() {
        return translations.size();
    }

    private void parseTranslations(String json) throws JSONException {
        Map<String, String> result = new HashMap<String, String>();
        JSONObject root = new JSONObject(json);
        flattenJson("", root, result);
        translations = result;
    }

    private void flattenJson(String prefix, JSONObject obj, Map<String, String> result) throws JSONException {
        Iterator<String> keys = obj.keys();
        while (keys.hasNext()) {
            String key = keys.next();
            String fullKey = prefix.isEmpty() ? key : prefix + "." + key;
            Object value = obj.get(key);
            if (value instanceof JSONObject) {
                flattenJson(fullKey, (JSONObject) value, result);
            } else {
                result.put(fullKey, String.valueOf(value));
            }
        }
    }
}
