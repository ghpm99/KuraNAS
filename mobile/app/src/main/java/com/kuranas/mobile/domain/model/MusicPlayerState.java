package com.kuranas.mobile.domain.model;

public final class MusicPlayerState {

    private final int id;
    private final String clientId;
    private final int playlistId;
    private final int currentFileId;
    private final double currentPosition;
    private final double volume;
    private final boolean shuffle;
    private final String repeatMode;
    private final String updatedAt;

    public MusicPlayerState(int id, String clientId, int playlistId, int currentFileId,
                            double currentPosition, double volume, boolean shuffle,
                            String repeatMode, String updatedAt) {
        this.id = id;
        this.clientId = clientId;
        this.playlistId = playlistId;
        this.currentFileId = currentFileId;
        this.currentPosition = currentPosition;
        this.volume = volume;
        this.shuffle = shuffle;
        this.repeatMode = repeatMode;
        this.updatedAt = updatedAt;
    }

    public int getId() { return id; }
    public String getClientId() { return clientId; }
    public int getPlaylistId() { return playlistId; }
    public int getCurrentFileId() { return currentFileId; }
    public double getCurrentPosition() { return currentPosition; }
    public double getVolume() { return volume; }
    public boolean isShuffle() { return shuffle; }
    public String getRepeatMode() { return repeatMode; }
    public String getUpdatedAt() { return updatedAt; }
}
