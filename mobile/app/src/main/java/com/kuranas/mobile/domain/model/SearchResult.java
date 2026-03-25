package com.kuranas.mobile.domain.model;

import java.util.List;

public final class SearchResult {

    private final String query;
    private final String suggestion;
    private final List<FileItem> files;
    private final List<FileItem> folders;
    private final List<VideoItem> videos;
    private final List<FileItem> images;

    public SearchResult(String query, String suggestion, List<FileItem> files,
                        List<FileItem> folders, List<VideoItem> videos,
                        List<FileItem> images) {
        this.query = query;
        this.suggestion = suggestion;
        this.files = files;
        this.folders = folders;
        this.videos = videos;
        this.images = images;
    }

    public String getQuery() { return query; }
    public String getSuggestion() { return suggestion; }
    public List<FileItem> getFiles() { return files; }
    public List<FileItem> getFolders() { return folders; }
    public List<VideoItem> getVideos() { return videos; }
    public List<FileItem> getImages() { return images; }

    public boolean isEmpty() {
        return (files == null || files.isEmpty())
                && (folders == null || folders.isEmpty())
                && (videos == null || videos.isEmpty())
                && (images == null || images.isEmpty());
    }

    public int totalCount() {
        int count = 0;
        if (files != null) count += files.size();
        if (folders != null) count += folders.size();
        if (videos != null) count += videos.size();
        if (images != null) count += images.size();
        return count;
    }
}
