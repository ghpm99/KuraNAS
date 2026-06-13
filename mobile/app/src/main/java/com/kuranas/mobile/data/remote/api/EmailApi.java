package com.kuranas.mobile.data.remote.api;

import com.kuranas.mobile.data.mapper.EmailMapper;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.EmailItem;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.infra.http.HttpClient;
import com.kuranas.mobile.infra.http.HttpResponse;

import org.json.JSONException;

import java.util.List;

/**
 * Reads the most recent synced e-mails for the kiosk panel. When the e-mail
 * feature is disabled on the server (no {@code EMAIL_TOKEN_KEY}) the endpoint
 * answers 503; that surfaces as an {@link AppError} and the panel shows its
 * offline indicator — no special-casing here.
 */
public final class EmailApi {

    private final HttpClient httpClient;

    public EmailApi(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    public void getRecent(int pageSize, final ApiCallback<List<EmailItem>> callback) {
        String path = "/api/v1/email/messages?page=1&page_size=" + pageSize;
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
                    List<EmailItem> items = EmailMapper.fromPaginatedJson(response.toJsonObject());
                    callback.onSuccess(items);
                } catch (JSONException e) {
                    callback.onError(AppError.invalidPayload(e));
                }
            }
        });
    }
}
