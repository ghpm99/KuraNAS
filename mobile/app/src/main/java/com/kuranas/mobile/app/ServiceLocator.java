package com.kuranas.mobile.app;

import com.kuranas.mobile.BuildConfig;
import com.kuranas.mobile.data.mapper.FileMapper;
import com.kuranas.mobile.data.mapper.MusicMapper;
import com.kuranas.mobile.data.mapper.SearchMapper;
import com.kuranas.mobile.data.mapper.SettingsMapper;
import com.kuranas.mobile.data.mapper.VideoMapper;
import com.kuranas.mobile.data.remote.api.ConfigApi;
import com.kuranas.mobile.data.remote.api.FileApi;
import com.kuranas.mobile.data.remote.api.MusicApi;
import com.kuranas.mobile.data.remote.api.SearchApi;
import com.kuranas.mobile.data.remote.api.VideoApi;
import com.kuranas.mobile.data.repository.ConfigRepositoryImpl;
import com.kuranas.mobile.data.repository.FileRepositoryImpl;
import com.kuranas.mobile.data.repository.MusicRepositoryImpl;
import com.kuranas.mobile.data.repository.SearchRepositoryImpl;
import com.kuranas.mobile.data.repository.VideoRepositoryImpl;
import com.kuranas.mobile.domain.repository.ConfigRepository;
import com.kuranas.mobile.domain.repository.FileRepository;
import com.kuranas.mobile.domain.repository.MusicRepository;
import com.kuranas.mobile.domain.repository.SearchRepository;
import com.kuranas.mobile.domain.repository.VideoRepository;
import com.kuranas.mobile.i18n.TranslationManager;
import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.infra.http.HttpClient;

public final class ServiceLocator {

    private static ServiceLocator instance;

    private final HttpClient httpClient;
    private final TranslationManager translationManager;
    private final BitmapCache bitmapCache;

    private final FileApi fileApi;
    private final MusicApi musicApi;
    private final VideoApi videoApi;
    private final SearchApi searchApi;
    private final ConfigApi configApi;

    private final FileRepository fileRepository;
    private final MusicRepository musicRepository;
    private final VideoRepository videoRepository;
    private final SearchRepository searchRepository;
    private final ConfigRepository configRepository;

    private ServiceLocator() {
        httpClient = new HttpClient(BuildConfig.API_BASE_URL);
        translationManager = new TranslationManager(httpClient);
        bitmapCache = new BitmapCache();

        fileApi = new FileApi(httpClient);
        musicApi = new MusicApi(httpClient);
        videoApi = new VideoApi(httpClient);
        searchApi = new SearchApi(httpClient);
        configApi = new ConfigApi(httpClient);

        fileRepository = new FileRepositoryImpl(fileApi);
        musicRepository = new MusicRepositoryImpl(musicApi);
        videoRepository = new VideoRepositoryImpl(videoApi);
        searchRepository = new SearchRepositoryImpl(searchApi);
        configRepository = new ConfigRepositoryImpl(configApi);
    }

    public static synchronized ServiceLocator getInstance() {
        if (instance == null) {
            instance = new ServiceLocator();
        }
        return instance;
    }

    public HttpClient getHttpClient() {
        return httpClient;
    }

    public TranslationManager getTranslationManager() {
        return translationManager;
    }

    public BitmapCache getBitmapCache() {
        return bitmapCache;
    }

    public FileRepository getFileRepository() {
        return fileRepository;
    }

    public MusicRepository getMusicRepository() {
        return musicRepository;
    }

    public VideoRepository getVideoRepository() {
        return videoRepository;
    }

    public SearchRepository getSearchRepository() {
        return searchRepository;
    }

    public ConfigRepository getConfigRepository() {
        return configRepository;
    }

    public void shutdown() {
        httpClient.shutdown();
        bitmapCache.clear();
    }
}
