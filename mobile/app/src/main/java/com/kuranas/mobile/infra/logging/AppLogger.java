package com.kuranas.mobile.infra.logging;

import android.util.Log;

public final class AppLogger {

    private static final String TAG = "KuraNAS";
    private static boolean debugEnabled = true;

    private AppLogger() {
    }

    public static void setDebugEnabled(boolean enabled) {
        debugEnabled = enabled;
    }

    public static void d(String component, String message) {
        if (debugEnabled) {
            Log.d(TAG, "[" + component + "] " + message);
        }
    }

    public static void i(String component, String message) {
        Log.i(TAG, "[" + component + "] " + message);
    }

    public static void w(String component, String message) {
        Log.w(TAG, "[" + component + "] " + message);
    }

    public static void e(String component, String message) {
        Log.e(TAG, "[" + component + "] " + message);
    }

    public static void e(String component, String message, Throwable throwable) {
        Log.e(TAG, "[" + component + "] " + message, throwable);
    }
}
