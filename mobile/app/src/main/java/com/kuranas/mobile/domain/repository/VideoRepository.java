package com.kuranas.mobile.domain.repository;

import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.VideoItem;
import com.kuranas.mobile.domain.model.VideoPlaybackState;
import com.kuranas.mobile.infra.http.ApiCallback;

public interface VideoRepository {

    void getLibraryVideos(int page, int pageSize, ApiCallback<PaginatedResult<VideoItem>> callback);

    void startPlayback(int videoId, ApiCallback<VideoPlaybackState> callback);

    void getPlaybackState(ApiCallback<VideoPlaybackState> callback);

    void updatePlaybackState(int videoId, double currentTime, double duration,
                             boolean isPaused, boolean completed,
                             ApiCallback<VideoPlaybackState> callback);
}
