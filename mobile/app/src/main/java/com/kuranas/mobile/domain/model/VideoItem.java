package com.kuranas.mobile.domain.model;

public final class VideoItem {

    private final int id;
    private final String name;
    private final String path;
    private final String parentPath;
    private final String format;
    private final long size;
    private final String createdAt;
    private final String updatedAt;

    public VideoItem(int id, String name, String path, String parentPath,
                     String format, long size, String createdAt, String updatedAt) {
        this.id = id;
        this.name = name;
        this.path = path;
        this.parentPath = parentPath;
        this.format = format;
        this.size = size;
        this.createdAt = createdAt;
        this.updatedAt = updatedAt;
    }

    public int getId() { return id; }
    public String getName() { return name; }
    public String getPath() { return path; }
    public String getParentPath() { return parentPath; }
    public String getFormat() { return format; }
    public long getSize() { return size; }
    public String getCreatedAt() { return createdAt; }
    public String getUpdatedAt() { return updatedAt; }

    public String getDisplayName() {
        if (name == null) {
            return "";
        }
        int dotIndex = name.lastIndexOf('.');
        if (dotIndex > 0) {
            return name.substring(0, dotIndex);
        }
        return name;
    }
}
