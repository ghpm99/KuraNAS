package com.kuranas.mobile.presentation.images;

import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;

import androidx.recyclerview.widget.RecyclerView;

import com.kuranas.mobile.R;
import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.presentation.common.BitmapLoaderTask;

import java.util.ArrayList;
import java.util.List;

public class ImagesAdapter extends RecyclerView.Adapter<ImagesAdapter.ImageViewHolder> {

    private static final int THUMBNAIL_SIZE = 200;

    private final List<FileItem> items;
    private final OnImageClickListener listener;
    private final BitmapCache bitmapCache;
    private final String baseUrl;

    public interface OnImageClickListener {
        void onImageClick(int position);
    }

    public ImagesAdapter(List<FileItem> items, OnImageClickListener listener,
                         BitmapCache bitmapCache, String baseUrl) {
        this.items = new ArrayList<FileItem>(items);
        this.listener = listener;
        this.bitmapCache = bitmapCache;
        this.baseUrl = baseUrl;
    }

    @Override
    public ImageViewHolder onCreateViewHolder(ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_image_grid, parent, false);
        return new ImageViewHolder(view);
    }

    @Override
    public void onBindViewHolder(ImageViewHolder holder, final int position) {
        FileItem item = items.get(position);

        String thumbnailUrl = baseUrl + "/files/thumbnail/" + item.getId();
        holder.imageThumbnail.setImageResource(R.drawable.ic_image);
        BitmapLoaderTask.load(thumbnailUrl, holder.imageThumbnail, bitmapCache,
                THUMBNAIL_SIZE, THUMBNAIL_SIZE);

        holder.itemView.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                if (listener != null) {
                    listener.onImageClick(position);
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

    public void addItems(List<FileItem> moreItems) {
        int startPosition = items.size();
        items.addAll(moreItems);
        notifyItemRangeInserted(startPosition, moreItems.size());
    }

    static class ImageViewHolder extends RecyclerView.ViewHolder {

        final ImageView imageThumbnail;

        ImageViewHolder(View itemView) {
            super(itemView);
            imageThumbnail = (ImageView) itemView.findViewById(R.id.image_thumbnail);
        }
    }
}
