package com.kuranas.mobile.data.remote.api;

import com.kuranas.mobile.data.mapper.VideoMapper;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.VideoItem;
import com.kuranas.mobile.domain.model.VideoPlaybackState;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.infra.http.HttpClient;
import com.kuranas.mobile.infra.http.HttpResponse;

import org.json.JSONException;
import org.json.JSONObject;

public final class VideoApi {

    private final HttpClient httpClient;

    public VideoApi(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    public void getLibrary(int page, int pageSize, final ApiCallback<PaginatedResult<VideoItem>> callback) {
        String path = "/api/v1/video/library/files?page=" + page + "&page_size=" + pageSize;
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
                    PaginatedResult<VideoItem> result = VideoMapper.fromPaginatedJson(json);
                    callback.onSuccess(result);
                } catch (JSONException e) {
                    callback.onError(AppError.invalidPayload(e));
                }
            }
        });
    }

    public void startPlayback(int videoId, final ApiCallback<VideoPlaybackState> callback) {
        String path = "/api/v1/video/playback/start";
        try {
            JSONObject body = new JSONObject();
            body.put("video_id", videoId);

            httpClient.post(path, body.toString(), new HttpClient.Callback() {
                @Override
                public void onResponse(HttpResponse response) {
                    handlePlaybackStateResponse(response, callback);
                }
            });
        } catch (JSONException e) {
            callback.onError(AppError.invalidPayload(e));
        }
    }

    public void getPlaybackState(final ApiCallback<VideoPlaybackState> callback) {
        String path = "/api/v1/video/playback/state";
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                handlePlaybackStateDirectResponse(response, callback);
            }
        });
    }

    public void updatePlaybackState(int videoId, double currentTime, double duration,
                                    boolean isPaused, boolean completed,
                                    final ApiCallback<VideoPlaybackState> callback) {
        String path = "/api/v1/video/playback/state";
        try {
            JSONObject body = new JSONObject();
            body.put("video_id", videoId);
            body.put("current_time", currentTime);
            body.put("duration", duration);
            body.put("is_paused", isPaused);
            body.put("completed", completed);

            httpClient.put(path, body.toString(), new HttpClient.Callback() {
                @Override
                public void onResponse(HttpResponse response) {
                    handlePlaybackStateDirectResponse(response, callback);
                }
            });
        } catch (JSONException e) {
            callback.onError(AppError.invalidPayload(e));
        }
    }

    private void handlePlaybackStateResponse(HttpResponse response, ApiCallback<VideoPlaybackState> callback) {
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
            JSONObject stateJson = json.optJSONObject("playback_state");
            if (stateJson == null) {
                stateJson = json;
            }
            VideoPlaybackState state = VideoMapper.playbackStateFromJson(stateJson);
            callback.onSuccess(state);
        } catch (JSONException e) {
            callback.onError(AppError.invalidPayload(e));
        }
    }

    private void handlePlaybackStateDirectResponse(HttpResponse response, ApiCallback<VideoPlaybackState> callback) {
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
            VideoPlaybackState state = VideoMapper.playbackStateFromJson(json);
            callback.onSuccess(state);
        } catch (JSONException e) {
            callback.onError(AppError.invalidPayload(e));
        }
    }
}
