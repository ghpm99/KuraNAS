package com.kuranas.mobile.presentation.video;

import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.TextView;

import androidx.recyclerview.widget.RecyclerView;

import com.kuranas.mobile.R;
import com.kuranas.mobile.domain.model.VideoItem;
import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.presentation.common.BitmapLoaderTask;

import java.util.ArrayList;
import java.util.List;

public class VideoAdapter extends RecyclerView.Adapter<VideoAdapter.VideoViewHolder> {

    private static final int THUMBNAIL_WIDTH = 320;
    private static final int THUMBNAIL_HEIGHT = 180;

    private final List<VideoItem> items;
    private final OnVideoClickListener listener;
    private final BitmapCache bitmapCache;
    private final String baseUrl;

    public interface OnVideoClickListener {
        void onVideoClick(VideoItem video);
    }

    public VideoAdapter(List<VideoItem> items, OnVideoClickListener listener,
                        BitmapCache bitmapCache, String baseUrl) {
        this.items = new ArrayList<VideoItem>(items);
        this.listener = listener;
        this.bitmapCache = bitmapCache;
        this.baseUrl = baseUrl;
    }

    @Override
    public VideoViewHolder onCreateViewHolder(ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_video, parent, false);
        return new VideoViewHolder(view);
    }

    @Override
    public void onBindViewHolder(VideoViewHolder holder, int position) {
        final VideoItem video = items.get(position);

        holder.videoName.setText(video.getDisplayName());
        holder.videoInfo.setText(buildVideoInfo(video));

        holder.videoThumbnail.setImageResource(R.drawable.ic_video);
        String thumbnailUrl = baseUrl + "/files/thumbnail/" + video.getId();
        BitmapLoaderTask.load(thumbnailUrl, holder.videoThumbnail, bitmapCache,
                THUMBNAIL_WIDTH, THUMBNAIL_HEIGHT);

        holder.itemView.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                if (listener != null) {
                    listener.onVideoClick(video);
                }
            }
        });
    }

    @Override
    public int getItemCount() {
        return items.size();
    }

    public void updateItems(List<VideoItem> newItems) {
        items.clear();
        items.addAll(newItems);
        notifyDataSetChanged();
    }

    public void addItems(List<VideoItem> moreItems) {
        int startPosition = items.size();
        items.addAll(moreItems);
        notifyItemRangeInserted(startPosition, moreItems.size());
    }

    private String buildVideoInfo(VideoItem video) {
        String format = video.getFormat() != null ? video.getFormat().toUpperCase() : "";
        String size = formatFileSize(video.getSize());
        if (format.isEmpty()) {
            return size;
        }
        return format + " - " + size;
    }

    private String formatFileSize(long sizeBytes) {
        if (sizeBytes < 1024) {
            return sizeBytes + " B";
        }
        if (sizeBytes < 1024 * 1024) {
            return String.format("%.1f KB", sizeBytes / 1024.0);
        }
        if (sizeBytes < 1024 * 1024 * 1024) {
            return String.format("%.1f MB", sizeBytes / (1024.0 * 1024.0));
        }
        return String.format("%.1f GB", sizeBytes / (1024.0 * 1024.0 * 1024.0));
    }

    static class VideoViewHolder extends RecyclerView.ViewHolder {

        final ImageView videoThumbnail;
        final TextView videoName;
        final TextView videoInfo;

        VideoViewHolder(View itemView) {
            super(itemView);
            videoThumbnail = (ImageView) itemView.findViewById(R.id.video_thumbnail);
            videoName = (TextView) itemView.findViewById(R.id.video_name);
            videoInfo = (TextView) itemView.findViewById(R.id.video_info);
        }
    }
}
