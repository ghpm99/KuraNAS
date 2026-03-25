package com.kuranas.mobile.data.remote.api;

import com.kuranas.mobile.data.mapper.SearchMapper;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.SearchResult;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.infra.http.HttpClient;
import com.kuranas.mobile.infra.http.HttpResponse;

import org.json.JSONException;
import org.json.JSONObject;

public final class SearchApi {

    private final HttpClient httpClient;

    public SearchApi(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    public void searchGlobal(String query, int limit, final ApiCallback<SearchResult> callback) {
        String encodedQuery;
        try {
            encodedQuery = java.net.URLEncoder.encode(query, "UTF-8");
        } catch (java.io.UnsupportedEncodingException e) {
            encodedQuery = query;
        }
        String path = "/api/v1/search/global?q=" + encodedQuery + "&limit=" + limit;
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
                    JSONObject json = response.toJsonObject();
                    SearchResult result = SearchMapper.fromJson(json);
                    callback.onSuccess(result);
                } catch (JSONException e) {
                    callback.onError(AppError.invalidPayload(e));
                }
            }
        });
    }
}
