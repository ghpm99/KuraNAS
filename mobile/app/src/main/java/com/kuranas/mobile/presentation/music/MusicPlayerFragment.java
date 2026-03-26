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
    private static final long UPDATE_INTERVAL_MS = 1000;

    private ImageView albumArt;
    private TextView trackTitle;
    private TextView trackArtist;
    private SeekBar seekBar;
    private TextView currentTime;
    private TextView totalTime;
    private ImageButton btnPrevious;
    private ImageButton btnPlayPause;
    private ImageButton btnNext;

    private MediaPlayer mediaPlayer;
    private Handler handler;
    private Runnable updateRunnable;
    private MusicRepository musicRepository;

    private int fileId;
    private String title;
    private String artist;
    private boolean isPrepared;
    private boolean isUserSeeking;

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
        musicRepository = locator.getMusicRepository();
        handler = new Handler();

        albumArt = (ImageView) root.findViewById(R.id.album_art);
        trackTitle = (TextView) root.findViewById(R.id.track_title);
        trackArtist = (TextView) root.findViewById(R.id.track_artist);
        seekBar = (SeekBar) root.findViewById(R.id.seek_bar);
        currentTime = (TextView) root.findViewById(R.id.current_time);
        totalTime = (TextView) root.findViewById(R.id.total_time);
        btnPrevious = (ImageButton) root.findViewById(R.id.btn_previous);
        btnPlayPause = (ImageButton) root.findViewById(R.id.btn_play_pause);
        btnNext = (ImageButton) root.findViewById(R.id.btn_next);

        Bundle args = getArguments();
        if (args != null) {
            fileId = args.getInt(ARG_FILE_ID, 0);
            title = args.getString(ARG_TITLE, "");
            artist = args.getString(ARG_ARTIST, "");
        }

        trackTitle.setText(title);
        trackArtist.setText(artist);

        albumArt.setImageResource(R.drawable.ic_music);

        setupControls();
        startPlayback();

        return root;
    }

    private void setupControls() {
        btnPlayPause.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                togglePlayPause();
            }
        });

        btnPrevious.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                if (mediaPlayer != null && isPrepared) {
                    mediaPlayer.seekTo(0);
                    updateSeekBar();
                }
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
                if (fromUser && mediaPlayer != null && isPrepared) {
                    currentTime.setText(formatTime(progress));
                }
            }

            @Override
            public void onStartTrackingTouch(SeekBar seekBar) {
                isUserSeeking = true;
            }

            @Override
            public void onStopTrackingTouch(SeekBar seekBar) {
                isUserSeeking = false;
                if (mediaPlayer != null && isPrepared) {
                    mediaPlayer.seekTo(seekBar.getProgress());
                }
            }
        });
    }

    private void startPlayback() {
        String streamUrl = ServiceLocator.getInstance().getHttpClient().getBaseUrl()
                + "/api/v1/files/stream/" + fileId;

        mediaPlayer = new MediaPlayer();
        mediaPlayer.setAudioStreamType(AudioManager.STREAM_MUSIC);

        try {
            mediaPlayer.setDataSource(streamUrl);
        } catch (IOException e) {
            return;
        }

        mediaPlayer.setOnPreparedListener(new MediaPlayer.OnPreparedListener() {
            @Override
            public void onPrepared(MediaPlayer mp) {
                isPrepared = true;
                int duration = mp.getDuration();
                seekBar.setMax(duration);
                totalTime.setText(formatTime(duration));
                mp.start();
                btnPlayPause.setImageResource(R.drawable.ic_pause);
                startSeekBarUpdates();
            }
        });

        mediaPlayer.setOnCompletionListener(new MediaPlayer.OnCompletionListener() {
            @Override
            public void onCompletion(MediaPlayer mp) {
                btnPlayPause.setImageResource(R.drawable.ic_play);
                stopSeekBarUpdates();
                savePlayerState();
            }
        });

        mediaPlayer.setOnErrorListener(new MediaPlayer.OnErrorListener() {
            @Override
            public boolean onError(MediaPlayer mp, int what, int extra) {
                isPrepared = false;
                return true;
            }
        });

        mediaPlayer.prepareAsync();
    }

    private void togglePlayPause() {
        if (mediaPlayer == null || !isPrepared) {
            return;
        }
        if (mediaPlayer.isPlaying()) {
            mediaPlayer.pause();
            btnPlayPause.setImageResource(R.drawable.ic_play);
            stopSeekBarUpdates();
            savePlayerState();
        } else {
            mediaPlayer.start();
            btnPlayPause.setImageResource(R.drawable.ic_pause);
            startSeekBarUpdates();
        }
    }

    private void startSeekBarUpdates() {
        updateRunnable = new Runnable() {
            @Override
            public void run() {
                updateSeekBar();
                handler.postDelayed(this, UPDATE_INTERVAL_MS);
            }
        };
        handler.post(updateRunnable);
    }

    private void stopSeekBarUpdates() {
        if (handler != null && updateRunnable != null) {
            handler.removeCallbacks(updateRunnable);
        }
    }

    private void updateSeekBar() {
        if (mediaPlayer != null && isPrepared && !isUserSeeking) {
            int position = mediaPlayer.getCurrentPosition();
            seekBar.setProgress(position);
            currentTime.setText(formatTime(position));
        }
    }

    private void savePlayerState() {
        if (mediaPlayer == null || !isPrepared) {
            return;
        }
        double positionSeconds = mediaPlayer.getCurrentPosition() / 1000.0;
        musicRepository.updatePlayerState(0, fileId, positionSeconds,
                new ApiCallback<MusicPlayerState>() {
                    @Override
                    public void onSuccess(MusicPlayerState result) {
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
        int minutes = totalSeconds / 60;
        int seconds = totalSeconds % 60;
        return String.format("%d:%02d", minutes, seconds);
    }

    @Override
    public void onPause() {
        super.onPause();
        if (mediaPlayer != null && isPrepared && mediaPlayer.isPlaying()) {
            mediaPlayer.pause();
            btnPlayPause.setImageResource(R.drawable.ic_play);
            stopSeekBarUpdates();
            savePlayerState();
        }
    }

    @Override
    public void onResume() {
        super.onResume();
        if (mediaPlayer != null && isPrepared) {
            mediaPlayer.start();
            btnPlayPause.setImageResource(R.drawable.ic_pause);
            startSeekBarUpdates();
        }
    }

    @Override
    public void onDestroyView() {
        super.onDestroyView();
        stopSeekBarUpdates();
        if (mediaPlayer != null) {
            if (isPrepared) {
                savePlayerState();
            }
            mediaPlayer.release();
            mediaPlayer = null;
            isPrepared = false;
        }
    }
}
