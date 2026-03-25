package com.kuranas.mobile.data.remote.api;

import com.kuranas.mobile.data.mapper.FileMapper;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.infra.http.HttpClient;
import com.kuranas.mobile.infra.http.HttpResponse;

import org.json.JSONException;
import org.json.JSONObject;

public final class FileApi {

    private final HttpClient httpClient;

    public FileApi(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    public void getTree(int page, int pageSize, final ApiCallback<PaginatedResult<FileItem>> callback) {
        String path = "/api/v1/files/tree?page=" + page + "&page_size=" + pageSize;
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                handlePaginatedResponse(response, callback);
            }
        });
    }

    public void getByPath(String filePath, int page, int pageSize, final ApiCallback<PaginatedResult<FileItem>> callback) {
        String encodedPath;
        try {
            encodedPath = java.net.URLEncoder.encode(filePath, "UTF-8");
        } catch (java.io.UnsupportedEncodingException e) {
            encodedPath = filePath;
        }
        String path = "/api/v1/files/path?path=" + encodedPath + "&page=" + page + "&page_size=" + pageSize;
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                handlePaginatedResponse(response, callback);
            }
        });
    }

    public void getChildren(int parentId, int page, int pageSize, final ApiCallback<PaginatedResult<FileItem>> callback) {
        String path = "/api/v1/files/" + parentId + "?page=" + page + "&page_size=" + pageSize;
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                handlePaginatedResponse(response, callback);
            }
        });
    }

    public void getImages(int page, int pageSize, String groupBy, final ApiCallback<PaginatedResult<FileItem>> callback) {
        String path = "/api/v1/files/images?page=" + page + "&page_size=" + pageSize;
        if (groupBy != null && !groupBy.isEmpty()) {
            path = path + "&group_by=" + groupBy;
        }
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                handlePaginatedResponse(response, callback);
            }
        });
    }

    public void getStarred(int page, int pageSize, final ApiCallback<PaginatedResult<FileItem>> callback) {
        String path = "/api/v1/files/tree?category=starred&page=" + page + "&page_size=" + pageSize;
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                handlePaginatedResponse(response, callback);
            }
        });
    }

    private void handlePaginatedResponse(HttpResponse response, ApiCallback<PaginatedResult<FileItem>> callback) {
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
            PaginatedResult<FileItem> result = FileMapper.fromPaginatedJson(json);
            callback.onSuccess(result);
        } catch (JSONException e) {
            callback.onError(AppError.invalidPayload(e));
        }
    }
}
