package com.kuranas.mobile.data.repository;

import com.kuranas.mobile.data.remote.api.VideoApi;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.VideoItem;
import com.kuranas.mobile.domain.model.VideoPlaybackState;
import com.kuranas.mobile.domain.repository.VideoRepository;
import com.kuranas.mobile.infra.http.ApiCallback;

public final class VideoRepositoryImpl implements VideoRepository {

    private final VideoApi videoApi;

    public VideoRepositoryImpl(VideoApi videoApi) {
        this.videoApi = videoApi;
    }

    @Override
    public void getLibraryVideos(int page, int pageSize, ApiCallback<PaginatedResult<VideoItem>> callback) {
        videoApi.getLibrary(page, pageSize, callback);
    }

    @Override
    public void startPlayback(int videoId, ApiCallback<VideoPlaybackState> callback) {
        videoApi.startPlayback(videoId, callback);
    }

    @Override
    public void getPlaybackState(ApiCallback<VideoPlaybackState> callback) {
        videoApi.getPlaybackState(callback);
    }

    @Override
    public void updatePlaybackState(int videoId, double currentTime, double duration,
                                    boolean isPaused, boolean completed,
                                    ApiCallback<VideoPlaybackState> callback) {
        videoApi.updatePlaybackState(videoId, currentTime, duration, isPaused, completed, callback);
    }
}
