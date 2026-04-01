package com.kuranas.mobile.presentation.music;

import java.io.IOException;

public final class MusicPlayerController {

    static final long UPDATE_INTERVAL_MS = 1000;

    private final ViewContract viewContract;
    private final AudioEngine audioEngine;
    private final PlayerStateRepository stateRepository;
    private final Scheduler scheduler;

    private PlayerState playerState;
    private final Runnable updateRunnable = new Runnable() {
        @Override
        public void run() {
            updateSeekBar();
            scheduler.postDelayed(this, UPDATE_INTERVAL_MS);
        }
    };

    public MusicPlayerController(
            ViewContract viewContract,
            AudioEngine audioEngine,
            PlayerStateRepository stateRepository,
            Scheduler scheduler
    ) {
        this.viewContract = viewContract;
        this.audioEngine = audioEngine;
        this.stateRepository = stateRepository;
        this.scheduler = scheduler;
        this.playerState = PlayerState.empty();
    }

    public void start(String baseUrl, PlayerState state) {
        playerState = state == null ? PlayerState.empty() : state;

        viewContract.setTrackTitle(playerState.getTitle());
        viewContract.setTrackArtist(playerState.getArtist());
        viewContract.setDefaultAlbumArt();

        String streamUrl = buildStreamUrl(baseUrl, playerState.getFileId());
        audioEngine.setAudioStreamTypeMusic();

        try {
            audioEngine.setDataSource(streamUrl);
        } catch (IOException error) {
            return;
        }

        audioEngine.setOnPrepared(new Runnable() {
            @Override
            public void run() {
                playerState.setPrepared(true);
                int duration = audioEngine.getDuration();
                viewContract.setSeekMax(duration);
                viewContract.setTotalTimeText(formatTime(duration));
                audioEngine.start();
                viewContract.setPlayPausePlaying(true);
                startSeekBarUpdates();
            }
        });

        audioEngine.setOnCompletion(new Runnable() {
            @Override
            public void run() {
                viewContract.setPlayPausePlaying(false);
                stopSeekBarUpdates();
                savePlayerState();
            }
        });

        audioEngine.setOnError(new AudioEngine.ErrorCallback() {
            @Override
            public boolean onError(int what, int extra) {
                playerState.setPrepared(false);
                return true;
            }
        });

        audioEngine.prepareAsync();
    }

    public void onPlayPauseClicked() {
        if (!canControlPlayback()) {
            return;
        }

        if (audioEngine.isPlaying()) {
            audioEngine.pause();
            viewContract.setPlayPausePlaying(false);
            stopSeekBarUpdates();
            savePlayerState();
            return;
        }

        audioEngine.start();
        viewContract.setPlayPausePlaying(true);
        startSeekBarUpdates();
    }

    public void onPreviousClicked() {
        if (!canControlPlayback()) {
            return;
        }
        audioEngine.seekTo(0);
        updateSeekBar();
    }

    public void onSeekProgressChanged(int progress, boolean fromUser) {
        if (!fromUser || !canControlPlayback()) {
            return;
        }
        viewContract.setCurrentTimeText(formatTime(progress));
    }

    public void onSeekStartTracking() {
        playerState.setUserSeeking(true);
    }

    public void onSeekStopTracking(int progress) {
        playerState.setUserSeeking(false);
        if (!canControlPlayback()) {
            return;
        }
        audioEngine.seekTo(progress);
    }

    public void onPause() {
        if (canControlPlayback() && audioEngine.isPlaying()) {
            audioEngine.pause();
            viewContract.setPlayPausePlaying(false);
            stopSeekBarUpdates();
            savePlayerState();
        }
    }

    public void onResume() {
        if (canControlPlayback()) {
            audioEngine.start();
            viewContract.setPlayPausePlaying(true);
            startSeekBarUpdates();
        }
    }

    public void onDestroy() {
        stopSeekBarUpdates();
        if (playerState.isPrepared()) {
            savePlayerState();
        }
        audioEngine.release();
        playerState.setPrepared(false);
        scheduler.clear();
    }

    String formatTime(int millis) {
        int safeMillis = millis < 0 ? 0 : millis;
        int totalSeconds = safeMillis / 1000;
        int minutes = totalSeconds / 60;
        int seconds = totalSeconds % 60;
        return String.format("%d:%02d", minutes, seconds);
    }

    private String buildStreamUrl(String baseUrl, int fileId) {
        String safeBaseUrl = baseUrl == null ? "" : baseUrl;
        return safeBaseUrl + "/api/v1/files/stream/" + fileId;
    }

    private void startSeekBarUpdates() {
        scheduler.post(updateRunnable);
    }

    private void stopSeekBarUpdates() {
        scheduler.removeCallbacks(updateRunnable);
    }

    private void updateSeekBar() {
        if (canControlPlayback() && !playerState.isUserSeeking()) {
            int position = audioEngine.getCurrentPosition();
            viewContract.setSeekProgress(position);
            viewContract.setCurrentTimeText(formatTime(position));
        }
    }

    private void savePlayerState() {
        if (!canControlPlayback()) {
            return;
        }
        double positionSeconds = audioEngine.getCurrentPosition() / 1000.0;
        stateRepository.updatePlayerState(0, playerState.getFileId(), positionSeconds);
    }

    private boolean canControlPlayback() {
        return playerState.isPrepared();
    }

    public interface ViewContract {
        void setTrackTitle(String title);
        void setTrackArtist(String artist);
        void setDefaultAlbumArt();
        void setSeekMax(int max);
        void setSeekProgress(int progress);
        void setCurrentTimeText(String text);
        void setTotalTimeText(String text);
        void setPlayPausePlaying(boolean playing);
    }

    public interface AudioEngine {
        void setAudioStreamTypeMusic();
        void setDataSource(String streamUrl) throws IOException;
        void setOnPrepared(Runnable runnable);
        void setOnCompletion(Runnable runnable);
        void setOnError(ErrorCallback callback);
        void prepareAsync();
        boolean isPlaying();
        void start();
        void pause();
        void seekTo(int positionMs);
        int getDuration();
        int getCurrentPosition();
        void release();

        interface ErrorCallback {
            boolean onError(int what, int extra);
        }
    }

    public interface PlayerStateRepository {
        void updatePlayerState(int playlistId, int fileId, double positionSeconds);
    }

    public interface Scheduler {
        void post(Runnable runnable);
        void postDelayed(Runnable runnable, long delayMs);
        void removeCallbacks(Runnable runnable);
        void clear();
    }

    public static final class PlayerState {
        private final int fileId;
        private final String title;
        private final String artist;
        private boolean prepared;
        private boolean userSeeking;

        public PlayerState(int fileId, String title, String artist) {
            this.fileId = fileId;
            this.title = title == null ? "" : title;
            this.artist = artist == null ? "" : artist;
        }

        public static PlayerState empty() {
            return new PlayerState(0, "", "");
        }

        public int getFileId() {
            return fileId;
        }

        public String getTitle() {
            return title;
        }

        public String getArtist() {
            return artist;
        }

        public boolean isPrepared() {
            return prepared;
        }

        public void setPrepared(boolean prepared) {
            this.prepared = prepared;
        }

        public boolean isUserSeeking() {
            return userSeeking;
        }

        public void setUserSeeking(boolean userSeeking) {
            this.userSeeking = userSeeking;
        }
    }
}
