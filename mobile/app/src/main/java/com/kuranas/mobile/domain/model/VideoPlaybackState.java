package com.kuranas.mobile.domain.model;

public final class VideoPlaybackState {

    private final int id;
    private final String clientId;
    private final int playlistId;
    private final int videoId;
    private final double currentTime;
    private final double duration;
    private final boolean paused;
    private final boolean completed;
    private final String lastUpdate;

    public VideoPlaybackState(int id, String clientId, int playlistId, int videoId,
                              double currentTime, double duration, boolean paused,
                              boolean completed, String lastUpdate) {
        this.id = id;
        this.clientId = clientId;
        this.playlistId = playlistId;
        this.videoId = videoId;
        this.currentTime = currentTime;
        this.duration = duration;
        this.paused = paused;
        this.completed = completed;
        this.lastUpdate = lastUpdate;
    }

    public int getId() { return id; }
    public String getClientId() { return clientId; }
    public int getPlaylistId() { return playlistId; }
    public int getVideoId() { return videoId; }
    public double getCurrentTime() { return currentTime; }
    public double getDuration() { return duration; }
    public boolean isPaused() { return paused; }
    public boolean isCompleted() { return completed; }
    public String getLastUpdate() { return lastUpdate; }
}
