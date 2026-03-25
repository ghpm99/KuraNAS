package com.kuranas.mobile.data.remote.api;

import com.kuranas.mobile.data.mapper.MusicMapper;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.MusicPlayerState;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.Track;
import com.kuranas.mobile.infra.http.ApiCallback;
import com.kuranas.mobile.infra.http.HttpClient;
import com.kuranas.mobile.infra.http.HttpResponse;

import org.json.JSONException;
import org.json.JSONObject;

public final class MusicApi {

    private final HttpClient httpClient;

    public MusicApi(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    public void getLibrary(int page, int pageSize, final ApiCallback<PaginatedResult<Track>> callback) {
        String path = "/api/v1/music/library?page=" + page + "&page_size=" + pageSize;
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                handleTrackListResponse(response, callback);
            }
        });
    }

    public void getPlaylistTracks(int playlistId, int page, int pageSize, final ApiCallback<PaginatedResult<Track>> callback) {
        String path = "/api/v1/music/playlists/" + playlistId + "/tracks?page=" + page + "&page_size=" + pageSize;
        httpClient.get(path, new HttpClient.Callback() {
            @Override
            public void onResponse(HttpResponse response) {
                handleTrackListResponse(response, callback);
            }
        });
    }

    public void getPlayerState(final ApiCallback<MusicPlayerState> callback) {
        String path = "/api/v1/music/player-state/";
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
                    MusicPlayerState state = MusicMapper.playerStateFromJson(json);
                    callback.onSuccess(state);
                } catch (JSONException e) {
                    callback.onError(AppError.invalidPayload(e));
                }
            }
        });
    }

    public void updatePlayerState(int playlistId, int fileId, double position, final ApiCallback<MusicPlayerState> callback) {
        String path = "/api/v1/music/player-state/";
        try {
            JSONObject body = new JSONObject();
            body.put("playlist_id", playlistId);
            body.put("current_file_id", fileId);
            body.put("current_position", position);

            httpClient.put(path, body.toString(), new HttpClient.Callback() {
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
                        MusicPlayerState state = MusicMapper.playerStateFromJson(json);
                        callback.onSuccess(state);
                    } catch (JSONException e) {
                        callback.onError(AppError.invalidPayload(e));
                    }
                }
            });
        } catch (JSONException e) {
            callback.onError(AppError.invalidPayload(e));
        }
    }

    private void handleTrackListResponse(HttpResponse response, ApiCallback<PaginatedResult<Track>> callback) {
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
            PaginatedResult<Track> result = MusicMapper.trackListFromJson(json);
            callback.onSuccess(result);
        } catch (JSONException e) {
            callback.onError(AppError.invalidPayload(e));
        }
    }
}
