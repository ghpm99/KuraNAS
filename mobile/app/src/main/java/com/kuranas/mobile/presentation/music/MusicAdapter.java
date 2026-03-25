package com.kuranas.mobile.presentation.music;

import android.graphics.Color;
import android.graphics.Typeface;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.recyclerview.widget.RecyclerView;

import com.kuranas.mobile.R;
import com.kuranas.mobile.domain.model.Track;
import com.kuranas.mobile.i18n.TranslationManager;

import java.util.ArrayList;
import java.util.List;

public class MusicAdapter extends RecyclerView.Adapter<MusicAdapter.TrackViewHolder> {

    private static final int HIGHLIGHT_COLOR = Color.parseColor("#BB86FC");

    private final List<Track> items;
    private final OnTrackClickListener listener;
    private final TranslationManager translations;
    private int currentTrackFileId = -1;

    public interface OnTrackClickListener {
        void onTrackClick(Track track, int position);
    }

    public MusicAdapter(List<Track> items, OnTrackClickListener listener,
                        TranslationManager translations) {
        this.items = new ArrayList<Track>(items);
        this.listener = listener;
        this.translations = translations;
    }

    @Override
    public TrackViewHolder onCreateViewHolder(ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_track, parent, false);
        return new TrackViewHolder(view);
    }

    @Override
    public void onBindViewHolder(TrackViewHolder holder, final int position) {
        final Track track = items.get(position);

        holder.trackNumber.setText(String.valueOf(position + 1));
        holder.trackTitle.setText(track.getDisplayTitle());
        holder.trackArtist.setText(track.getDisplayArtist());
        holder.trackDuration.setText(track.getFormattedDuration());

        boolean isPlaying = track.getFileId() == currentTrackFileId;
        applyHighlight(holder, isPlaying);

        holder.itemView.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                if (listener != null) {
                    listener.onTrackClick(track, position);
                }
            }
        });
    }

    @Override
    public int getItemCount() {
        return items.size();
    }

    public void setCurrentTrackId(int fileId) {
        int previousId = currentTrackFileId;
        currentTrackFileId = fileId;

        for (int i = 0; i < items.size(); i++) {
            int trackFileId = items.get(i).getFileId();
            if (trackFileId == previousId || trackFileId == fileId) {
                notifyItemChanged(i);
            }
        }
    }

    public void updateItems(List<Track> newItems) {
        items.clear();
        items.addAll(newItems);
        notifyDataSetChanged();
    }

    public void addItems(List<Track> moreItems) {
        int startPosition = items.size();
        items.addAll(moreItems);
        notifyItemRangeInserted(startPosition, moreItems.size());
    }

    private void applyHighlight(TrackViewHolder holder, boolean isPlaying) {
        int titleColor = isPlaying ? HIGHLIGHT_COLOR : Color.WHITE;
        int titleStyle = isPlaying ? Typeface.BOLD : Typeface.NORMAL;

        holder.trackTitle.setTextColor(titleColor);
        holder.trackTitle.setTypeface(null, titleStyle);
        holder.trackNumber.setTextColor(titleColor);
    }

    static class TrackViewHolder extends RecyclerView.ViewHolder {

        final TextView trackNumber;
        final TextView trackTitle;
        final TextView trackArtist;
        final TextView trackDuration;

        TrackViewHolder(View itemView) {
            super(itemView);
            trackNumber = (TextView) itemView.findViewById(R.id.track_number);
            trackTitle = (TextView) itemView.findViewById(R.id.track_title);
            trackArtist = (TextView) itemView.findViewById(R.id.track_artist);
            trackDuration = (TextView) itemView.findViewById(R.id.track_duration);
        }
    }
}
