package com.kuranas.mobile.presentation.video;

import com.kuranas.mobile.domain.model.VideoPlaybackState;

public final class VideoPlayerController {

    static final long CONTROLS_HIDE_DELAY_MS = 3000;
    static final long STATE_UPDATE_INTERVAL_MS = 5000;
    static final long SEEK_UPDATE_INTERVAL_MS = 1000;

    private final PlaybackStateRepository playbackStateRepository;
    private final PlaybackEngine playbackEngine;
    private final ViewContract viewContract;
    private final Scheduler scheduler;

    private int videoId;
    private String streamUrl;
    private boolean isPlaying;

    private final Runnable hideControlsRunnable = new Runnable() {
        @Override
        public void run() {
            viewContract.setControlsVisible(false);
        }
    };

    private final Runnable updateStateRunnable = new Runnable() {
        @Override
        public void run() {
            savePlaybackState(false);
            scheduler.postDelayed(this, STATE_UPDATE_INTERVAL_MS);
        }
    };

    private final Runnable updateSeekRunnable = new Runnable() {
        @Override
        public void run() {
            if (playbackEngine.isPlaying()) {
                int position = playbackEngine.getCurrentPosition();
                viewContract.setSeekProgress(position);
                updateTimeDisplay();
            }
            scheduler.postDelayed(this, SEEK_UPDATE_INTERVAL_MS);
        }
    };

    public VideoPlayerController(
            PlaybackStateRepository playbackStateRepository,
            PlaybackEngine playbackEngine,
            ViewContract viewContract,
            Scheduler scheduler
    ) {
        this.playbackStateRepository = playbackStateRepository;
        this.playbackEngine = playbackEngine;
        this.viewContract = viewContract;
        this.scheduler = scheduler;
        this.streamUrl = "";
    }

    public void start(int videoId, String streamUrl) {
        this.videoId = videoId;
        this.streamUrl = streamUrl == null ? "" : streamUrl;

        playbackStateRepository.startPlayback(videoId, new PlaybackStateRepository.StartPlaybackCallback() {
            @Override
            public void onSuccess(VideoPlaybackState state) {
                prepareAndPlay(state);
            }

            @Override
            public void onError() {
                prepareAndPlay(null);
            }
        });
    }

    public void onVideoTapped() {
        if (viewContract.isControlsVisible()) {
            viewContract.setControlsVisible(false);
            return;
        }
        viewContract.setControlsVisible(true);
        scheduleHideControls();
    }

    public void onPlayPauseClicked() {
        if (playbackEngine.isPlaying()) {
            playbackEngine.pause();
            isPlaying = false;
            viewContract.setPlayPausePlaying(false);
            stopSeekBarUpdates();
            savePlaybackState(false);
            return;
        }

        playbackEngine.start();
        isPlaying = true;
        viewContract.setPlayPausePlaying(true);
        startSeekBarUpdates();
        scheduleHideControls();
    }

    public void onSeekProgressChanged(int progress, boolean fromUser) {
        if (!fromUser) {
            return;
        }
        playbackEngine.seekTo(progress);
        updateTimeDisplay();
    }

    public void onSeekStartTracking() {
        cancelHideControls();
    }

    public void onSeekStopTracking() {
        scheduleHideControls();
    }

    public void onPause() {
        if (playbackEngine.isPlaying()) {
            playbackEngine.pause();
            isPlaying = false;
            savePlaybackState(false);
        }
        stopSeekBarUpdates();
        stopStateUpdates();
        cancelHideControls();
    }

    public void onResume() {
        if (isPlaying) {
            playbackEngine.start();
            startSeekBarUpdates();
            startStateUpdates();
        }
    }

    public void onDestroy() {
        scheduler.clear();
        playbackEngine.stop();
    }

