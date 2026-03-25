package com.kuranas.mobile.presentation.search;

import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.TextView;

import androidx.recyclerview.widget.RecyclerView;

import com.kuranas.mobile.R;
import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.SearchResult;
import com.kuranas.mobile.domain.model.VideoItem;

import java.util.ArrayList;
import java.util.List;

public class SearchAdapter extends RecyclerView.Adapter<SearchAdapter.SearchViewHolder> {

    private final List<SearchResultItem> items;
    private final OnResultClickListener listener;

    public interface OnResultClickListener {
        void onResultClick(SearchResultItem item);
    }

    public SearchAdapter(OnResultClickListener listener) {
        this.items = new ArrayList<SearchResultItem>();
        this.listener = listener;
    }

    @Override
    public SearchViewHolder onCreateViewHolder(ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_search_result, parent, false);
        return new SearchViewHolder(view);
    }

    @Override
    public void onBindViewHolder(SearchViewHolder holder, int position) {
        final SearchResultItem item = items.get(position);

        holder.resultName.setText(item.getName());
        holder.resultPath.setText(item.getPath());
        holder.resultIcon.setImageResource(getIconForType(item.getType()));

        holder.itemView.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                if (listener != null) {
                    listener.onResultClick(item);
                }
            }
        });
    }

    @Override
    public int getItemCount() {
        return items.size();
    }

    public void setResults(SearchResult result) {
        items.clear();

        if (result == null) {
            notifyDataSetChanged();
            return;
        }

        List<FileItem> folders = result.getFolders();
        if (folders != null) {
            for (int i = 0; i < folders.size(); i++) {
                items.add(SearchResultItem.fromFolder(folders.get(i)));
            }
        }

        List<FileItem> files = result.getFiles();
        if (files != null) {
            for (int i = 0; i < files.size(); i++) {
                items.add(SearchResultItem.fromFile(files.get(i)));
            }
        }

        List<FileItem> images = result.getImages();
        if (images != null) {
            for (int i = 0; i < images.size(); i++) {
                items.add(SearchResultItem.fromImage(images.get(i)));
            }
        }

        List<VideoItem> videos = result.getVideos();
        if (videos != null) {
            for (int i = 0; i < videos.size(); i++) {
                items.add(SearchResultItem.fromVideo(videos.get(i)));
            }
        }

        notifyDataSetChanged();
    }

    private int getIconForType(SearchResultItem.Type type) {
        switch (type) {
            case FOLDER:
                return R.drawable.ic_folder;
            case IMAGE:
                return R.drawable.ic_image;
            case VIDEO:
                return R.drawable.ic_video;
            case FILE:
            default:
                return R.drawable.ic_file;
        }
    }

    static class SearchViewHolder extends RecyclerView.ViewHolder {

        final ImageView resultIcon;
        final TextView resultName;
        final TextView resultPath;

        SearchViewHolder(View itemView) {
            super(itemView);
            resultIcon = (ImageView) itemView.findViewById(R.id.result_icon);
            resultName = (TextView) itemView.findViewById(R.id.result_name);
            resultPath = (TextView) itemView.findViewById(R.id.result_path);
        }
    }
}
