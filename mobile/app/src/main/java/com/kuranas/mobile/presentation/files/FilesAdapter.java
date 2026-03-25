package com.kuranas.mobile.presentation.files;

import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.TextView;

import androidx.recyclerview.widget.RecyclerView;

import com.kuranas.mobile.R;
import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.i18n.TranslationManager;
import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.presentation.common.BitmapLoaderTask;

import java.util.ArrayList;
import java.util.List;

public class FilesAdapter extends RecyclerView.Adapter<FilesAdapter.FileViewHolder> {

    private static final int THUMBNAIL_SIZE = 48;

    private final List<FileItem> items;
    private final OnItemClickListener listener;
    private final TranslationManager translations;
    private final BitmapCache bitmapCache;
    private final String baseUrl;

    public interface OnItemClickListener {
        void onItemClick(FileItem item);
    }

    public FilesAdapter(List<FileItem> items, OnItemClickListener listener,
                        TranslationManager translations, BitmapCache bitmapCache,
                        String baseUrl) {
        this.items = new ArrayList<FileItem>(items);
        this.listener = listener;
        this.translations = translations;
        this.bitmapCache = bitmapCache;
        this.baseUrl = baseUrl;
    }

    @Override
    public FileViewHolder onCreateViewHolder(ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_file, parent, false);
        return new FileViewHolder(view);
    }

    @Override
    public void onBindViewHolder(FileViewHolder holder, int position) {
        final FileItem item = items.get(position);

        holder.fileName.setText(item.getName());
        holder.fileInfo.setText(buildFileInfo(item));
        holder.fileIcon.setImageResource(getIconResource(item));

        if (item.isImage()) {
            String thumbnailUrl = baseUrl + "/files/thumbnail/" + item.getId();
            BitmapLoaderTask.load(thumbnailUrl, holder.fileIcon, bitmapCache,
                    THUMBNAIL_SIZE, THUMBNAIL_SIZE);
        }

        if (holder.starIcon != null) {
            holder.starIcon.setVisibility(item.isStarred() ? View.VISIBLE : View.GONE);
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

    public void addItems(List<FileItem> moreItems) {
        int startPosition = items.size();
        items.addAll(moreItems);
        notifyItemRangeInserted(startPosition, moreItems.size());
    }

    private String buildFileInfo(FileItem item) {
        if (item.isDirectory()) {
            int count = item.getDirectoryContentCount();
            String itemsLabel = translations != null
                    ? translations.t("DIRECTORY_CONTENT_COUNT", "items")
                    : "items";
            return count + " " + itemsLabel;
        }
        return formatFileSize(item.getSize());
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

    static class FileViewHolder extends RecyclerView.ViewHolder {

        final ImageView fileIcon;
        final TextView fileName;
        final TextView fileInfo;
        final ImageView starIcon;

        FileViewHolder(View itemView) {
            super(itemView);
            fileIcon = (ImageView) itemView.findViewById(R.id.file_icon);
            fileName = (TextView) itemView.findViewById(R.id.file_name);
            fileInfo = (TextView) itemView.findViewById(R.id.file_info);
            starIcon = (ImageView) itemView.findViewById(R.id.star_icon);
        }
    }
}
