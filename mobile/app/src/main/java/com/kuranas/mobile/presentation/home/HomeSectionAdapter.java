package com.kuranas.mobile.presentation.home;

import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.TextView;

import androidx.recyclerview.widget.RecyclerView;

import com.kuranas.mobile.R;
import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.presentation.common.BitmapLoaderTask;

import java.util.ArrayList;
import java.util.List;

public class HomeSectionAdapter extends RecyclerView.Adapter<HomeSectionAdapter.HomeSectionViewHolder> {

    private static final int THUMBNAIL_SIZE = 120;

    private final List<FileItem> items;
    private final OnItemClickListener listener;
    private final BitmapCache bitmapCache;
    private final String baseUrl;

    public interface OnItemClickListener {
        void onItemClick(FileItem item);
    }

    public HomeSectionAdapter(List<FileItem> items, OnItemClickListener listener,
                              BitmapCache bitmapCache, String baseUrl) {
        this.items = new ArrayList<FileItem>(items);
        this.listener = listener;
        this.bitmapCache = bitmapCache;
        this.baseUrl = baseUrl;
    }

    @Override
    public HomeSectionViewHolder onCreateViewHolder(ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_home_section, parent, false);
        return new HomeSectionViewHolder(view);
    }

    @Override
    public void onBindViewHolder(HomeSectionViewHolder holder, int position) {
        final FileItem item = items.get(position);

        holder.sectionTitle.setText(item.getName());

        holder.sectionThumbnail.setImageResource(getIconResource(item));
        if (item.isImage() || item.isVideo()) {
            String thumbnailUrl = baseUrl + "/api/v1/files/thumbnail/" + item.getId()
                    + "?width=" + THUMBNAIL_SIZE
                    + "&height=" + THUMBNAIL_SIZE;
            BitmapLoaderTask.load(thumbnailUrl, holder.sectionThumbnail, bitmapCache,
                    THUMBNAIL_SIZE, THUMBNAIL_SIZE);
        }

        holder.itemView.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                if (listener != null) {
                    listener.onItemClick(item);
                }
            }
        });
    }

    @Override
    public int getItemCount() {
        return items.size();
    }

    public void updateItems(List<FileItem> newItems) {
        items.clear();
        items.addAll(newItems);
        notifyDataSetChanged();
    }

    private int getIconResource(FileItem item) {
        if (item.isDirectory()) {
            return R.drawable.ic_folder;
        }
        if (item.isImage()) {
            return R.drawable.ic_image;
        }
        if (item.isAudio()) {
            return R.drawable.ic_music;
        }
        if (item.isVideo()) {
            return R.drawable.ic_video;
        }
        return R.drawable.ic_file;
    }

    static class HomeSectionViewHolder extends RecyclerView.ViewHolder {

        final ImageView sectionThumbnail;
        final TextView sectionTitle;

        HomeSectionViewHolder(View itemView) {
            super(itemView);
            sectionThumbnail = (ImageView) itemView.findViewById(R.id.section_thumbnail);
            sectionTitle = (TextView) itemView.findViewById(R.id.section_title);
        }
    }
}
