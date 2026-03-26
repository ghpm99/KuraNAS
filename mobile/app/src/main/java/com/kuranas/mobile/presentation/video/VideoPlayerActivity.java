package com.kuranas.mobile.presentation.video;

import android.net.Uri;
import android.os.Bundle;
import android.os.Handler;
import android.view.View;
import android.widget.ImageButton;
import android.widget.LinearLayout;
import android.widget.MediaController;
import android.widget.SeekBar;
import android.widget.TextView;
import android.widget.VideoView;

import androidx.appcompat.app.AppCompatActivity;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.infra.kiosk.KioskManager;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.VideoPlaybackState;
import com.kuranas.mobile.domain.repository.VideoRepository;
import com.kuranas.mobile.infra.http.ApiCallback;

public class VideoPlayerActivity extends AppCompatActivity {

    public static final String EXTRA_VIDEO_ID = "extra_video_id";
    public static final String EXTRA_VIDEO_NAME = "extra_video_name";
    public static final String EXTRA_STREAM_URL = "extra_stream_url";

    private static final long CONTROLS_HIDE_DELAY_MS = 3000;
    private static final long STATE_UPDATE_INTERVAL_MS = 5000;

    private VideoView videoView;
    private LinearLayout controlsOverlay;
    private SeekBar videoSeek;
    private ImageButton btnPlayPause;
    private TextView videoTime;
    private TextView videoTitle;

    private KioskManager kioskManager;
    private VideoRepository videoRepository;
    private Handler handler;

    private int videoId;
    private String videoName;
    private String streamUrl;
    private boolean isPlaying;