    String formatTime(int millis) {
        int safeMillis = millis < 0 ? 0 : millis;
        int totalSeconds = safeMillis / 1000;
        int hours = totalSeconds / 3600;
        int minutes = (totalSeconds % 3600) / 60;
        int seconds = totalSeconds % 60;

        if (hours > 0) {
            return String.format("%d:%02d:%02d", hours, minutes, seconds);
        }
        return String.format("%d:%02d", minutes, seconds);
    }

    private void prepareAndPlay(final VideoPlaybackState state) {
        playbackEngine.setVideoUri(streamUrl);
        playbackEngine.setOnPrepared(new Runnable() {
            @Override
            public void run() {
                int duration = playbackEngine.getDuration();
                viewContract.setSeekMax(duration);

                if (state != null && state.getCurrentTime() > 0 && !state.isCompleted()) {
                    int seekPosition = (int) (state.getCurrentTime() * 1000);
                    playbackEngine.seekTo(seekPosition);
                }

                playbackEngine.start();
                isPlaying = true;
                viewContract.setPlayPausePlaying(true);
                updateTimeDisplay();
                scheduleHideControls();
                startSeekBarUpdates();
                startStateUpdates();
            }
        });
        playbackEngine.setOnCompletion(new Runnable() {
            @Override
            public void run() {
                isPlaying = false;
                viewContract.setPlayPausePlaying(false);
                viewContract.setControlsVisible(true);
                stopSeekBarUpdates();
                stopStateUpdates();
                savePlaybackState(true);
            }
        });
    }

    private void updateTimeDisplay() {
        int current = playbackEngine.getCurrentPosition();
        int duration = playbackEngine.getDuration();
        String timeText = formatTime(current) + " / " + formatTime(duration);
        viewContract.setTimeText(timeText);
    }

    private void scheduleHideControls() {
        cancelHideControls();
        scheduler.postDelayed(hideControlsRunnable, CONTROLS_HIDE_DELAY_MS);
    }

    private void cancelHideControls() {
        scheduler.removeCallbacks(hideControlsRunnable);
    }

    private void startSeekBarUpdates() {
        scheduler.post(updateSeekRunnable);
    }

    private void stopSeekBarUpdates() {
        scheduler.removeCallbacks(updateSeekRunnable);
    }

    private void startStateUpdates() {
        scheduler.postDelayed(updateStateRunnable, STATE_UPDATE_INTERVAL_MS);
    }

    private void stopStateUpdates() {
        scheduler.removeCallbacks(updateStateRunnable);
    }

    private void savePlaybackState(boolean completed) {
        double currentTimeSec = playbackEngine.getCurrentPosition() / 1000.0;
        double durationSec = playbackEngine.getDuration() / 1000.0;
        boolean paused = !playbackEngine.isPlaying();

        playbackStateRepository.updatePlaybackState(
                videoId,
                currentTimeSec,
                durationSec,
                paused,
                completed
        );
    }

    public interface ViewContract {
        void setPlayPausePlaying(boolean playing);
        void setControlsVisible(boolean visible);
        boolean isControlsVisible();
        void setSeekMax(int max);
        void setSeekProgress(int progress);
        void setTimeText(String text);
    }

    public interface PlaybackEngine {
        void setVideoUri(String streamUrl);
        void setOnPrepared(Runnable runnable);
        void setOnCompletion(Runnable runnable);
        int getDuration();
        int getCurrentPosition();
        boolean isPlaying();
        void seekTo(int positionMs);
        void start();
        void pause();
        void stop();
    }

    public interface PlaybackStateRepository {
        void startPlayback(int videoId, StartPlaybackCallback callback);
        void updatePlaybackState(
                int videoId,
                double currentTimeSec,
                double durationSec,
                boolean paused,
                boolean completed
        );

        interface StartPlaybackCallback {
            void onSuccess(VideoPlaybackState state);
            void onError();
        }
    }

    public interface Scheduler {
        void post(Runnable runnable);
        void postDelayed(Runnable runnable, long delayMs);
        void removeCallbacks(Runnable runnable);
        void clear();
    }
}
