package com.kuranas.mobile.app;

import android.app.Application;

import com.kuranas.mobile.infra.logging.AppLogger;
import com.kuranas.mobile.presentation.base.BaseFragment;

public class KuranasApp extends Application {

    private static final String LOG_TAG = "KuranasApp";

    @Override
    public void onCreate() {
        super.onCreate();
        AppLogger.i(LOG_TAG, "Application starting");

        ServiceLocator locator = ServiceLocator.getInstance();
        BaseFragment.setTranslationManager(locator.getTranslationManager());

        locator.getTranslationManager().loadLocalFallback(this);

        locator.getTranslationManager().loadAsync(new Runnable() {
            @Override
            public void run() {
                AppLogger.i(LOG_TAG, "Remote translations loaded");
            }
        });
    }

    @Override
    public void onTerminate() {
        super.onTerminate();
        ServiceLocator.getInstance().shutdown();
    }
}
