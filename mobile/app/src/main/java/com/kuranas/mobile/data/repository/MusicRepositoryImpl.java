package com.kuranas.mobile.data.repository;

import com.kuranas.mobile.data.remote.api.MusicApi;
import com.kuranas.mobile.domain.model.MusicPlayerState;
import com.kuranas.mobile.domain.model.PaginatedResult;
import com.kuranas.mobile.domain.model.Track;
import com.kuranas.mobile.domain.repository.MusicRepository;
import com.kuranas.mobile.infra.http.ApiCallback;

public final class MusicRepositoryImpl implements MusicRepository {

    private final MusicApi musicApi;

    public MusicRepositoryImpl(MusicApi musicApi) {
        this.musicApi = musicApi;
    }

    @Override
    public void getLibraryTracks(int page, int pageSize, ApiCallback<PaginatedResult<Track>> callback) {
        musicApi.getLibrary(page, pageSize, callback);
    }

    @Override
    public void getPlaylistTracks(int playlistId, int page, int pageSize, ApiCallback<PaginatedResult<Track>> callback) {
        musicApi.getPlaylistTracks(playlistId, page, pageSize, callback);
    }

    @Override
    public void getPlayerState(ApiCallback<MusicPlayerState> callback) {
        musicApi.getPlayerState(callback);
    }

    @Override
    public void updatePlayerState(int playlistId, int fileId, double position, ApiCallback<MusicPlayerState> callback) {
        musicApi.updatePlayerState(playlistId, fileId, position, callback);
    }
}
