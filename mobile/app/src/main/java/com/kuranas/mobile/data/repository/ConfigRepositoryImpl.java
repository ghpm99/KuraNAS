package com.kuranas.mobile.data.repository;

import com.kuranas.mobile.data.remote.api.ConfigApi;
import com.kuranas.mobile.domain.model.AppSettings;
import com.kuranas.mobile.domain.repository.ConfigRepository;
import com.kuranas.mobile.infra.http.ApiCallback;

public final class ConfigRepositoryImpl implements ConfigRepository {

    private final ConfigApi configApi;

    public ConfigRepositoryImpl(ConfigApi configApi) {
        this.configApi = configApi;
    }

    @Override
    public void getSettings(ApiCallback<AppSettings> callback) {
        configApi.getSettings(callback);
    }

    @Override
    public void updateLanguage(String language, ApiCallback<AppSettings> callback) {
        configApi.updateLanguage(language, callback);
    }
}
