package com.kuranas.mobile.presentation.music;

import android.media.AudioManager;
import android.media.MediaPlayer;
import android.os.Bundle;
import android.os.Handler;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageButton;
import android.widget.ImageView;
import android.widget.SeekBar;
import android.widget.TextView;

import androidx.fragment.app.Fragment;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.MusicPlayerState;
import com.kuranas.mobile.domain.repository.MusicRepository;
import com.kuranas.mobile.infra.http.ApiCallback;

import java.io.IOException;

public class MusicPlayerFragment extends Fragment {

    private static final String ARG_FILE_ID = "file_id";
    private static final String ARG_TITLE = "title";
    private static final String ARG_ARTIST = "artist";
    private static final String LEGACY_ARG_FILE_ID = "fileId";

    private ImageView albumArt;
    private TextView trackTitle;
    private TextView trackArtist;
    private SeekBar seekBar;
    private TextView currentTime;
    private TextView totalTime;
    private ImageButton btnPrevious;
    private ImageButton btnPlayPause;
    private ImageButton btnNext;

    private MusicPlayerController controller;

    public static MusicPlayerFragment newInstance(int fileId, String title, String artist) {
        MusicPlayerFragment fragment = new MusicPlayerFragment();
        Bundle args = new Bundle();
        args.putInt(ARG_FILE_ID, fileId);
        args.putString(ARG_TITLE, title);
        args.putString(ARG_ARTIST, artist);
        fragment.setArguments(args);
        return fragment;
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_music_player, container, false);

        ServiceLocator locator = ServiceLocator.getInstance();
        bindViews(root);

        controller = new MusicPlayerController(
                new FragmentViewContract(),
                new MediaPlayerAudioEngine(new MediaPlayer()),
                new RepositoryAdapter(locator.getMusicRepository()),
                new HandlerScheduler(new Handler())
        );

        setupControls();
        controller.start(locator.getHttpClient().getBaseUrl(), extractPlayerState(getArguments()));

