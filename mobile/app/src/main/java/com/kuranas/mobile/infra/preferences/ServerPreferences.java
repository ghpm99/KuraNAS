package com.kuranas.mobile.infra.preferences;

import android.content.Context;
import android.content.SharedPreferences;

public final class ServerPreferences {

    private static final String PREFS_NAME = "kuranas_server";
    private static final String KEY_SERVER_URL = "server_url";
    private static final String KEY_LAST_CONNECTED = "last_connected";

    private final SharedPreferences prefs;

    public ServerPreferences(Context context) {
        this.prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE);
    }

    public String getServerUrl() {
        return prefs.getString(KEY_SERVER_URL, null);
    }

    public void saveServerUrl(String url) {
        prefs.edit()
                .putString(KEY_SERVER_URL, url)
                .putLong(KEY_LAST_CONNECTED, System.currentTimeMillis())
                .apply();
    }

    public long getLastConnected() {
        return prefs.getLong(KEY_LAST_CONNECTED, 0);
    }

    public void clear() {
        prefs.edit().clear().apply();
    }
}
