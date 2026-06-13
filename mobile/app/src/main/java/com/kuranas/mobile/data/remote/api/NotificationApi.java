package com.kuranas.mobile.data.remote.api;

import com.kuranas.mobile.data.mapper.NotificationMapper;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.NotificationItem;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.infra.http.HttpClient;
import com.kuranas.mobile.infra.http.HttpResponse;

import org.json.JSONException;

import java.util.List;

/**
 * Reads the most recent notifications for the kiosk panel. Mirrors the removed
 * media APIs' shape (task 17 history): a thin wrapper translating an HTTP
 * response into a domain list or an {@link AppError}.
 */
public final class NotificationApi {

    private final HttpClient httpClient;

    public NotificationApi(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    public void getRecent(int pageSize, final ApiCallback<List<NotificationItem>> callback) {
        String path = "/api/v1/notifications?page=1&page_size=" + pageSize;
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                if (response.getError() != null) {
                    callback.onError(AppError.networkUnavailable(response.getError()));
                    return;
                }
                if (!response.isSuccessful()) {
                    callback.onError(AppError.fromHttpResponse(response.getStatusCode(), null));
                    return;
                }
                try {
                    List<NotificationItem> items = NotificationMapper.fromPaginatedJson(response.toJsonObject());
                    callback.onSuccess(items);
                } catch (JSONException e) {
                    callback.onError(AppError.invalidPayload(e));
                }
            }
        });
    }
}
