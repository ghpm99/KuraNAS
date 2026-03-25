package com.kuranas.mobile.domain.repository;

import com.kuranas.mobile.domain.model.AppSettings;
import com.kuranas.mobile.infra.http.ApiCallback;

public interface ConfigRepository {

    void getSettings(ApiCallback<AppSettings> callback);

    void updateLanguage(String language, ApiCallback<AppSettings> callback);
}