    private Runnable hideControlsRunnable;
    private Runnable updateStateRunnable;
    private Runnable updateSeekRunnable;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_video_player);

        kioskManager = new KioskManager(this);
        kioskManager.engage();

        ServiceLocator locator = ServiceLocator.getInstance();
        videoRepository = locator.getVideoRepository();
        handler = new Handler();

        videoView = (VideoView) findViewById(R.id.video_view);
        controlsOverlay = (LinearLayout) findViewById(R.id.controls_overlay);
        videoSeek = (SeekBar) findViewById(R.id.video_seek);
        btnPlayPause = (ImageButton) findViewById(R.id.btn_play_pause);
        videoTime = (TextView) findViewById(R.id.video_time);
        videoTitle = (TextView) findViewById(R.id.video_title);

        videoId = getIntent().getIntExtra(EXTRA_VIDEO_ID, 0);
        videoName = getIntent().getStringExtra(EXTRA_VIDEO_NAME);
        streamUrl = getIntent().getStringExtra(EXTRA_STREAM_URL);

        if (videoName == null) {
            videoName = "";
        }
        videoTitle.setText(videoName);

        setupControls();
        startVideoPlayback();
    }

    private void setupControls() {
        videoView.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                toggleControlsVisibility();
            }
        });

        btnPlayPause.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                togglePlayPause();
            }
        });

        videoSeek.setOnSeekBarChangeListener(new SeekBar.OnSeekBarChangeListener() {
            @Override
            public void onProgressChanged(SeekBar seekBar, int progress, boolean fromUser) {
                if (fromUser) {
                    videoView.seekTo(progress);
                    updateTimeDisplay();
                }
            }

            @Override
            public void onStartTrackingTouch(SeekBar seekBar) {
                // Keep controls visible while seeking
                cancelHideControls();
            }

            @Override
            public void onStopTrackingTouch(SeekBar seekBar) {
                scheduleHideControls();
            }
        });

        hideControlsRunnable = new Runnable() {
            @Override
            public void run() {
                controlsOverlay.setVisibility(View.GONE);
            }
        };

        updateStateRunnable = new Runnable() {
            @Override
            public void run() {
                savePlaybackState(false);
                handler.postDelayed(this, STATE_UPDATE_INTERVAL_MS);
            }
        };

        updateSeekRunnable = new Runnable() {
            @Override
            public void run() {
                if (videoView.isPlaying()) {
                    int position = videoView.getCurrentPosition();
                    videoSeek.setProgress(position);
                    updateTimeDisplay();
                }
                handler.postDelayed(this, 1000);
            }
        };
    }

    private void startVideoPlayback() {
        videoRepository.startPlayback(videoId, new ApiCallback<VideoPlaybackState>() {
            @Override
            public void onSuccess(VideoPlaybackState state) {
                playVideo(state);
            }

            @Override
            public void onError(AppError error) {
                // Play from beginning on error
                playVideo(null);
            }
        });
    }

    private void playVideo(final VideoPlaybackState state) {
        videoView.setVideoURI(Uri.parse(streamUrl));

        videoView.setOnPreparedListener(new android.media.MediaPlayer.OnPreparedListener() {
            @Override
            public void onPrepared(android.media.MediaPlayer mp) {
                int duration = videoView.getDuration();
                videoSeek.setMax(duration);

                if (state != null && state.getCurrentTime() > 0 && !state.isCompleted()) {
                    int seekPosition = (int) (state.getCurrentTime() * 1000);
                    videoView.seekTo(seekPosition);
                }

                videoView.start();
                isPlaying = true;
                btnPlayPause.setImageResource(R.drawable.ic_pause);
                updateTimeDisplay();
                scheduleHideControls();
                startSeekBarUpdates();
                startStateUpdates();
            }
        });

        videoView.setOnCompletionListener(new android.media.MediaPlayer.OnCompletionListener() {
            @Override
            public void onCompletion(android.media.MediaPlayer mp) {
                isPlaying = false;
                btnPlayPause.setImageResource(R.drawable.ic_play);
                controlsOverlay.setVisibility(View.VISIBLE);
                stopSeekBarUpdates();
                stopStateUpdates();
                savePlaybackState(true);
            }
        });
    }

    private void togglePlayPause() {
        if (videoView.isPlaying()) {
            videoView.pause();
            isPlaying = false;
            btnPlayPause.setImageResource(R.drawable.ic_play);
            stopSeekBarUpdates();
            savePlaybackState(false);
        } else {
            videoView.start();
            isPlaying = true;
            btnPlayPause.setImageResource(R.drawable.ic_pause);
            startSeekBarUpdates();
            scheduleHideControls();
        }
    }

    private void toggleControlsVisibility() {
        if (controlsOverlay.getVisibility() == View.VISIBLE) {
            controlsOverlay.setVisibility(View.GONE);
        } else {
            controlsOverlay.setVisibility(View.VISIBLE);
            scheduleHideControls();
        }
    }

    private void scheduleHideControls() {
        cancelHideControls();
        handler.postDelayed(hideControlsRunnable, CONTROLS_HIDE_DELAY_MS);
    }

    private void cancelHideControls() {
        handler.removeCallbacks(hideControlsRunnable);
    }

    private void startSeekBarUpdates() {
        handler.post(updateSeekRunnable);
    }

    private void stopSeekBarUpdates() {
        handler.removeCallbacks(updateSeekRunnable);
    }

    private void startStateUpdates() {
        handler.postDelayed(updateStateRunnable, STATE_UPDATE_INTERVAL_MS);
    }

    private void stopStateUpdates() {
        handler.removeCallbacks(updateStateRunnable);
    }

    private void updateTimeDisplay() {
        int current = videoView.getCurrentPosition();
        int duration = videoView.getDuration();
        String timeText = formatTime(current) + " / " + formatTime(duration);
        videoTime.setText(timeText);
    }

    private void savePlaybackState(boolean completed) {
        double currentTimeSec = videoView.getCurrentPosition() / 1000.0;
        double durationSec = videoView.getDuration() / 1000.0;
        boolean isPaused = !videoView.isPlaying();

        videoRepository.updatePlaybackState(videoId, currentTimeSec, durationSec,
                isPaused, completed, new ApiCallback<VideoPlaybackState>() {
                    @Override
                    public void onSuccess(VideoPlaybackState result) {
                        // State saved
                    }

                    @Override
                    public void onError(AppError error) {
                        // Silent failure
                    }
                });
    }

    private String formatTime(int millis) {
        int totalSeconds = millis / 1000;
        int hours = totalSeconds / 3600;
        int minutes = (totalSeconds % 3600) / 60;
        int seconds = totalSeconds % 60;
        if (hours > 0) {
            return String.format("%d:%02d:%02d", hours, minutes, seconds);
        }
        return String.format("%d:%02d", minutes, seconds);
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (videoView.isPlaying()) {
            videoView.pause();
            isPlaying = false;
            savePlaybackState(false);
        }
        stopSeekBarUpdates();
        stopStateUpdates();
        cancelHideControls();
    }

    @Override
    protected void onResume() {
        super.onResume();
        kioskManager.engage();
        if (isPlaying) {
            videoView.start();
            startSeekBarUpdates();
            startStateUpdates();
        }
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        handler.removeCallbacksAndMessages(null);
        videoView.stopPlayback();
    }
}
