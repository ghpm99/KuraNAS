package com.kuranas.mobile.infra.http;

import com.kuranas.mobile.domain.error.AppError;

public interface ApiCallback<T> {

    void onSuccess(T result);

    void onError(AppError error);
}
