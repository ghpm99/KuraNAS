package com.kuranas.mobile.infra.kiosk;

import android.app.Activity;
import android.view.View;
import android.view.WindowManager;

/**
 * Manages kiosk mode behavior for activities: keeps the screen always on,
 * hides system UI (status bar and navigation bar), and re-hides system UI
 * whenever it becomes visible again (e.g. after user touch on API 16).
 */
public class KioskManager {

    private static final int FULLSCREEN_FLAGS =
            View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                    | View.SYSTEM_UI_FLAG_FULLSCREEN
                    | View.SYSTEM_UI_FLAG_LOW_PROFILE;

    private final Activity activity;

    public KioskManager(Activity activity) {
        this.activity = activity;
    }

    public void engage() {
        keepScreenOn();
        applyFullscreen();
        registerSystemUiListener();
    }

    private void keepScreenOn() {
        activity.getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
    }

    private void applyFullscreen() {
        View decorView = activity.getWindow().getDecorView();
        decorView.setSystemUiVisibility(FULLSCREEN_FLAGS);
    }

    private void registerSystemUiListener() {
        final View decorView = activity.getWindow().getDecorView();
        decorView.setOnSystemUiVisibilityChangeListener(
                new View.OnSystemUiVisibilityChangeListener() {
                    @Override
                    public void onSystemUiVisibilityChange(int visibility) {
                        boolean barsVisible = (visibility & View.SYSTEM_UI_FLAG_HIDE_NAVIGATION) == 0;
                        if (barsVisible) {
                            decorView.setSystemUiVisibility(FULLSCREEN_FLAGS);
                        }
                    }
                });
    }
}
