package com.kuranas.mobile.app;

import android.app.Application;

import com.kuranas.mobile.i18n.TranslationManager;
import com.kuranas.mobile.infra.logging.AppLogger;
import com.kuranas.mobile.presentation.base.BaseFragment;

public class KuranasApp extends Application {

    private static final String LOG_TAG = "KuranasApp";

    @Override
    public void onCreate() {
        super.onCreate();
        AppLogger.i(LOG_TAG, "Application starting");

        TranslationManager.initInstance(this);
        BaseFragment.setTranslationManager(TranslationManager.getInstance());
    }

    @Override
    public void onTerminate() {
        super.onTerminate();
        try {
            ServiceLocator.getInstance().shutdown();
        } catch (IllegalStateException e) {
            // ServiceLocator was never initialized
        }
    }
}
