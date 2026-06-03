package com.kuranas.mobile.presentation.video;

import android.view.View;
import android.widget.ImageButton;
import android.widget.LinearLayout;
import android.widget.SeekBar;
import android.widget.TextView;
import android.widget.VideoView;

import androidx.appcompat.app.AppCompatActivity;

import com.kuranas.mobile.R;

public final class VideoPlayerViewBinder implements VideoPlayerController.ViewContract {

    private final VideoView videoView;
    private final LinearLayout controlsOverlay;
    private final SeekBar videoSeek;
    private final ImageButton btnPlayPause;
    private final TextView videoTime;
    private final TextView videoTitle;

    public VideoPlayerViewBinder(AppCompatActivity activity) {
        this.videoView = (VideoView) activity.findViewById(R.id.video_view);
        this.controlsOverlay = (LinearLayout) activity.findViewById(R.id.controls_overlay);
        this.videoSeek = (SeekBar) activity.findViewById(R.id.video_seek);
        this.btnPlayPause = (ImageButton) activity.findViewById(R.id.btn_play_pause);
        this.videoTime = (TextView) activity.findViewById(R.id.video_time);
        this.videoTitle = (TextView) activity.findViewById(R.id.video_title);
    }

    public VideoView getVideoView() {
        return videoView;
    }

    public void setVideoTitle(String title) {
        videoTitle.setText(title == null ? "" : title);
    }

    public void setOnVideoClickListener(View.OnClickListener listener) {
        videoView.setOnClickListener(listener);
    }

    public void setOnPlayPauseClickListener(View.OnClickListener listener) {
        btnPlayPause.setOnClickListener(listener);
    }

    public void setOnSeekBarChangeListener(SeekBar.OnSeekBarChangeListener listener) {
        videoSeek.setOnSeekBarChangeListener(listener);
    }

    @Override
    public void setPlayPausePlaying(boolean playing) {
        btnPlayPause.setImageResource(playing ? R.drawable.ic_pause : R.drawable.ic_play);
    }

    @Override
    public void setControlsVisible(boolean visible) {
        controlsOverlay.setVisibility(visible ? View.VISIBLE : View.GONE);
    }

    @Override
    public boolean isControlsVisible() {
        return controlsOverlay.getVisibility() == View.VISIBLE;
    }

    @Override
    public void setSeekMax(int max) {
        videoSeek.setMax(max);
    }

    @Override
    public void setSeekProgress(int progress) {
        videoSeek.setProgress(progress);
    }

    @Override
    public void setTimeText(String text) {
        videoTime.setText(text);
    }
}
