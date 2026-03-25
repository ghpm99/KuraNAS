package com.kuranas.mobile.domain.model;

public final class Track {

    private final int id;
    private final int fileId;
    private final int position;
    private final String name;
    private final String title;
    private final String artist;
    private final String album;
    private final String genre;
    private final String year;
    private final int trackNumber;
    private final double durationSeconds;
    private final String format;
    private final String addedAt;

    public Track(int id, int fileId, int position, String name, String title,
                 String artist, String album, String genre, String year,
                 int trackNumber, double durationSeconds, String format, String addedAt) {
        this.id = id;
        this.fileId = fileId;
        this.position = position;
        this.name = name;
        this.title = title;
        this.artist = artist;
        this.album = album;
        this.genre = genre;
        this.year = year;
        this.trackNumber = trackNumber;
        this.durationSeconds = durationSeconds;
        this.format = format;
        this.addedAt = addedAt;
    }

    public int getId() { return id; }
    public int getFileId() { return fileId; }
    public int getPosition() { return position; }
    public String getName() { return name; }
    public String getTitle() { return title; }
    public String getArtist() { return artist; }
    public String getAlbum() { return album; }
    public String getGenre() { return genre; }
    public String getYear() { return year; }
    public int getTrackNumber() { return trackNumber; }
    public double getDurationSeconds() { return durationSeconds; }
    public String getFormat() { return format; }
    public String getAddedAt() { return addedAt; }

    public String getDisplayTitle() {
        if (title != null && !title.isEmpty()) {
            return title;
        }
        return name;
    }

    public String getDisplayArtist() {
        if (artist != null && !artist.isEmpty()) {
            return artist;
        }
        return "";
    }

    public String getFormattedDuration() {
        int totalSeconds = (int) durationSeconds;
        int minutes = totalSeconds / 60;
        int seconds = totalSeconds % 60;
        return String.format("%d:%02d", minutes, seconds);
    }
}
