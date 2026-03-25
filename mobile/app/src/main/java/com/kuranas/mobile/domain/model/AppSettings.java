package com.kuranas.mobile.domain.model;

import java.util.List;

public final class AppSettings {

    private final boolean rememberMusicQueue;
    private final boolean rememberVideoProgress;
    private final boolean autoplayNextVideo;
    private final int imageSlideshowSeconds;
    private final String currentLanguage;
    private final List<String> availableLanguages;

    public AppSettings(boolean rememberMusicQueue, boolean rememberVideoProgress,
                       boolean autoplayNextVideo, int imageSlideshowSeconds,
                       String currentLanguage, List<String> availableLanguages) {
        this.rememberMusicQueue = rememberMusicQueue;
        this.rememberVideoProgress = rememberVideoProgress;
        this.autoplayNextVideo = autoplayNextVideo;
        this.imageSlideshowSeconds = imageSlideshowSeconds;
        this.currentLanguage = currentLanguage;
        this.availableLanguages = availableLanguages;
    }

    public boolean isRememberMusicQueue() { return rememberMusicQueue; }
    public boolean isRememberVideoProgress() { return rememberVideoProgress; }
    public boolean isAutoplayNextVideo() { return autoplayNextVideo; }
    public int getImageSlideshowSeconds() { return imageSlideshowSeconds; }
    public String getCurrentLanguage() { return currentLanguage; }
    public List<String> getAvailableLanguages() { return availableLanguages; }
}
