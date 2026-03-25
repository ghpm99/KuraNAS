package com.kuranas.mobile.domain.model;

public final class FileItem {

    public static final int TYPE_DIRECTORY = 1;
    public static final int TYPE_FILE = 2;

    private final int id;
    private final String name;
    private final String path;
    private final String parentPath;
    private final int type;
    private final String format;
    private final long size;
    private final String updatedAt;
    private final String createdAt;
    private final boolean starred;
    private final int directoryContentCount;

    public FileItem(int id, String name, String path, String parentPath, int type,
                    String format, long size, String updatedAt, String createdAt,
                    boolean starred, int directoryContentCount) {
        this.id = id;
        this.name = name;
        this.path = path;
        this.parentPath = parentPath;
        this.type = type;
        this.format = format;
        this.size = size;
        this.updatedAt = updatedAt;
        this.createdAt = createdAt;
        this.starred = starred;
        this.directoryContentCount = directoryContentCount;
    }

    public int getId() { return id; }
    public String getName() { return name; }
    public String getPath() { return path; }
    public String getParentPath() { return parentPath; }
    public int getType() { return type; }
    public String getFormat() { return format; }
    public long getSize() { return size; }
    public String getUpdatedAt() { return updatedAt; }
    public String getCreatedAt() { return createdAt; }
    public boolean isStarred() { return starred; }
    public int getDirectoryContentCount() { return directoryContentCount; }

    public boolean isDirectory() { return type == TYPE_DIRECTORY; }
    public boolean isFile() { return type == TYPE_FILE; }

    public boolean isImage() {
        return format != null && (format.equals("jpg") || format.equals("jpeg")
                || format.equals("png") || format.equals("gif")
                || format.equals("webp") || format.equals("bmp"));
    }

    public boolean isAudio() {
        return format != null && (format.equals("mp3") || format.equals("flac")
                || format.equals("ogg") || format.equals("wav")
                || format.equals("m4a") || format.equals("aac"));
    }

    public boolean isVideo() {
        return format != null && (format.equals("mp4") || format.equals("mkv")
                || format.equals("avi") || format.equals("mov")
                || format.equals("wmv") || format.equals("webm"));
    }
}