        return root;
    }

    private void bindViews(View root) {
        albumArt = (ImageView) root.findViewById(R.id.album_art);
        trackTitle = (TextView) root.findViewById(R.id.track_title);
        trackArtist = (TextView) root.findViewById(R.id.track_artist);
        seekBar = (SeekBar) root.findViewById(R.id.seek_bar);
        currentTime = (TextView) root.findViewById(R.id.current_time);
        totalTime = (TextView) root.findViewById(R.id.total_time);
        btnPrevious = (ImageButton) root.findViewById(R.id.btn_previous);
        btnPlayPause = (ImageButton) root.findViewById(R.id.btn_play_pause);
        btnNext = (ImageButton) root.findViewById(R.id.btn_next);
    }

    private MusicPlayerController.PlayerState extractPlayerState(Bundle args) {
        int fileId = 0;
        String title = "";
        String artist = "";

        if (args != null) {
            fileId = args.getInt(ARG_FILE_ID, args.getInt(LEGACY_ARG_FILE_ID, 0));
            title = defaultString(args.getString(ARG_TITLE, ""));
            artist = defaultString(args.getString(ARG_ARTIST, ""));
        }

        return new MusicPlayerController.PlayerState(fileId, title, artist);
    }

    private String defaultString(String value) {
        return value == null ? "" : value;
    }

    private void setupControls() {
        btnPlayPause.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                controller.onPlayPauseClicked();
            }
        });

        btnPrevious.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                controller.onPreviousClicked();
            }
        });

        btnNext.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                // No playlist context in single-track mode
            }
        });

        seekBar.setOnSeekBarChangeListener(new SeekBar.OnSeekBarChangeListener() {
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
                controller.onSeekStopTracking(seekBar.getProgress());
            }
        });
    }

    @Override
    public void onPause() {
        super.onPause();
        if (controller != null) {
            controller.onPause();
        }
    }

    @Override
    public void onResume() {
        super.onResume();
        if (controller != null) {
            controller.onResume();
        }
    }

    @Override
    public void onDestroyView() {
        super.onDestroyView();
        if (controller != null) {
            controller.onDestroy();
            controller = null;
        }
    }

    private final class FragmentViewContract implements MusicPlayerController.ViewContract {

        @Override
        public void setTrackTitle(String title) {
            trackTitle.setText(title);
        }

        @Override
        public void setTrackArtist(String artist) {
            trackArtist.setText(artist);
        }

        @Override
        public void setDefaultAlbumArt() {
            albumArt.setImageResource(R.drawable.ic_music);
        }

        @Override
        public void setSeekMax(int max) {
            seekBar.setMax(max);
        }

        @Override
        public void setSeekProgress(int progress) {
            seekBar.setProgress(progress);
        }

        @Override
        public void setCurrentTimeText(String text) {
            currentTime.setText(text);
        }

        @Override
        public void setTotalTimeText(String text) {
            totalTime.setText(text);
        }

        @Override
        public void setPlayPausePlaying(boolean playing) {
            btnPlayPause.setImageResource(playing ? R.drawable.ic_pause : R.drawable.ic_play);
        }
    }

    private static final class MediaPlayerAudioEngine implements MusicPlayerController.AudioEngine {

        private final MediaPlayer mediaPlayer;

        private MediaPlayerAudioEngine(MediaPlayer mediaPlayer) {
            this.mediaPlayer = mediaPlayer;
        }

        @Override
        public void setAudioStreamTypeMusic() {
            mediaPlayer.setAudioStreamType(AudioManager.STREAM_MUSIC);
        }

        @Override
        public void setDataSource(String streamUrl) throws IOException {
            mediaPlayer.setDataSource(streamUrl);
        }

        @Override
        public void setOnPrepared(final Runnable runnable) {
            mediaPlayer.setOnPreparedListener(new MediaPlayer.OnPreparedListener() {
                @Override
                public void onPrepared(MediaPlayer mp) {
                    runnable.run();
                }
            });
        }

        @Override
        public void setOnCompletion(final Runnable runnable) {
            mediaPlayer.setOnCompletionListener(new MediaPlayer.OnCompletionListener() {
                @Override
                public void onCompletion(MediaPlayer mp) {
                    runnable.run();
                }
            });
        }

        @Override
        public void setOnError(final ErrorCallback callback) {
            mediaPlayer.setOnErrorListener(new MediaPlayer.OnErrorListener() {
                @Override
                public boolean onError(MediaPlayer mp, int what, int extra) {
                    return callback.onError(what, extra);
                }
            });
        }

        @Override
        public void prepareAsync() {
            mediaPlayer.prepareAsync();
        }

        @Override
        public boolean isPlaying() {
            return mediaPlayer.isPlaying();
        }

        @Override
        public void start() {
            mediaPlayer.start();
        }

        @Override
        public void pause() {
            mediaPlayer.pause();
        }

        @Override
        public void seekTo(int positionMs) {
            mediaPlayer.seekTo(positionMs);
        }

        @Override
        public int getDuration() {
            return mediaPlayer.getDuration();
        }

        @Override
        public int getCurrentPosition() {
            return mediaPlayer.getCurrentPosition();
        }

        @Override
        public void release() {
            mediaPlayer.release();
        }
    }

    private static final class RepositoryAdapter implements MusicPlayerController.PlayerStateRepository {

        private final MusicRepository musicRepository;

        private RepositoryAdapter(MusicRepository musicRepository) {
            this.musicRepository = musicRepository;
        }

        @Override
        public void updatePlayerState(int playlistId, int fileId, double positionSeconds) {
            musicRepository.updatePlayerState(
                    playlistId,
                    fileId,
                    positionSeconds,
                    new ApiCallback<MusicPlayerState>() {
                        @Override
                        public void onSuccess(MusicPlayerState result) {
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

    private static final class HandlerScheduler implements MusicPlayerController.Scheduler {

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
