package com.kuranas.mobile.app;

import com.kuranas.mobile.data.remote.api.EmailApi;
import com.kuranas.mobile.data.remote.api.NotificationApi;
import com.kuranas.mobile.i18n.TranslationManager;
import com.kuranas.mobile.infra.http.HttpClient;

public final class ServiceLocator {

    private static ServiceLocator instance;

    private final HttpClient httpClient;
    private final TranslationManager translationManager;
    private final NotificationApi notificationApi;
    private final EmailApi emailApi;

    private ServiceLocator(String baseUrl) {
        httpClient = new HttpClient(baseUrl);

        TranslationManager tm = TranslationManager.getInstance();
        tm.setHttpClient(httpClient);
        translationManager = tm;

        notificationApi = new NotificationApi(httpClient);
        emailApi = new EmailApi(httpClient);
    }

    public static synchronized void initialize(String baseUrl) {
        if (instance != null) {
            instance.shutdown();
        }
        instance = new ServiceLocator(baseUrl);
    }

    public static synchronized ServiceLocator getInstance() {
        if (instance == null) {
            throw new IllegalStateException("ServiceLocator not initialized. Call initialize() first.");
        }
        return instance;
    }

    public HttpClient getHttpClient() {
        return httpClient;
    }

    public TranslationManager getTranslationManager() {
        return translationManager;
    }

    public NotificationApi getNotificationApi() {
        return notificationApi;
    }

    public EmailApi getEmailApi() {
        return emailApi;
    }

    public void shutdown() {
        httpClient.shutdown();
    }
}
