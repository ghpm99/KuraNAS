package com.kuranas.mobile.presentation.search;

import com.kuranas.mobile.domain.model.FileItem;
import com.kuranas.mobile.domain.model.VideoItem;

public class SearchResultItem {

    public enum Type {
        FILE,
        FOLDER,
        VIDEO,
        IMAGE
    }

    private final Type type;
    private final FileItem fileItem;
    private final VideoItem videoItem;

    private SearchResultItem(Type type, FileItem fileItem, VideoItem videoItem) {
        this.type = type;
        this.fileItem = fileItem;
        this.videoItem = videoItem;
    }

    public static SearchResultItem fromFile(FileItem item) {
        return new SearchResultItem(Type.FILE, item, null);
    }

    public static SearchResultItem fromFolder(FileItem item) {
        return new SearchResultItem(Type.FOLDER, item, null);
    }

    public static SearchResultItem fromVideo(VideoItem item) {
        return new SearchResultItem(Type.VIDEO, null, item);
    }

    public static SearchResultItem fromImage(FileItem item) {
        return new SearchResultItem(Type.IMAGE, item, null);
    }

    public Type getType() {
        return type;
    }

    public FileItem getFileItem() {
        return fileItem;
    }

    public VideoItem getVideoItem() {
        return videoItem;
    }

    public String getName() {
        if (videoItem != null) {
            return videoItem.getDisplayName();
        }
        if (fileItem != null) {
            return fileItem.getName();
        }
        return "";
    }

    public String getPath() {
        if (videoItem != null) {
            return videoItem.getPath();
        }
        if (fileItem != null) {
            return fileItem.getPath();
        }
        return "";
    }
}
