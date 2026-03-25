package com.kuranas.mobile.domain.repository;

import com.kuranas.mobile.domain.model.MusicPlayerState;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.Track;
import com.kuranas.mobile.infra.http.ApiCallback;

public interface MusicRepository {

    void getLibraryTracks(int page, int pageSize, ApiCallback<PaginatedResult<Track>> callback);

    void getPlaylistTracks(int playlistId, int page, int pageSize, ApiCallback<PaginatedResult<Track>> callback);

    void getPlayerState(ApiCallback<MusicPlayerState> callback);

    void updatePlayerState(int playlistId, int fileId, double position, ApiCallback<MusicPlayerState> callback);
}
