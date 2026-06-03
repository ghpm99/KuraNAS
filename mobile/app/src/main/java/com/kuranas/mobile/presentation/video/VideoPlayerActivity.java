package com.kuranas.mobile.presentation.video;

import android.os.Bundle;
import android.os.Handler;
import android.view.View;
import android.widget.SeekBar;
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

    private KioskManager kioskManager;
    private VideoPlayerController controller;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_video_player);

        kioskManager = new KioskManager(this);
        kioskManager.engage();

        ServiceLocator locator = ServiceLocator.getInstance();
        VideoPlayerViewBinder viewBinder = new VideoPlayerViewBinder(this);

        int videoId = getIntent().getIntExtra(EXTRA_VIDEO_ID, 0);
        String videoName = getIntent().getStringExtra(EXTRA_VIDEO_NAME);
        String streamUrl = getIntent().getStringExtra(EXTRA_STREAM_URL);

        viewBinder.setVideoTitle(videoName);

        controller = new VideoPlayerController(
                new VideoPlaybackRepositoryAdapter(locator.getVideoRepository()),
                new VideoViewPlaybackEngine(viewBinder.getVideoView()),
                viewBinder,
                new HandlerScheduler(new Handler())
        );

        bindControls(viewBinder);
        controller.start(videoId, streamUrl);
    }

    private void bindControls(VideoPlayerViewBinder viewBinder) {
        viewBinder.setOnVideoClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                controller.onVideoTapped();
            }
        });

        viewBinder.setOnPlayPauseClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                controller.onPlayPauseClicked();
            }
        });

        viewBinder.setOnSeekBarChangeListener(new SeekBar.OnSeekBarChangeListener() {
            @Override
            public void onProgressChanged(SeekBar seekBar, int progress, boolean fromUser) {
                controller.onSeekProgressChanged(progress, fromUser);
            }

            @Override
            public void onStartTrackingTouch(SeekBar seekBar) {
                controller.onSeekStartTracking();
            }

            @Override
            public void onStopTrackingTouch(SeekBar seekBar) {
                controller.onSeekStopTracking();
            }
        });
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (controller != null) {
            controller.onPause();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        kioskManager.engage();
        if (controller != null) {
            controller.onResume();
        }
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        if (controller != null) {
            controller.onDestroy();
        }
    }

    private static final class VideoPlaybackRepositoryAdapter
            implements VideoPlayerController.PlaybackStateRepository {

        private final VideoRepository videoRepository;

        private VideoPlaybackRepositoryAdapter(VideoRepository videoRepository) {
            this.videoRepository = videoRepository;
        }

        @Override
        public void startPlayback(int videoId, final StartPlaybackCallback callback) {
            videoRepository.startPlayback(videoId, new ApiCallback<VideoPlaybackState>() {
                @Override
                public void onSuccess(VideoPlaybackState state) {
                    callback.onSuccess(state);
                }

                @Override
                public void onError(AppError error) {
                    callback.onError();
                }
            });
        }

        @Override
        public void updatePlaybackState(
                int videoId,
                double currentTimeSec,
                double durationSec,
                boolean paused,
                boolean completed
        ) {
            videoRepository.updatePlaybackState(
                    videoId,
                    currentTimeSec,
                    durationSec,
                    paused,
                    completed,
                    new ApiCallback<VideoPlaybackState>() {
                        @Override
                        public void onSuccess(VideoPlaybackState result) {
                            // Silent success
                        }

                        @Override
                        public void onError(AppError error) {
                            // Silent failure
                        }
                    }
            );
        }
    }

    private static final class VideoViewPlaybackEngine
            implements VideoPlayerController.PlaybackEngine {

        private final VideoView videoView;

        private VideoViewPlaybackEngine(VideoView videoView) {
            this.videoView = videoView;
        }

        @Override
        public void setVideoUri(String streamUrl) {
            videoView.setVideoURI(android.net.Uri.parse(streamUrl));
        }

        @Override
        public void setOnPrepared(final Runnable runnable) {
            videoView.setOnPreparedListener(new android.media.MediaPlayer.OnPreparedListener() {
                @Override
                public void onPrepared(android.media.MediaPlayer mp) {
                    runnable.run();
                }
            });
        }

        @Override
        public void setOnCompletion(final Runnable runnable) {
            videoView.setOnCompletionListener(new android.media.MediaPlayer.OnCompletionListener() {
                @Override
                public void onCompletion(android.media.MediaPlayer mp) {
                    runnable.run();
                }
            });
        }

        @Override
        public int getDuration() {
            return videoView.getDuration();
        }

        @Override
        public int getCurrentPosition() {
            return videoView.getCurrentPosition();
        }

        @Override
        public boolean isPlaying() {
            return videoView.isPlaying();
        }

        @Override
        public void seekTo(int positionMs) {
            videoView.seekTo(positionMs);
        }

        @Override
        public void start() {
            videoView.start();
        }

        @Override
        public void pause() {
            videoView.pause();
        }

        @Override
        public void stop() {
            videoView.stopPlayback();
        }
    }

    private static final class HandlerScheduler implements VideoPlayerController.Scheduler {

        private final Handler handler;

        private HandlerScheduler(Handler handler) {
            this.handler = handler;
        }

        @Override
        public void post(Runnable runnable) {
            handler.post(runnable);
        }

        @Override
        public void postDelayed(Runnable runnable, long delayMs) {
            handler.postDelayed(runnable, delayMs);
        }

        @Override
        public void removeCallbacks(Runnable runnable) {
            handler.removeCallbacks(runnable);
        }

        @Override
        public void clear() {
            handler.removeCallbacksAndMessages(null);
        }
    }
}
