package com.kuranas.mobile.i18n;

import android.content.Context;

import com.kuranas.mobile.infra.http.HttpClient;
import com.kuranas.mobile.infra.http.HttpResponse;
import com.kuranas.mobile.infra.logging.AppLogger;

import org.json.JSONException;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;

public final class TranslationManager {

    private static final String LOG_TAG = "TranslationManager";
    private static final String LOCAL_TRANSLATIONS_PATH = "translations/pt-BR.json";

    private static TranslationManager instance;

    private HttpClient httpClient;
    private volatile Map<String, String> translations;
    private volatile boolean loaded;

    public TranslationManager(HttpClient httpClient) {
        this.httpClient = httpClient;
        this.translations = new HashMap<String, String>();
        this.loaded = false;
    }

    private TranslationManager() {
        this.translations = new HashMap<String, String>();
        this.loaded = false;
    }

    public static synchronized void initInstance(Context context) {
        if (instance == null) {
            instance = new TranslationManager();
            instance.loadLocalFallback(context);
        }
    }

    public static synchronized TranslationManager getInstance() {
        if (instance == null) {
            throw new IllegalStateException("TranslationManager not initialized. Call initInstance() first.");
        }
        return instance;
    }

    public void setHttpClient(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    public String get(String key) {
        return t(key);
    }

    public void loadLocalFallback(Context context) {
        try {
            InputStream is = context.getAssets().open(LOCAL_TRANSLATIONS_PATH);
            BufferedReader reader = new BufferedReader(new InputStreamReader(is, "UTF-8"));
            try {
                StringBuilder sb = new StringBuilder();
                String line;
                while ((line = reader.readLine()) != null) {
                    sb.append(line);
                }
                parseTranslations(sb.toString());
                loaded = true;
                AppLogger.i(LOG_TAG, "Loaded " + translations.size() + " local fallback translations");
            } finally {
                reader.close();
                is.close();
            }
        } catch (IOException e) {
            AppLogger.e(LOG_TAG, "Failed to load local fallback translations", e);
        } catch (JSONException e) {
            AppLogger.e(LOG_TAG, "Failed to parse local fallback translations", e);
        }
    }

    public void loadSync() {
        HttpResponse response = httpClient.getSync("/api/v1/configuration/translation");
        if (response.isSuccessful()) {
            try {
                parseTranslations(response.getBody());
                loaded = true;
                AppLogger.i(LOG_TAG, "Loaded " + translations.size() + " remote translations");
            } catch (JSONException e) {
                AppLogger.e(LOG_TAG, "Failed to parse translations", e);
            }
        } else {
            AppLogger.w(LOG_TAG, "Failed to load remote translations (status " + response.getStatusCode() + "), using local fallback");
        }
    }

    public void loadAsync(final Runnable onComplete) {
        httpClient.get("/api/v1/configuration/translation", new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                if (response.isSuccessful()) {
                    try {
                        parseTranslations(response.getBody());
                        loaded = true;
                        AppLogger.i(LOG_TAG, "Loaded " + translations.size() + " remote translations");
                    } catch (JSONException e) {
                        AppLogger.e(LOG_TAG, "Failed to parse translations", e);
                    }
                } else {
                    AppLogger.w(LOG_TAG, "Failed to load remote translations (status " + response.getStatusCode() + "), using local fallback");
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
